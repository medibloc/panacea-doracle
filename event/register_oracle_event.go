package event

import (
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/medibloc/panacea-doracle/sgx"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var _ Event = (*RegisterOracleEvent)(nil)

type RegisterOracleEvent struct {
	reactor reactor
}

// reactor contains all ingredients needed for handling this type of event
type reactor interface {
	GRPCClient() *panacea.GrpcClient
	UniqueID() string
}

func NewRegisterOracleEvent(r reactor) RegisterOracleEvent {
	return RegisterOracleEvent{r}
}

func (e RegisterOracleEvent) GetEventType() string {
	return "message"
}

func (e RegisterOracleEvent) GetEventAttributeKey() string {
	return "action"
}

func (e RegisterOracleEvent) GetEventAttributeValue() string {
	return "'RegisterOracle'"
}

func (e RegisterOracleEvent) EventHandler(event ctypes.ResultEvent) error {
	addressValue := event.Events[types.EventTypeRegistrationVote+"."+types.AttributeKeyOracleAddress][0]

	oracleRegistration, err := e.reactor.GRPCClient().GetOracleRegistration(addressValue, e.reactor.UniqueID())
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
