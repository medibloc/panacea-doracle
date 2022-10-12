package datadeal

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/event"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/medibloc/panacea-doracle/sgx"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var _ event.Event = (*DataVerificationEvent)(nil)

type DataVerificationEvent struct {
	reactor reactor
}

type reactor interface {
	GRPCClient() *panacea.GrpcClient
	EnclaveInfo() *sgx.EnclaveInfo
	OracleAcc() *panacea.OracleAccount
	OraclePrivKey() *btcec.PrivateKey
	Config() *config.Config
	QueryClient() *panacea.QueryClient
}

func NewDataVerificationEvent(r reactor) DataVerificationEvent {
	return DataVerificationEvent{r}
}

func (d DataVerificationEvent) GetEventType() string {
	return "message"
}

func (d DataVerificationEvent) GetEventAttributeKey() string {
	return "action"
}

func (d DataVerificationEvent) GetEventAttributeValue() string {
	return "SellData"
}

func (d DataVerificationEvent) EventHandler(event ctypes.ResultEvent) error {
	verifiableCID := event.Events[types.EventTypeDataVerificationVote+"."+types.AttributeKeyVerifiableCID][0]

	out := fmt.Sprintf("%s.txt", verifiableCID)
	err := sh.Get(verifiableCID, out)
	if err != nil {
		return fmt.Errorf("error: %s", err)
	}

	return nil
}
