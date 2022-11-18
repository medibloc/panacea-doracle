package oracle

import (
	"github.com/medibloc/panacea-doracle/event"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

type UpgradeOracleEvent struct {
	reactor event.Reactor
}

var _ event.Event = (*UpgradeOracleEvent)(nil)

func NewUpgradeOracleEvent(s event.Reactor) UpgradeOracleEvent {
	return UpgradeOracleEvent{s}
}

func (e UpgradeOracleEvent) GetEventQuery() string {
	return "message.action = 'OracleUpgrade'"
}

func (e UpgradeOracleEvent) EventHandler(_ ctypes.ResultEvent) error {
	//uniqueID := event.Events[oracletypes.EventTypeUpgradeVote+"."+oracletypes.AttributeKeyUniqueID][0]
	//votingTargetAddress := event.Events[oracletypes.EventTypeUpgradeVote+"."+oracletypes.AttributeKeyOracleAddress][0]
	//
	//msgVoteOracleRegistration, err := e.verifyAndGetMsgVoteOracleRegistration(uniqueID, votingTargetAddress)
	//if err != nil {
	//	return err
	//}
	//
	//log.Infof("oracle upgrade voting info. uniqueID(%s), voterAddress(%s), votingTargetAddress(%s), voteOption(%s)",
	//	msgVoteOracleRegistration.OracleRegistrationVote.UniqueId,
	//	msgVoteOracleRegistration.OracleRegistrationVote.VoterAddress,
	//	msgVoteOracleRegistration.OracleRegistrationVote.VotingTargetAddress,
	//	msgVoteOracleRegistration.OracleRegistrationVote.VoteOption,
	//)
	//
	//txBuilder := panacea.NewTxBuilder(*e.reactor.QueryClient())
	//txBytes, err := txBuilder.GenerateTxBytes(e.reactor.OracleAcc().GetPrivKey(), e.reactor.Config(), msgVoteOracleRegistration)
	//if err != nil {
	//	return err
	//}
	//
	//txHeight, txHash, err := e.reactor.BroadcastTx(txBytes)
	//if err != nil {
	//	return fmt.Errorf("failed to oracleRegistrationVote transaction for oracle upgrade: %v", err)
	//} else {
	//	log.Infof("succeeded to oracleRegistrationVote transaction for oracle upgrade. height(%v), hash(%s)", txHeight, txHash)
	//}

	return nil
}

//func (e UpgradeOracleEvent) verifyAndGetMsgVoteOracleRegistration(uniqueID, votingTargetAddress string) (*oracletypes.MsgVoteOracleRegistration, error) {
//	queryClient := e.reactor.QueryClient()
//	voterAddress := e.reactor.OracleAcc().GetAddress()
//	oraclePrivKeyBz := e.reactor.OraclePrivKey().Serialize()
//	voterUniqueID := e.reactor.EnclaveInfo().UniqueIDHex()
//
//	oracleRegistration, err := queryClient.GetOracleRegistration(votingTargetAddress, uniqueID)
//	if err != nil {
//		log.Infof("failed to get oracleRegistration. uniqueID(%s), address(%s). %v", uniqueID, votingTargetAddress, err)
//		return makeMsgVoteOracleRegistrationVoteTypeNo(uniqueID, voterUniqueID, voterAddress, votingTargetAddress, oraclePrivKeyBz)
//	}
//
//	voteOption, err := e.verifyAndGetVoteOption(oracleRegistration)
//	if err != nil {
//		log.Infof("vote No due to error while verify: %v", err)
//	}
//
//	return makeMsgVoteOracleRegistration(
//		uniqueID,
//		voterUniqueID,
//		voterAddress,
//		votingTargetAddress,
//		voteOption,
//		oraclePrivKeyBz,
//		oracleRegistration.NodePubKey,
//		oracleRegistration.Nonce,
//	)
//
//}
//
//func (e UpgradeOracleEvent) verifyAndGetVoteOption(oracleRegistration *oracletypes.OracleRegistration) (oracletypes.VoteOption, error) {
//	queryClient := e.reactor.QueryClient()
//	upgradeInfo, err := queryClient.GetOracleUpgradeInfo()
//	if err != nil {
//		return oracletypes.VOTE_OPTION_NO, fmt.Errorf("failed to get oracle upgrade info. %v", err)
//	}
//	if upgradeInfo.UniqueId != oracleRegistration.UniqueId {
//		return oracletypes.VOTE_OPTION_NO, fmt.Errorf("oracle's uniqueID does not match the uniqueID being upgraded. expected uniqueID(%s), oracle's uniqueID(%s), ",
//			upgradeInfo.UniqueId,
//			oracleRegistration.UniqueId,
//		)
//	}
//
//	if err := verifyTrustedBlockInfo(queryClient, oracleRegistration.TrustedBlockHeight, oracleRegistration.TrustedBlockHash); err != nil {
//		return oracletypes.VOTE_OPTION_NO, err
//	}
//
//	if err := e.verifyRemoteReport(oracleRegistration); err != nil {
//		return oracletypes.VOTE_OPTION_NO, fmt.Errorf("failed to verification report. uniqueID(%s), address(%s), err(%v)", oracleRegistration.UniqueId, oracleRegistration.Address, err)
//	}
//
//	return oracletypes.VOTE_OPTION_YES, nil
//
//}
//
//func (e UpgradeOracleEvent) verifyRemoteReport(oracleRegistration *oracletypes.OracleRegistration) error {
//	reportBz := oracleRegistration.NodePubKeyRemoteReport
//	expectedNodePubKeyHash := crypto.KDFSHA256(oracleRegistration.NodePubKey)
//	expectedUniqueID := oracleRegistration.UniqueId
//
//	enclaveInfo := e.reactor.EnclaveInfo()
//	report, err := enclave.VerifyRemoteReport(reportBz)
//	if err != nil {
//		return err
//	}
//
//	if report.SecurityVersion < sgx.PromisedMinSecurityVersion {
//		return fmt.Errorf("invalid security version in the report")
//	}
//	if !bytes.Equal(report.ProductID, enclaveInfo.ProductID) {
//		return fmt.Errorf("invalid product ID in the report")
//	}
//	if !bytes.Equal(report.SignerID, enclaveInfo.SignerID) {
//		return fmt.Errorf("invalid signer ID in the report")
//	}
//	uniqueID, err := hex.DecodeString(expectedUniqueID)
//	if err != nil {
//		return fmt.Errorf("invalid uniqueID format. uniqueID(%s). %w", uniqueID, err)
//	}
//	if !bytes.Equal(report.UniqueID, uniqueID) {
//		return fmt.Errorf("invalid unique ID in the report")
//	}
//	if !bytes.Equal(report.Data[:len(expectedNodePubKeyHash)], expectedNodePubKeyHash) {
//		return fmt.Errorf("invalid data in the report. expected(%s), got(%s)",
//			hex.EncodeToString(expectedNodePubKeyHash),
//			hex.EncodeToString(report.Data[:len(expectedNodePubKeyHash)]),
//		)
//	}
//
//	return nil
//}
