package oracle

import (
	"errors"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-doracle/event"
	"github.com/medibloc/panacea-doracle/panacea"
	log "github.com/sirupsen/logrus"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

type UpgradeOracleEvent struct {
	service reactor
}

var _ event.Event = (*UpgradeOracleEvent)(nil)

func NewUpgradeOracleEvent(s reactor) UpgradeOracleEvent {
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
	addressValue := event.Events[oracletypes.EventTypeUpgradeVote+"."+oracletypes.AttributeKeyOracleAddress][0]
	service := e.service
	queryClient := service.QueryClient()

	uniqueID := service.EnclaveInfo().UniqueIDHex()
	oracleRegistration, err := queryClient.GetOracleRegistration(addressValue, uniqueID)
	if err != nil {
		return err
	}

	voteOption, err := e.verifyAndGetVoteOption(oracleRegistration)
	if err != nil {
		return err
	}

	msgVoteOracleRegistration, err := makeOracleRegistrationVote(
		uniqueID,
		e.service.OracleAcc().GetAddress(),
		addressValue,
		voteOption,
		e.service.OraclePrivKey().Serialize(),
		oracleRegistration.NodePubKey,
		oracleRegistration.Nonce,
	)
	if err != nil {
		return err
	}

	txBuilder := panacea.NewTxBuilder(*e.service.QueryClient())

	txBytes, err := generateTxBytes(msgVoteOracleRegistration, e.service.OracleAcc().GetPrivKey(), e.service.Config(), txBuilder)
	if err != nil {
		return err
	}

	if err := broadcastTx(e.service.GRPCClient(), txBytes); err != nil {
		return err
	}

	return nil
}

func (e UpgradeOracleEvent) verifyAndGetVoteOption(r *oracletypes.OracleRegistration) (oracletypes.VoteOption, error) {
	upgradeInfo, err := e.service.QueryClient().GetOracleUpgradeInfo()
	if err != nil {
		if errors.Is(err, panacea.ErrEmptyValue) {
			log.Infof("not exist oracle upgrade info")
			return oracletypes.VOTE_OPTION_NO, nil
		}
		log.Errorf("failed to get oracle upgrade info. %v", err)
		return oracletypes.VOTE_OPTION_NO, nil
	}
	if upgradeInfo.UniqueId != e.service.EnclaveInfo().UniqueIDHex() {
		log.Infof("oracle's uniqueID does not match the uniqueID being upgraded. expected uniqueID(%s), oracle's uniqueID(%s), ",
			upgradeInfo.UniqueId,
			e.service.EnclaveInfo().UniqueIDHex())
		return oracletypes.VOTE_OPTION_NO, nil
	}
	return verifyAndGetVoteOption(e.service, r)
}
