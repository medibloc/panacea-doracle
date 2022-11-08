package oracle

import (
	"crypto/sha256"
	"fmt"

	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-doracle/event"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/medibloc/panacea-doracle/sgx"
	log "github.com/sirupsen/logrus"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var _ event.Event = (*RegisterOracleEvent)(nil)

type RegisterOracleEvent struct {
	reactor event.Reactor
	enable  bool
}

func NewRegisterOracleEvent(r event.Reactor) *RegisterOracleEvent {
	return &RegisterOracleEvent{
		reactor: r,
	}
}

func (e *RegisterOracleEvent) Prepare() error {
	return nil
}

func (e *RegisterOracleEvent) GetEventName() string {
	return "RegisterOracleEvent"
}

func (e *RegisterOracleEvent) GetEventType() string {
	return "message"
}

func (e *RegisterOracleEvent) GetEventAttributeKey() string {
	return "action"
}

func (e *RegisterOracleEvent) GetEventAttributeValue() string {
	return "'RegisterOracle'"
}

func (e *RegisterOracleEvent) SetEnable(enable bool) {
	e.enable = enable
}

func (e *RegisterOracleEvent) Enabled() bool {
	return e.enable
}

func (e *RegisterOracleEvent) EventHandler(event ctypes.ResultEvent) error {
	uniqueID := event.Events[oracletypes.EventTypeRegistrationVote+"."+oracletypes.AttributeKeyUniqueID][0]
	votingTargetAddress := event.Events[oracletypes.EventTypeRegistrationVote+"."+oracletypes.AttributeKeyOracleAddress][0]

	msgVoteOracleRegistration, err := e.verifyAndGetMsgVoteOracleRegistration(uniqueID, votingTargetAddress)
	if err != nil {
		return err
	}

	log.Infof("new oracle registeration voting info. uniqueID(%s), voterAddress(%s), votingTargetAddress(%s), voteOption(%s)",
		msgVoteOracleRegistration.OracleRegistrationVote.UniqueId,
		msgVoteOracleRegistration.OracleRegistrationVote.VoterAddress,
		msgVoteOracleRegistration.OracleRegistrationVote.VotingTargetAddress,
		msgVoteOracleRegistration.OracleRegistrationVote.VoteOption,
	)

	txBuilder := panacea.NewTxBuilder(*e.reactor.QueryClient())
	txBytes, err := txBuilder.GenerateTxBytes(e.reactor.OracleAcc().GetPrivKey(), e.reactor.Config(), msgVoteOracleRegistration)
	if err != nil {
		return err
	}

	if err := e.broadcastTx(txBytes); err != nil {
		return err
	}

	return nil
}

func (e *RegisterOracleEvent) verifyAndGetMsgVoteOracleRegistration(uniqueID, votingTargetAddress string) (*oracletypes.MsgVoteOracleRegistration, error) {
	log.Infof("verifing oracleRegistration. uniqueID(%s), votingTargetAddress(%s)", uniqueID, votingTargetAddress)

	queryClient := e.reactor.QueryClient()
	voterAddress := e.reactor.OracleAcc().GetAddress()
	oraclePrivKeyBz := e.reactor.OraclePrivKey().Serialize()
	voterUniqueID := e.reactor.EnclaveInfo().UniqueIDHex()

	if uniqueID != voterUniqueID {
		log.Infof("vote No due to because oracle's uniqueID does not match the requested uniqueID. expected(%s) got(%s)",
			voterUniqueID,
			uniqueID,
		)
		return makeMsgVoteOracleRegistrationVoteTypeNo(
			voterUniqueID,
			voterUniqueID,
			voterAddress,
			votingTargetAddress,
			oraclePrivKeyBz,
		)
	} else {
		oracleRegistration, err := queryClient.GetOracleRegistration(votingTargetAddress, uniqueID)
		if err != nil {
			return makeMsgVoteOracleRegistrationVoteTypeNo(
				voterUniqueID,
				voterUniqueID,
				voterAddress,
				votingTargetAddress,
				oraclePrivKeyBz,
			)
		}

		voteOption, err := e.verifyAndGetVoteOption(oracleRegistration)
		if err != nil {
			log.Infof("vote No due to error while verify: %v", err)
		} else {
			log.Infof("verification success. uniqueID(%s), votingTargetAddress(%s)", uniqueID, votingTargetAddress)
		}

		return makeMsgVoteOracleRegistration(
			voterUniqueID,
			voterUniqueID,
			voterAddress,
			votingTargetAddress,
			voteOption,
			oraclePrivKeyBz,
			oracleRegistration.NodePubKey,
			oracleRegistration.Nonce,
		)
	}
}

// verifyAndGetVoteOption performs a verification to determine a vote.
// - Verify that trustedBlockInfo registered in OracleRegistration is valid
// - Verify that the RemoteReport is valid
func (e *RegisterOracleEvent) verifyAndGetVoteOption(oracleRegistration *oracletypes.OracleRegistration) (oracletypes.VoteOption, error) {
	if err := verifyTrustedBlockInfo(e.reactor.QueryClient(), oracleRegistration.TrustedBlockHeight, oracleRegistration.TrustedBlockHash); err != nil {
		return oracletypes.VOTE_OPTION_NO, err
	}

	nodePubKeyHash := sha256.Sum256(oracleRegistration.NodePubKey)

	if err := sgx.VerifyRemoteReport(oracleRegistration.NodePubKeyRemoteReport, nodePubKeyHash[:], *e.reactor.EnclaveInfo()); err != nil {
		log.Warnf("failed to verification report. uniqueID(%s), address(%s), err(%v)", oracleRegistration.UniqueId, oracleRegistration.Address, err)
		return oracletypes.VOTE_OPTION_NO, nil
	} else {
		return oracletypes.VOTE_OPTION_YES, nil
	}
}

// broadcastTx broadcast transaction to blockchain.
func (e *RegisterOracleEvent) broadcastTx(txBz []byte) error {
	resp, err := e.reactor.GRPCClient().BroadcastTx(txBz)
	if err != nil {
		return err
	}
	if resp.TxResponse.Code != 0 {
		return fmt.Errorf("failed to oracleRegistrationVote transaction for new oracle registration: %v", resp.TxResponse.RawLog)
	}

	log.Infof("succeeded to oracleRegistrationVote transaction for new oracle registration. height(%v), hash(%s)", resp.TxResponse.Height, resp.TxResponse.TxHash)

	return nil
}
