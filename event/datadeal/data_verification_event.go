package datadeal

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	datadealtypes "github.com/medibloc/panacea-core/v2/x/datadeal/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/crypto"
	"github.com/medibloc/panacea-doracle/event"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/medibloc/panacea-doracle/validation"
	log "github.com/sirupsen/logrus"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var _ event.Event = (*DataVerificationEvent)(nil)

type DataVerificationEvent struct {
	reactor event.Reactor
}

func NewDataVerificationEvent(r event.Reactor) DataVerificationEvent {
	return DataVerificationEvent{r}
}

func (e DataVerificationEvent) GetEventQuery() string {
	return "message.action = 'SellData'"
}

func (e DataVerificationEvent) EventHandler(event ctypes.ResultEvent) error {
	dealIDStr := event.Events[datadealtypes.EventTypeDataVerificationVote+"."+datadealtypes.AttributeKeyDealID][0]
	dataHash := event.Events[datadealtypes.EventTypeDataVerificationVote+"."+datadealtypes.AttributeKeyDataHash][0]

	dealID, err := strconv.ParseUint(dealIDStr, 10, 64)
	if err != nil {
		return err
	}

	voteOption, err := e.verifyAndGetVoteOption(dealID, dataHash)
	if err != nil {
		log.Infof("vote No due to error while verify. dealID(%d). dataHash(%s)", dealID, dataHash)
	}

	msgVoteDataVerification, err := makeDataVerificationVote(
		e.reactor.OracleAcc().GetAddress(),
		dataHash,
		dealID,
		voteOption,
		e.reactor.OraclePrivKey().Serialize(),
	)
	if err != nil {
		return err
	}

	txBuilder := panacea.NewTxBuilder(*e.reactor.QueryClient())

	txBytes, err := generateTxBytes(msgVoteDataVerification, e.reactor.OracleAcc().GetPrivKey(), e.reactor.Config(), txBuilder)
	if err != nil {
		return fmt.Errorf("generate tx failed. dealID(%d). dataHash(%s): %w", dealID, dataHash, err)
	}

	if err := broadcastTx(e.reactor.GRPCClient(), txBytes); err != nil {
		return fmt.Errorf("broadcast transaction failed. dealID(%d). dataHash(%s): %w", dealID, dataHash, err)
	}

	return nil
}

func (e DataVerificationEvent) decryptData(decryptedSharedKey, nonce, encryptedDataBz []byte) ([]byte, error) {
	decryptedData, err := crypto.DecryptWithAES256(decryptedSharedKey, nonce, encryptedDataBz)
	if err != nil {
		return nil, err
	}
	return decryptedData, nil
}

func (e DataVerificationEvent) compareDataHash(dataSale *datadealtypes.DataSale, decryptedData []byte) bool {
	decryptedDataHash := sha256.Sum256(decryptedData)
	decryptedDataHashStr := hex.EncodeToString(decryptedDataHash[:])

	return decryptedDataHashStr == dataSale.DataHash
}

func (e DataVerificationEvent) convertSellerData(deal *datadealtypes.Deal, dataSale *datadealtypes.DataSale) ([]byte, error) {
	encryptedDataBz, err := e.reactor.Ipfs().Get(dataSale.VerifiableCid)
	if err != nil {
		log.Infof("failed to get data from IPFS: %v", err)
		return nil, err
	}

	oraclePrivKey := e.reactor.OraclePrivKey()

	sellerAcc, err := e.reactor.QueryClient().GetAccount(dataSale.SellerAddress)
	if err != nil {
		return nil, err
	}

	sellerPubKeyBytes := sellerAcc.GetPubKey().Bytes()

	sellerPubKey, err := btcec.ParsePubKey(sellerPubKeyBytes, btcec.S256())
	if err != nil {
		return nil, err
	}

	decryptSharedKey := crypto.DeriveSharedKey(oraclePrivKey, sellerPubKey, crypto.KDFSHA256)

	decryptedData, err := e.decryptData(decryptSharedKey, deal.Nonce, encryptedDataBz)
	if err != nil {
		return nil, err
	}

	return decryptedData, nil
}

func (e DataVerificationEvent) verifyAndGetVoteOption(dealID uint64, dataHash string) (oracletypes.VoteOption, error) {
	deal, err := e.reactor.QueryClient().GetDeal(dealID)
	if err != nil {
		return oracletypes.VOTE_OPTION_NO, fmt.Errorf("failed to get deal. %v", err)
	}

	dataSale, err := e.reactor.QueryClient().GetDataSale(dataHash, dealID)
	if err != nil {
		return oracletypes.VOTE_OPTION_NO, fmt.Errorf("failed to get dataSale (%v)", err)
	}

	if dataSale.Status != datadealtypes.DATA_SALE_STATUS_VERIFICATION_VOTING_PERIOD {
		return oracletypes.VOTE_OPTION_NO, errors.New("dataSale's status is not DATA_SALE_STATUS_VERIFICATION_VOTING_PERIOD")
	}

	decryptedData, err := e.convertSellerData(deal, dataSale)
	if err != nil {
		return oracletypes.VOTE_OPTION_NO, fmt.Errorf("failed to decrypt seller data, error (%v)", err)
	}

	if !e.compareDataHash(dataSale, decryptedData) {
		log.Infof("invalid data hash")
		return oracletypes.VOTE_OPTION_NO, err
	}

	err = validation.ValidateJSONSchemata(decryptedData, deal.DataSchema)
	if err != nil {
		log.Infof("failed to verify data. error(%s)", err)
		return oracletypes.VOTE_OPTION_NO, err
	}

	return oracletypes.VOTE_OPTION_YES, nil
}

func makeDataVerificationVote(voterAddress, dataHash string, dealID uint64, voteOption oracletypes.VoteOption, oraclePrivKey []byte) (*datadealtypes.MsgVoteDataVerification, error) {
	dataVerificationVote := &datadealtypes.DataVerificationVote{
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

	msgVoteDataVerification := &datadealtypes.MsgVoteDataVerification{
		DataVerificationVote: dataVerificationVote,
		Signature:            sig,
	}

	return msgVoteDataVerification, nil
}

// generateTxBytes generates transaction byte array.
// TODO: generateTxBytes function will be refactored.
func generateTxBytes(msgVoteDataVerification *datadealtypes.MsgVoteDataVerification, privKey cryptotypes.PrivKey, conf *config.Config, txBuilder *panacea.TxBuilder) ([]byte, error) {
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
