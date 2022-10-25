package oracle

import (
	"errors"

	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-doracle/event"
	"github.com/medibloc/panacea-doracle/panacea"
	log "github.com/sirupsen/logrus"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

type UpgradeOracleEvent struct {
	reactor event.Reactor
}

var _ event.Event = (*UpgradeOracleEvent)(nil)

func NewUpgradeOracleEvent(s event.Reactor) UpgradeOracleEvent {
	return UpgradeOracleEvent{s}
}

func (e UpgradeOracleEvent) GetEventType() string {
	return "message"
}

func (e UpgradeOracleEvent) GetEventAttributeKey() string {
	return "action"
}

func (e UpgradeOracleEvent) GetEventAttributeValue() string {
	return "'UpgradeOracle'"
}

func (e UpgradeOracleEvent) EventHandler(event ctypes.ResultEvent) error {
	uniqueID := event.Events[oracletypes.EventTypeUpgradeVote+"."+oracletypes.AttributeKeyUniqueID][0]
	addressValue := event.Events[oracletypes.EventTypeUpgradeVote+"."+oracletypes.AttributeKeyOracleAddress][0]
	queryClient := e.reactor.QueryClient()

	oracleRegistration, err := queryClient.GetOracleRegistration(addressValue, uniqueID)
	if err != nil {
		log.Infof("failed to get oracleRegistration, voting ignored. uniqueID(%s), address(%s)", uniqueID, addressValue)
		return err
	}

	voteOption, err := e.verifyAndGetVoteOption(oracleRegistration)
	if err != nil {
		return err
	}

	msgVoteOracleRegistration, err := makeOracleRegistrationVote(
		uniqueID,
		e.reactor.OracleAcc().GetAddress(),
		addressValue,
		voteOption,
		e.reactor.OraclePrivKey().Serialize(),
		oracleRegistration.NodePubKey,
		oracleRegistration.Nonce,
	)
	if err != nil {
		return err
	}

	txBuilder := panacea.NewTxBuilder(*e.reactor.QueryClient())

	txBytes, err := generateTxBytes(msgVoteOracleRegistration, e.reactor.OracleAcc().GetPrivKey(), e.reactor.Config(), txBuilder)
	if err != nil {
		return err
	}

	if err := broadcastTx(e.reactor.GRPCClient(), txBytes); err != nil {
		return err
	}

	return nil
}

func (e UpgradeOracleEvent) verifyAndGetVoteOption(r *oracletypes.OracleRegistration) (oracletypes.VoteOption, error) {
	upgradeInfo, err := e.reactor.QueryClient().GetOracleUpgradeInfo()
	if err != nil {
		if errors.Is(err, panacea.ErrEmptyValue) {
			log.Infof("not exist oracle upgrade info")
			return oracletypes.VOTE_OPTION_NO, nil
		}
		log.Errorf("failed to get oracle upgrade info. %v", err)
		return oracletypes.VOTE_OPTION_UNSPECIFIED, err
	}
	if upgradeInfo.UniqueId != r.UniqueId {
		log.Infof("oracle's uniqueID does not match the uniqueID being upgraded. expected uniqueID(%s), oracle's uniqueID(%s), ",
			upgradeInfo.UniqueId,
			r.UniqueId)
		return oracletypes.VOTE_OPTION_NO, nil
	}
	return verifyAndGetVoteOption(e.reactor, r)
}
