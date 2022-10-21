package oracle

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/crypto"
	"github.com/medibloc/panacea-doracle/event"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/medibloc/panacea-doracle/sgx"
	log "github.com/sirupsen/logrus"
	"github.com/tendermint/tendermint/light/provider"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var _ event.Event = (*RegisterOracleEvent)(nil)

type RegisterOracleEvent struct {
	reactor reactor
}

func NewRegisterOracleEvent(s reactor) RegisterOracleEvent {
	return RegisterOracleEvent{s}
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
	addressValue := event.Events[oracletypes.EventTypeRegistrationVote+"."+oracletypes.AttributeKeyOracleAddress][0]

	uniqueID := e.reactor.EnclaveInfo().UniqueIDHex()
	oracleRegistration, err := e.reactor.QueryClient().GetOracleRegistration(addressValue, uniqueID)
	if err != nil {
		return err
	}

	voteOption, err := verifyAndGetVoteOption(e.reactor, oracleRegistration)
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

// verifyAndGetVoteOption performs a verification to determine a vote.
// - Verify that trustedBlockInfo registered in OracleRegistration is valid
// - Verify that the RemoteReport is valid
func verifyAndGetVoteOption(reactor reactor, oracleRegistration *oracletypes.OracleRegistration) (oracletypes.VoteOption, error) {
	block, err := reactor.QueryClient().GetLightBlock(oracleRegistration.TrustedBlockHeight)
	if err != nil {
		switch err {
		case provider.ErrLightBlockNotFound, provider.ErrHeightTooHigh:
			return oracletypes.VOTE_OPTION_NO, nil
		default:
			return oracletypes.VOTE_OPTION_UNSPECIFIED, err
		}
	}

	if !bytes.Equal(block.Hash().Bytes(), oracleRegistration.TrustedBlockHash) {
		log.Warnf("failed to verify trusted block information. height(%v), expected block hash(%s), got block hash(%s)",
			oracleRegistration.TrustedBlockHeight,
			hex.EncodeToString(block.Hash().Bytes()),
			hex.EncodeToString(oracleRegistration.TrustedBlockHash),
		)
		return oracletypes.VOTE_OPTION_NO, nil
	}

	nodePubKeyHash := sha256.Sum256(oracleRegistration.NodePubKey)

	if err := sgx.VerifyRemoteReport(oracleRegistration.NodePubKeyRemoteReport, nodePubKeyHash[:], *reactor.EnclaveInfo()); err != nil {
		log.Warnf("failed to verification report. uniqueID(%s), address(%s), err(%v)", oracleRegistration.UniqueId, oracleRegistration.Address, err)
		return oracletypes.VOTE_OPTION_NO, nil
	} else {
		return oracletypes.VOTE_OPTION_YES, nil
	}
}

// makeOracleRegistrationVote makes a vote for oracle registration with VOTE_OPTION
func makeOracleRegistrationVote(uniqueID, voterAddr, votingTargetAddr string, voteOption oracletypes.VoteOption, oraclePrivKey, nodePubKey, nonce []byte) (*oracletypes.MsgVoteOracleRegistration, error) {
	privKey, _ := crypto.PrivKeyFromBytes(oraclePrivKey)

	pubKey, err := btcec.ParsePubKey(nodePubKey, btcec.S256())
	if err != nil {
		return nil, err
	}

	shareKey := crypto.DeriveSharedKey(privKey, pubKey, crypto.KDFSHA256)
	encryptedOraclePrivKey, err := crypto.EncryptWithAES256(shareKey, nonce, oraclePrivKey)
	if err != nil {
		return nil, err
	}

	registrationVote := &oracletypes.OracleRegistrationVote{
		UniqueId:               uniqueID,
		VoterAddress:           voterAddr,
		VotingTargetAddress:    votingTargetAddr,
		VoteOption:             voteOption,
		EncryptedOraclePrivKey: encryptedOraclePrivKey,
	}

	key := secp256k1.PrivKey{
		Key: oraclePrivKey,
	}

	marshaledRegistrationVote, err := registrationVote.Marshal()
	if err != nil {
		return nil, err
	}

	sig, err := key.Sign(marshaledRegistrationVote)
	if err != nil {
		return nil, err
	}

	msgVoteOracleRegistration := &oracletypes.MsgVoteOracleRegistration{
		OracleRegistrationVote: registrationVote,
		Signature:              sig,
	}

	return msgVoteOracleRegistration, nil
}

// generateTxBytes generates transaction byte array.
func generateTxBytes(msgVoteOracleRegistration *oracletypes.MsgVoteOracleRegistration, privKey cryptotypes.PrivKey, conf *config.Config, txBuilder *panacea.TxBuilder) ([]byte, error) {
	defaultFeeAmount, _ := sdk.ParseCoinsNormalized(conf.Panacea.DefaultFeeAmount)
	txBytes, err := txBuilder.GenerateSignedTxBytes(privKey, conf.Panacea.DefaultGasLimit, defaultFeeAmount, msgVoteOracleRegistration)
	if err != nil {
		return nil, err
	}

	return txBytes, nil
}

// broadcastTx broadcast transaction to blockchain.
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
