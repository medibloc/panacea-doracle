package oracle

import (
	"fmt"

	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-doracle/event"
	log "github.com/sirupsen/logrus"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var _ event.Event = (*UpgradedOracleEvent)(nil)

type UpgradedOracleEvent struct {
	reactor    event.Reactor
	enable     bool
	voteEvents []event.Event
}

func (e *UpgradedOracleEvent) GetEventName() string {
	return "UpgradedOracleEvent"
}

func NewUpgradedOracleEvent(r event.Reactor, voteEvents []event.Event) *UpgradedOracleEvent {
	return &UpgradedOracleEvent{
		reactor:    r,
		enable:     true,
		voteEvents: voteEvents,
	}
}

func (e *UpgradedOracleEvent) Prepare() error {
	return e.setEnableVoteEvents()
}

func (e *UpgradedOracleEvent) GetEventQuery() string {
	return fmt.Sprintf("tm.event='NewBlock' AND %s.%s='%s'",
		oracletypes.EventTypeUpgradeVote,
		oracletypes.AttributeKeyVoteStatus,
		oracletypes.AttributeValueUpgradeStatusEnded,
	)
}

func (e *UpgradedOracleEvent) SetEnable(enable bool) {
	e.enable = enable
}

func (e *UpgradedOracleEvent) Enabled() bool {
	return e.enable
}

func (e *UpgradedOracleEvent) EventHandler(event ctypes.ResultEvent) error {
	return e.setEnableVoteEvents()
}

func (e *UpgradedOracleEvent) setEnableVoteEvents() error {
	log.Info("checking is activeUniqueID")
	uniqueID, err := e.reactor.QueryClient().GetOracleParamsUniqueID()
	if err != nil {
		return err
	}

	enable := e.reactor.EnclaveInfo().UniqueIDHex() == uniqueID
	log.Infof("activeUniqueID(%s), my uniqueID(%s) enable(%v)", uniqueID, e.reactor.EnclaveInfo().UniqueIDHex(), enable)

	for _, e := range e.voteEvents {
		e.SetEnable(enable)
	}
	return nil
}
