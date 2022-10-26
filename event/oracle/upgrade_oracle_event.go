package oracle

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/edgelesssys/ego/enclave"
	"github.com/medibloc/panacea-doracle/sgx"
	"github.com/tendermint/tendermint/light/provider"

	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-doracle/event"
	"github.com/medibloc/panacea-doracle/panacea"
	log "github.com/sirupsen/logrus"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

type UpgradeOracleEvent struct {
	reactor event.Reactor
}

var _ event.Event = (*UpgradeOracleEvent)(nil)

func NewUpgradeOracleEvent(s event.Reactor) UpgradeOracleEvent {
	return UpgradeOracleEvent{s}
}

func (e UpgradeOracleEvent) GetEventType() string {
	return "message"
}

func (e UpgradeOracleEvent) GetEventAttributeKey() string {
	return "action"
}

func (e UpgradeOracleEvent) GetEventAttributeValue() string {
	return "'UpgradeOracle'"
}

func (e UpgradeOracleEvent) EventHandler(event ctypes.ResultEvent) error {
	uniqueID := event.Events[oracletypes.EventTypeUpgradeVote+"."+oracletypes.AttributeKeyUniqueID][0]
	addressValue := event.Events[oracletypes.EventTypeUpgradeVote+"."+oracletypes.AttributeKeyOracleAddress][0]
	queryClient := e.reactor.QueryClient()

	oracleRegistration, err := queryClient.GetOracleRegistration(addressValue, uniqueID)
	if err != nil {
		log.Infof("failed to get oracleRegistration, voting ignored. uniqueID(%s), address(%s)", uniqueID, addressValue)
		return err
	}

	voteOption, err := e.verifyAndGetVoteOption(oracleRegistration)
	if err != nil {
		return err
	}

	msgVoteOracleRegistration, err := makeOracleRegistrationVote(
		uniqueID,
		e.reactor.OracleAcc().GetAddress(),
		addressValue,
		voteOption,
		e.reactor.OraclePrivKey().Serialize(),
		oracleRegistration.NodePubKey,
		oracleRegistration.Nonce,
	)
	if err != nil {
		return err
	}

	txBuilder := panacea.NewTxBuilder(*e.reactor.QueryClient())

	txBytes, err := generateTxBytes(msgVoteOracleRegistration, e.reactor.OracleAcc().GetPrivKey(), e.reactor.Config(), txBuilder)
	if err != nil {
		return err
	}

	if err := broadcastTx(e.reactor.GRPCClient(), txBytes); err != nil {
		return err
	}

	return nil
}

func (e UpgradeOracleEvent) verifyAndGetVoteOption(r *oracletypes.OracleRegistration) (oracletypes.VoteOption, error) {
	upgradeInfo, err := e.reactor.QueryClient().GetOracleUpgradeInfo()
	if err != nil {
		if errors.Is(err, panacea.ErrEmptyValue) {
			log.Infof("not exist oracle upgrade info")
			return oracletypes.VOTE_OPTION_NO, nil
		}
		log.Errorf("failed to get oracle upgrade info. %v", err)
		return oracletypes.VOTE_OPTION_UNSPECIFIED, err
	}
	if upgradeInfo.UniqueId != r.UniqueId {
		log.Infof("oracle's uniqueID does not match the uniqueID being upgraded. expected uniqueID(%s), oracle's uniqueID(%s), ",
			upgradeInfo.UniqueId,
			r.UniqueId)
		return oracletypes.VOTE_OPTION_NO, nil
	}

	block, err := e.reactor.QueryClient().GetLightBlock(r.TrustedBlockHeight)
	if err != nil {
		switch err {
		case provider.ErrLightBlockNotFound, provider.ErrHeightTooHigh:
			return oracletypes.VOTE_OPTION_NO, nil
		default:
			return oracletypes.VOTE_OPTION_UNSPECIFIED, err
		}
	}

	if !bytes.Equal(block.Hash().Bytes(), r.TrustedBlockHash) {
		log.Warnf("failed to verify trusted block information. height(%v), expected block hash(%s), got block hash(%s)",
			r.TrustedBlockHeight,
			hex.EncodeToString(block.Hash().Bytes()),
			hex.EncodeToString(r.TrustedBlockHash),
		)
		return oracletypes.VOTE_OPTION_NO, nil
	}

	if err := e.VerifyRemoteReport(r.NodePubKeyRemoteReport, r.NodePubKey, r.UniqueId); err != nil {
		log.Warnf("failed to verification report. uniqueID(%s), address(%s), err(%v)", r.UniqueId, r.Address, err)
		return oracletypes.VOTE_OPTION_NO, nil
	} else {
		return oracletypes.VOTE_OPTION_YES, nil
	}
}

func (e UpgradeOracleEvent) VerifyRemoteReport(reportBytes, expectedData []byte, expectedUniqueID string) error {
	enclaveInfo := e.reactor.EnclaveInfo()
	report, err := enclave.VerifyRemoteReport(reportBytes)
	if err != nil {
		return err
	}

	if report.SecurityVersion < sgx.PromisedMinSecurityVersion {
		return fmt.Errorf("invalid security version in the report")
	}
	if !bytes.Equal(report.ProductID, enclaveInfo.ProductID) {
		return fmt.Errorf("invalid product ID in the report")
	}
	if !bytes.Equal(report.SignerID, enclaveInfo.SignerID) {
		return fmt.Errorf("invalid signer ID in the report")
	}
	uniqueID, err := hex.DecodeString(expectedUniqueID)
	if err != nil {
		return fmt.Errorf("invalid uniqueID format. uniqueID(%s). %w", uniqueID, err)
	}
	if !bytes.Equal(report.UniqueID, uniqueID) {
		return fmt.Errorf("invalid unique ID in the report")
	}
	if !bytes.Equal(report.Data[:len(expectedData)], expectedData) {
		return fmt.Errorf("invalid data in the report")
	}

	return nil
}
