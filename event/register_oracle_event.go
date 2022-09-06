package event

import (
	"encoding/hex"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-doracle/service"
	"github.com/medibloc/panacea-doracle/sgx"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var _ Event = (*RegisterOracleEvent)(nil)

type RegisterOracleEvent struct{}

func (e RegisterOracleEvent) GetEventType() string {
	return "message"
}

func (e RegisterOracleEvent) GetEventAttributeKey() string {
	return "action"
}

func (e RegisterOracleEvent) GetEventAttributeValue() string {
	return "'RegisterOracle'"
}

func (e RegisterOracleEvent) EventHandler(event ctypes.ResultEvent, svc *service.Service) error {
	addressValue := event.Events[types.EventTypeRegistrationVote+"."+types.AttributeKeyOracleAddress][0]

	oracleRegistration, err := svc.GrpcClient.GetOracleRegistration(addressValue, hex.EncodeToString(svc.EnclaveInfo.UniqueID))
	if err != nil {
		return err
	}

	err = sgx.VerifyRemoteReport(oracleRegistration.NodePubKeyRemoteReport, oracleRegistration.NodePubKey, *svc.EnclaveInfo)
	if err != nil {
		return err
	}

	// TODO: Executing Vote Txs
	
	return nil
}

func yesVote(selfEnclave *sgx.EnclaveInfo, svc *service.Service, address string) types.OracleRegistrationVote {
	registrationVoteYes := types.OracleRegistrationVote{
		UniqueId:               hex.EncodeToString(selfEnclave.UniqueID),
		VoterAddress:           svc.OracleAccount.GetAddress(),
		VotingTargetAddress:    address,
		VoteOption:             types.VOTE_OPTION_YES,
		EncryptedOraclePrivKey: svc.OraclePrivKey,
	}

	return registrationVoteYes
}

func noVote(selfEnclave *sgx.EnclaveInfo, svc *service.Service, address string) types.OracleRegistrationVote {
	registrationVoteNo := types.OracleRegistrationVote{
		UniqueId:               hex.EncodeToString(selfEnclave.UniqueID),
		VoterAddress:           svc.OracleAccount.GetAddress(),
		VotingTargetAddress:    address,
		VoteOption:             types.VOTE_OPTION_NO,
		EncryptedOraclePrivKey: svc.OraclePrivKey,
	}

	return registrationVoteNo
}
