package event

import (
	"fmt"
	"github.com/edgelesssys/ego/enclave"
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
	addressValue := event.Events[types.EventTypeRegistrationVote+"."+types.AttributeKeyOracleAddress][0]

	oracleRegistration, err := svc.GrpcClient.GetOracleRegistration(addressValue, svc.UniqueID)
	if err != nil {
		return err
	}

	err = verifyRemoteReport(oracleRegistration.NodePubKeyRemoteReport)
	if err != nil {
		return err
	}

	// TODO: Executing Vote Txs

	return nil
}

func verifyRemoteReport(nodePubKeyRemoteReport []byte) error {
	report, err := enclave.VerifyRemoteReport(nodePubKeyRemoteReport)
	if err != nil {
		return err
	}

	fmt.Println("securityVersion: ", report.SecurityVersion)
	fmt.Println("productID: ", report.ProductID)
	fmt.Println("productID: ", report.UniqueID)

	return nil
}
