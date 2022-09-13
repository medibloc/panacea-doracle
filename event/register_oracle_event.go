package event

import (
	"crypto/sha256"
	"fmt"
	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/crypto"
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
	OraclePrivKey() *btcec.PrivateKey
	Config() *config.Config
	QueryClient() *panacea.QueryClient
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

	uniqueID := e.reactor.EnclaveInfo().UniqueIDHex()
	oracleRegistration, err := e.reactor.QueryClient().GetOracleRegistration(addressValue, uniqueID)
	if err != nil {
		return err
	}

	txBuilder := panacea.NewTxBuilder(*e.reactor.QueryClient())

	voteOption := verifyReportAndGetVoteOption(oracleRegistration, e)

	msgVoteOracleRegistration, err := makeOracleRegistrationVote(uniqueID, e.reactor.OracleAcc().GetAddress(), addressValue, voteOption, e.reactor.OraclePrivKey().Serialize(), oracleRegistration.NodePubKey)
	if err != nil {
		return err
	}

	txBytes, err := generateTxBytes(msgVoteOracleRegistration, e.reactor.OracleAcc().GetPrivKey(), e.reactor.Config(), txBuilder)
	if err != nil {
		return err
	}

	if err := broadcastTx(e.reactor.GRPCClient(), txBytes); err != nil {
		return err
	}

	return nil
}

// verifyReportAndGetVoteOption validates the RemoteReport and returns the voting result according to the verification result.
func verifyReportAndGetVoteOption(oracleRegistration *types.OracleRegistration, e RegisterOracleEvent) types.VoteOption {
	nodePubKeyHash := sha256.Sum256(oracleRegistration.NodePubKey)

	if err := sgx.VerifyRemoteReport(oracleRegistration.NodePubKeyRemoteReport, nodePubKeyHash[:], *e.reactor.EnclaveInfo()); err != nil {
		log.Warnf("failed to verification report. uniqueID(%s), address(%s), err(%v)", oracleRegistration.UniqueId, oracleRegistration.Address, err)
		return types.VOTE_OPTION_NO
	} else {
		return types.VOTE_OPTION_YES
	}
}

// makeOracleRegistrationVote makes a vote for oracle registration with VOTE_OPTION
func makeOracleRegistrationVote(uniqueID, voterAddr, votingTargetAddr string, voteOption types.VoteOption, oraclePrivKey []byte, nodePubKey []byte) (*types.MsgVoteOracleRegistration, error) {
	pubKey, err := btcec.ParsePubKey(nodePubKey, btcec.S256())
	if err != nil {
		return nil, err
	}

	encryptedOraclePrivKey, err := crypto.Encrypt(pubKey, oraclePrivKey)
	if err != nil {
		return nil, err
	}

	registrationVote := &types.OracleRegistrationVote{
		UniqueId:               uniqueID,
		VoterAddress:           voterAddr,
		VotingTargetAddress:    votingTargetAddr,
		VoteOption:             voteOption,
		EncryptedOraclePrivKey: encryptedOraclePrivKey,
	}

	key := secp256k1.PrivKey{
		Key: oraclePrivKey,
	}

	bytes, err := registrationVote.Marshal()
	if err != nil {
		return nil, err
	}

	sig, err := key.Sign(bytes)
	if err != nil {
		return nil, err
	}

	msgVoteOracleRegistration := &types.MsgVoteOracleRegistration{
		OracleRegistrationVote: registrationVote,
		Signature:              sig,
	}

	return msgVoteOracleRegistration, nil
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

// broadcastTx
func broadcastTx(grpcClient *panacea.GrpcClient, txBytes []byte) error {
	resp, err := grpcClient.BroadcastTx(txBytes)
	if err != nil {
		return err
	}

	if resp.TxResponse.Code != 0 {
		return fmt.Errorf("register oracle vote trasnsaction failed: %v", resp.TxResponse.RawLog)
	}

	log.Infof("MsgVoteOracleRegistration transaction succeed. height(%v), hash(%s)", resp.TxResponse.Height, resp.TxResponse.TxHash)

	return nil
}
