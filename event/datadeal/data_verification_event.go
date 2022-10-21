package event

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/crypto"
	"github.com/medibloc/panacea-doracle/event"
	"github.com/medibloc/panacea-doracle/ipfs"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/medibloc/panacea-doracle/sgx"
	"github.com/medibloc/panacea-doracle/validation"
	log "github.com/sirupsen/logrus"
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
	Ipfs() *ipfs.Ipfs
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
	return "'SellData'"
}

func (d DataVerificationEvent) EventHandler(event ctypes.ResultEvent) error {
	dealIDStr := event.Events[types.EventTypeDataVerificationVote+"."+types.AttributeKeyDealID][0]
	dataHash := event.Events[types.EventTypeDataVerificationVote+"."+types.AttributeKeyDataHash][0]

	dealID, err := strconv.ParseUint(dealIDStr, 10, 64)
	if err != nil {
		return err
	}

	deal, err := d.reactor.QueryClient().GetDeal(dealID)
	if err != nil {
		return err
	}

	dataSale, err := d.reactor.QueryClient().GetDataSale(dealID, dataHash)
	if err != nil {
		return err
	}

	encryptedDataBz, err := d.reactor.Ipfs().Get(dataSale.VerifiableCid)
	if err != nil {
		return err
	}

	oraclePrivKey := d.reactor.OraclePrivKey()

	decryptedData, err := d.decryptData(oraclePrivKey, dataSale, deal, encryptedDataBz)
	if err != nil {
		return err
	}

	if !d.compareDataHash(dataSale, decryptedData) {
		return fmt.Errorf("invalid data hash (%s)", dataSale.DataHash)
	}

	voteOption := d.verifyDataSaleAndGetVoteOption(decryptedData, deal.DataSchema)

	msgVoteDataVerification, err := makeDataVerificationVote(
		d.reactor.OracleAcc().GetAddress(),
		dataHash,
		dealID,
		voteOption,
		oraclePrivKey.Serialize(),
	)
	if err != nil {
		return err
	}

	txBuilder := panacea.NewTxBuilder(*d.reactor.QueryClient())

	txBytes, err := generateTxBytes(msgVoteDataVerification, d.reactor.OracleAcc().GetPrivKey(), d.reactor.Config(), txBuilder)
	if err != nil {
		return err
	}

	if err := broadcastTx(d.reactor.GRPCClient(), txBytes); err != nil {
		return err
	}

	return nil
}

func (d DataVerificationEvent) decryptData(oraclePrivKey *btcec.PrivateKey, dataSale *types.DataSale, deal *types.Deal, encryptedDataBz []byte) ([]byte, error) {
	sellerAcc, err := d.reactor.QueryClient().GetAccount(dataSale.SellerAddress)
	if err != nil {
		return nil, err
	}

	sellerPubKeyBytes := sellerAcc.GetPubKey().Bytes()
	sellerPubKey, err := btcec.ParsePubKey(sellerPubKeyBytes, btcec.S256())
	if err != nil {
		return nil, err
	}

	decryptSharedKey := crypto.DeriveSharedKey(oraclePrivKey, sellerPubKey, crypto.KDFSHA256)

	decryptedData, err := crypto.DecryptWithAES256(decryptSharedKey, deal.Nonce, encryptedDataBz)
	if err != nil {
		return nil, err
	}
	return decryptedData, nil
}

func (d DataVerificationEvent) compareDataHash(dataSale *types.DataSale, decryptedData []byte) bool {
	decryptedDataHash := sha256.Sum256(decryptedData)
	decryptedDataHashStr := hex.EncodeToString(decryptedDataHash[:])

	return decryptedDataHashStr != dataSale.DataHash
}

func (d DataVerificationEvent) verifyDataSaleAndGetVoteOption(jsonInput []byte, dataSchema []string) oracletypes.VoteOption {
	err := validation.ValidateJSONSchemata(jsonInput, dataSchema)
	if err != nil {
		log.Warnf("failed to verify data. data(%s)", string(jsonInput))
		return oracletypes.VOTE_OPTION_NO
	}

	return oracletypes.VOTE_OPTION_YES
}

func makeDataVerificationVote(voterAddress, dataHash string, dealID uint64, voteOption oracletypes.VoteOption, oraclePrivKey []byte) (*types.MsgVoteDataVerification, error) {
	dataVerificationVote := &types.DataVerificationVote{
		VoterAddress: voterAddress,
		DealId:       dealID,
		DataHash:     dataHash,
		VoteOption:   voteOption,
	}

	key := secp256k1.PrivKey{
		Key: oraclePrivKey,
	}

	marshaledVerificationVote, err := dataVerificationVote.Marshal()
	if err != nil {
		return nil, err
	}

	sig, err := key.Sign(marshaledVerificationVote)
	if err != nil {
		return nil, err
	}

	msgVoteDataVerification := &types.MsgVoteDataVerification{
		DataVerificationVote: dataVerificationVote,
		Signature:            sig,
	}

	return msgVoteDataVerification, nil
}

// generateTxBytes generates transaction byte array.
// TODO: generateTxBytes function will be refactored.
func generateTxBytes(msgVoteDataVerification *types.MsgVoteDataVerification, privKey cryptotypes.PrivKey, conf *config.Config, txBuilder *panacea.TxBuilder) ([]byte, error) {
	defaultFeeAmount, _ := sdk.ParseCoinsNormalized(conf.Panacea.DefaultFeeAmount)
	txBytes, err := txBuilder.GenerateSignedTxBytes(privKey, conf.Panacea.DefaultGasLimit, defaultFeeAmount, msgVoteDataVerification)
	if err != nil {
		return nil, err
	}

	return txBytes, nil
}

// broadcastTx broadcast transaction to blockchain.
// TODO: broadcastTx function will be refactored.
func broadcastTx(grpcClient *panacea.GrpcClient, txBytes []byte) error {
	resp, err := grpcClient.BroadcastTx(txBytes)
	if err != nil {
		return err
	}

	if resp.TxResponse.Code != 0 {
		return fmt.Errorf("data verification vote trasnsaction failed: %v", resp.TxResponse.RawLog)
	}

	log.Infof("MsgVoteDataVerification transaction succeed. height(%v), hash(%s)", resp.TxResponse.Height, resp.TxResponse.TxHash)

	return nil
}
