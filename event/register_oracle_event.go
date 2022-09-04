package event

import (
	"fmt"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-doracle/service"
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
	// TODO: Verifying Remote Attestation and Executing Vote Txs
	addressValue := event.Events[types.EventTypeRegistrationVote+"."+types.AttributeKeyOracleAddress][0]
	fmt.Println(addressValue)

	oracleRegistration, err := svc.GrpcClient.GetOracleRegistration(svc.UniqueID, addressValue)
	if err != nil {
		return err
	}

	fmt.Println(oracleRegistration)

	fmt.Println("RegisterOracle Event Handler")
	return nil
}
