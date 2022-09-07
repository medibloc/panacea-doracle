package event

import (
	"crypto/sha256"
	"encoding/hex"

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
	EnclaveInfo() *sgx.EnclaveInfo
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

	uniqueID := hex.EncodeToString(e.reactor.EnclaveInfo().UniqueID)
	oracleRegistration, err := e.reactor.GRPCClient().GetOracleRegistration(addressValue, uniqueID)
	if err != nil {
		return err
	}

	nodePubKeyHash := sha256.Sum256(oracleRegistration.NodePubKey)

	err = sgx.VerifyRemoteReport(oracleRegistration.NodePubKeyRemoteReport, nodePubKeyHash[:], *e.reactor.EnclaveInfo())
	if err != nil {
		return err
	}

	// TODO: Executing Vote Txs
	return nil
}
