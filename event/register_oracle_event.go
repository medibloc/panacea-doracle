package event

import (
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

	oracleRegistration, err := svc.GrpcClient.GetOracleRegistration(addressValue, svc.UniqueID)
	if err != nil {
		return err
	}

	err = verifyRemoteReport(oracleRegistration)
	if err != nil {
		return err
	}

	// TODO: Executing Vote Txs
	return nil
}

func verifyRemoteReport(oracleRegistration *types.OracleRegistration) error {
	selfEnclaveInfo, err := sgx.GetSelfEnclaveInfo()
	if err != nil {
		return err
	}

	err = sgx.VerifyRemoteReport(oracleRegistration.NodePubKeyRemoteReport, oracleRegistration.NodePubKey, *selfEnclaveInfo)
	if err != nil {
		return err
	}
	return nil
}
