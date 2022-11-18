package oracle

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/medibloc/panacea-doracle/panacea"

	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-doracle/event"
	"github.com/medibloc/panacea-doracle/sgx"
	log "github.com/sirupsen/logrus"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var _ event.Event = (*RegisterOracleEvent)(nil)

type RegisterOracleEvent struct {
	reactor event.Reactor
}

func NewRegisterOracleEvent(s event.Reactor) RegisterOracleEvent {
	return RegisterOracleEvent{s}
}

func (e RegisterOracleEvent) GetEventQuery() string {
	return "message.action = 'RegisterOracle'"
}

func (e RegisterOracleEvent) EventHandler(event ctypes.ResultEvent) error {
	uniqueID := event.Events[oracletypes.EventTypeRegistrationVote+"."+oracletypes.AttributeKeyUniqueID][0]
	votingTargetAddress := event.Events[oracletypes.EventTypeRegistrationVote+"."+oracletypes.AttributeKeyOracleAddress][0]
	oraclePubKey := event.Events[oracletypes.EventTypeRegistrationVote+"."+oracletypes.AttributeKeyOraclePubKey][0]

	queryClient := e.reactor.QueryClient()
	voterAddress := e.reactor.OracleAcc().GetAddress()
	oraclePrivKey := e.reactor.OraclePrivKey().Serialize()
	voterUniqueID := e.reactor.EnclaveInfo().UniqueIDHex()

	if uniqueID != voterUniqueID {
		return fmt.Errorf("uniqueID mismatch")
	}

	oracleRegistration, err := queryClient.GetOracleRegistration(votingTargetAddress, uniqueID, oraclePubKey)
	if err != nil {
		return fmt.Errorf("error occurs when getting oracle registration")
	}

	if err := verifyTrustedBlockInfo(e.reactor.QueryClient(), oracleRegistration.TrustedBlockHeight, oracleRegistration.TrustedBlockHash); err != nil {
		return fmt.Errorf("failed to verify trusted block info: %w", err)
	}

	oraclePubKeyBz, err := hex.DecodeString(oracleRegistration.OraclePubKey)
	if err != nil {
		return fmt.Errorf("failed to decode oracle public key: %w", err)
	}

	oraclePubKeyHash := sha256.Sum256(oraclePubKeyBz)

	if err := sgx.VerifyRemoteReport(oracleRegistration.OraclePubKeyRemoteReport, oraclePubKeyHash[:], *e.reactor.EnclaveInfo()); err != nil {
		log.Warnf("failed to verification report. uniqueID(%s), address(%s), err(%v)", oracleRegistration.UniqueId, oracleRegistration.Address, err)
		return fmt.Errorf("failed to verify remote report: %w", err)
	}

	registrationVote := &oracletypes.OracleRegistrationVote{
		UniqueId:            uniqueID,
		VoterUniqueId:       voterUniqueID,
		VoterAddress:        voterAddress,
		VotingTargetAddress: oracleRegistration.Address,
		VotingTargetPubKey:  oracleRegistration.OraclePubKey,
	}

	key := secp256k1.PrivKey{
		Key: oraclePrivKey,
	}

	marshaledRegistrationVote, err := registrationVote.Marshal()
	if err != nil {
		return err
	}

	sig, err := key.Sign(marshaledRegistrationVote)
	if err != nil {
		return err
	}

	msgVoteOracleRegistration := &oracletypes.MsgVoteOracleRegistration{
		OracleRegistrationVote: registrationVote,
		Signature:              sig,
	}

	log.Infof("new oracle registeration voting info. uniqueID(%s), voterAddress(%s), votingTargetAddress(%s), votingTargetPubKey(%s)",
		msgVoteOracleRegistration.OracleRegistrationVote.UniqueId,
		msgVoteOracleRegistration.OracleRegistrationVote.VoterAddress,
		msgVoteOracleRegistration.OracleRegistrationVote.VotingTargetAddress,
		msgVoteOracleRegistration.OracleRegistrationVote.VotingTargetPubKey,
	)

	txBuilder := panacea.NewTxBuilder(*e.reactor.QueryClient())
	txBytes, err := txBuilder.GenerateTxBytes(e.reactor.OracleAcc().GetPrivKey(), e.reactor.Config(), msgVoteOracleRegistration)
	if err != nil {
		return err
	}

	txHeight, txHash, err := e.reactor.BroadcastTx(txBytes)
	if err != nil {
		return fmt.Errorf("failed to oracleRegistrationVote transaction for new oracle registration: %v", err)
	} else {
		log.Infof("succeeded to oracleRegistrationVote transaction for new oracle registration. height(%v), hash(%s)", txHeight, txHash)
	}

	return nil
}
