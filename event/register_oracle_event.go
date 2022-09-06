package event

import (
	"encoding/hex"
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/medibloc/panacea-doracle/sgx"
	log "github.com/sirupsen/logrus"
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
	OracleAcc() *panacea.OracleAccount
	OraclePrivKey() []byte
	TxBuilder() *panacea.TxBuilder
	Config() *config.Config
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

	fmt.Println(oracleRegistration)

	err = sgx.VerifyRemoteReport(oracleRegistration.NodePubKeyRemoteReport, oracleRegistration.NodePubKey, *e.reactor.EnclaveInfo())
	if err != nil {
		msgVoteOracleRegistrationNo, err := makeOracleRegistrationVote(uniqueID, e.reactor.OracleAcc().GetAddress(), addressValue, types.VOTE_OPTION_NO, e.reactor.OraclePrivKey())
		if err != nil {
			return err
		}

		txBytes, err := generateTxBytes(msgVoteOracleRegistrationNo, e.reactor.OracleAcc().GetPrivKey(), e.reactor.Config(), e.reactor.TxBuilder())
		if err != nil {
			return err
		}

		err = broadCastTx(e.reactor.GRPCClient(), txBytes)
		if err != nil {
			return err
		}
		return err
	}

	msgVoteOracleRegistrationYes, err := makeOracleRegistrationVote(uniqueID, e.reactor.OracleAcc().GetAddress(), addressValue, types.VOTE_OPTION_YES, e.reactor.OraclePrivKey())
	txBytes, err := generateTxBytes(msgVoteOracleRegistrationYes, e.reactor.OracleAcc().GetPrivKey(), e.reactor.Config(), e.reactor.TxBuilder())
	if err != nil {
		return err
	}

	err = broadCastTx(e.reactor.GRPCClient(), txBytes)
	if err != nil {
		return err
	}

	log.Info("oracle registration success")
	return nil
}

// makeOracleRegistrationVote makes a vote for oracle registration with VOTE_OPTION
func makeOracleRegistrationVote(uniqueID, voterAddr, votingTargetAddr string, voteOption types.VoteOption, oraclePrivKey []byte) (*types.MsgVoteOracleRegistration, error) {
	registrationVote := &types.OracleRegistrationVote{
		UniqueId:               uniqueID,
		VoterAddress:           voterAddr,
		VotingTargetAddress:    votingTargetAddr,
		VoteOption:             voteOption,
		EncryptedOraclePrivKey: oraclePrivKey,
	}

	key := secp256k1.PrivKey{
		Key: oraclePrivKey,
	}

	bytes, err := registrationVote.Marshal()

	sign, err := key.Sign(bytes)
	if err != nil {
		return nil, err
	}

	msgVoteOracleRegistrationNo := &types.MsgVoteOracleRegistration{
		OracleRegistrationVote: registrationVote,
		Signature:              sign,
	}

	return msgVoteOracleRegistrationNo, nil
}

// generateTxBytes
func generateTxBytes(msgVoteOracleRegistration *types.MsgVoteOracleRegistration, privKey cryptotypes.PrivKey, conf *config.Config, txBuilder *panacea.TxBuilder) ([]byte, error) {
	defaultFeeAmount, _ := sdk.ParseCoinsNormalized(conf.Panacea.DefaultFeeAmount)
	txBytes, err := txBuilder.GenerateSignedTxBytes(privKey, conf.Panacea.DefaultGasLimit, defaultFeeAmount, msgVoteOracleRegistration)
	if err != nil {
		return nil, err
	}

	return txBytes, nil
}

// broadCastTx
func broadCastTx(grpcClient *panacea.GrpcClient, txBytes []byte) error {
	resp, err := grpcClient.BroadcastTx(txBytes)
	if err != nil {
		return err
	}

	if resp.TxResponse.Code != 0 {
		return fmt.Errorf("register oracle vote failed: %v", resp.TxResponse.RawLog)
	}
	return nil
}
