package datadeal

import (
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
	log "github.com/sirupsen/logrus"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

type DataDeliveryVoteEvent struct {
	reactor event.Reactor
}

func NewDataDeliveryVoteEvent(s event.Reactor) DataDeliveryVoteEvent {
	return DataDeliveryVoteEvent{s}
}

func (e DataDeliveryVoteEvent) GetEventType() string {
	return "data_delivery"
}

func (e DataDeliveryVoteEvent) GetEventAttributeKey() string {
	return "vote_status"
}

func (e DataDeliveryVoteEvent) GetEventAttributeValue() string {
	return "'started'"
}

func (e DataDeliveryVoteEvent) EventHandler(event ctypes.ResultEvent) error {

	dealIDStr := event.Events[datadealtypes.EventTypeDataDeliveryVote+"."+datadealtypes.AttributeKeyDealID][0]
	dataHash := event.Events[datadealtypes.EventTypeDataDeliveryVote+"."+datadealtypes.AttributeKeyDataHash][0]

	dealID, err := strconv.ParseUint(dealIDStr, 10, 64)
	if err != nil {
		return err
	}

	voteOption, deliveredCid, err := e.verifyAndGetVoteOption(dealID, dataHash)
	if err != nil {
		if voteOption == oracletypes.VOTE_OPTION_UNSPECIFIED {
			return fmt.Errorf("can't vote due to error while verify. dealID(%d). dataHash(%s): %v", dealID, dataHash, err)
		} else {
			log.Infof("vote NO due to error while verify. dealID(%d). dataHash(%s): %v", dealID, dataHash, err)
		}
	}

	msgVoteDataDelivery, err := e.makeDataDeliveryVote(
		e.reactor.OracleAcc().GetAddress(),
		dataHash,
		deliveredCid,
		dealID,
		voteOption,
		e.reactor.OraclePrivKey().Serialize(),
	)
	if err != nil {
		return fmt.Errorf("make DataDeliveryVote failed. dealID(%d). dataHash(%s): %v", dealID, dataHash, err)
	}

	log.Infof("data delivery vote info. dealID(%d), dataHash(%s), deliveredCid(%s),voterAddress(%s), voteOption(%s)",
		msgVoteDataDelivery.DataDeliveryVote.DealId,
		msgVoteDataDelivery.DataDeliveryVote.DataHash,
		msgVoteDataDelivery.DataDeliveryVote.DeliveredCid,
		msgVoteDataDelivery.DataDeliveryVote.VoterAddress,
		msgVoteDataDelivery.DataDeliveryVote.VoteOption,
	)

	txBuilder := panacea.NewTxBuilder(*e.reactor.QueryClient())

	txBytes, err := e.generateTxBytes(msgVoteDataDelivery, e.reactor.OracleAcc().GetPrivKey(), e.reactor.Config(), txBuilder)
	if err != nil {
		return fmt.Errorf("generate tx failed. dealID(%d). dataHash(%s): %v", dealID, dataHash, err)
	}

	if err := e.broadcastTx(e.reactor.GRPCClient(), txBytes); err != nil {
		return fmt.Errorf("broadcast transaction failed. dealID(%d). dataHash(%s): %v", dealID, dataHash, err)
	}

	return nil

}

func (e DataDeliveryVoteEvent) verifyAndGetVoteOption(dealID uint64, dataHash string) (oracletypes.VoteOption, string, error) {

	dataSale, err := e.reactor.QueryClient().GetDataSale(dataHash, dealID)
	if err != nil {
		return oracletypes.VOTE_OPTION_UNSPECIFIED, "", fmt.Errorf("failed to get dataSale. %v", err)
	}

	if dataSale.Status != datadealtypes.DATA_SALE_STATUS_DELIVERY_VOTING_PERIOD {
		return oracletypes.VOTE_OPTION_UNSPECIFIED, "", errors.New("datasale status is not DATA_SALE_STATUS_DELIVERY_VOTING_PERIOD")
	}

	if len(dataSale.VerifiableCid) == 0 {
		return oracletypes.VOTE_OPTION_NO, "", errors.New("there is no verifiableCid")
	}

	deal, err := e.reactor.QueryClient().GetDeal(dataSale.DealId)
	if err != nil {
		if errors.Is(err, panacea.ErrEmptyValue) {
			return oracletypes.VOTE_OPTION_NO, "", fmt.Errorf("not found deal. %v", err)
		} else {
			return oracletypes.VOTE_OPTION_UNSPECIFIED, "", fmt.Errorf("failed to get deal. %v", err)
		}
	}

	deliveredCID, err := e.convertBuyerDataAndAddToIpfs(deal, dataSale, e.reactor.OraclePrivKey())
	if err != nil {
		return oracletypes.VOTE_OPTION_NO, "", fmt.Errorf("error while make deliveredCid: %v", err)
	}

	return oracletypes.VOTE_OPTION_YES, deliveredCID, nil

}

func (e DataDeliveryVoteEvent) convertBuyerDataAndAddToIpfs(deal *datadealtypes.Deal, dataSale *datadealtypes.DataSale, oraclePrivKey *btcec.PrivateKey) (string, error) {
	// get encrypted data from ipfs
	encryptedDataBz, err := e.reactor.Ipfs().Get(dataSale.VerifiableCid)
	if err != nil {
		return "", fmt.Errorf("failed to get data from ipfs. verifiableCid(%s) .%v", dataSale.VerifiableCid, err)
	}

	// get shared key oraclePrivKey + sellerPublicKey
	sellerAcc, err := e.reactor.QueryClient().GetAccount(dataSale.SellerAddress)
	if err != nil {
		return "", fmt.Errorf("failed to get seller account. %v", err)
	}
	sellerPubKeyBytes := sellerAcc.GetPubKey().Bytes()

	sellerPubKey, err := btcec.ParsePubKey(sellerPubKeyBytes, btcec.S256())
	if err != nil {
		return "", fmt.Errorf("failed to parse seller public key. %v", err)
	}
	decryptSharedKey := crypto.DeriveSharedKey(oraclePrivKey, sellerPubKey, crypto.KDFSHA256)

	// decrypt data
	decryptedData, err := crypto.DecryptWithAES256(decryptSharedKey, deal.Nonce, encryptedDataBz)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt data. %v", err)
	}

	// get oraclePrivateKey & buyerPublicKey and make shared key
	buyerAccount, err := e.reactor.QueryClient().GetAccount(deal.BuyerAddress)
	if err != nil {
		return "", fmt.Errorf("failed to get buyer account. %v", err)
	}

	buyerPubKeyBytes := buyerAccount.GetPubKey().Bytes()
	buyerPubKey, err := btcec.ParsePubKey(buyerPubKeyBytes, btcec.S256())
	if err != nil {
		return "", fmt.Errorf("failed to parse buyer public key. %v", err)
	}
	encryptSharedKey := crypto.DeriveSharedKey(oraclePrivKey, buyerPubKey, crypto.KDFSHA256)

	// encrypt data
	encryptDataWithBuyerKey, err := crypto.EncryptWithAES256(encryptSharedKey, deal.Nonce, decryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt data. %v", err)
	}

	// ipfs.add (decrypted data) & get CID
	deliveredCid, err := e.reactor.Ipfs().Add(encryptDataWithBuyerKey)
	if err != nil {
		return "", fmt.Errorf("failed to add data to Ipfs. %v", err)
	}

	return deliveredCid, nil
}

func (e DataDeliveryVoteEvent) makeDataDeliveryVote(voterAddress, dataHash, deliveredCid string, dealID uint64, voteOption oracletypes.VoteOption, oraclePrivKey []byte) (*datadealtypes.MsgVoteDataDelivery, error) {

	dataDeliveryVote := &datadealtypes.DataDeliveryVote{
		VoterAddress: voterAddress,
		DealId:       dealID,
		DataHash:     dataHash,
		DeliveredCid: deliveredCid,
		VoteOption:   voteOption,
	}

	key := secp256k1.PrivKey{
		Key: oraclePrivKey,
	}

	marshaledDataDeliveryVote, err := dataDeliveryVote.Marshal()
	if err != nil {
		return nil, err
	}

	sig, err := key.Sign(marshaledDataDeliveryVote)
	if err != nil {
		return nil, err
	}

	msgVoteDataDelivery := &datadealtypes.MsgVoteDataDelivery{
		DataDeliveryVote: dataDeliveryVote,
		Signature:        sig,
	}

	return msgVoteDataDelivery, nil
}

func (e DataDeliveryVoteEvent) generateTxBytes(msgVoteDataDelivery *datadealtypes.MsgVoteDataDelivery, privKey cryptotypes.PrivKey, conf *config.Config, txBuilder *panacea.TxBuilder) ([]byte, error) {
	defaultFeeAmount, _ := sdk.ParseCoinsNormalized(conf.Panacea.DefaultFeeAmount)
	txBytes, err := txBuilder.GenerateSignedTxBytes(privKey, conf.Panacea.DefaultGasLimit, defaultFeeAmount, msgVoteDataDelivery)
	if err != nil {
		return nil, err
	}

	return txBytes, nil
}

// broadcastTx broadcast transaction to blockchain.
func (e DataDeliveryVoteEvent) broadcastTx(grpcClient *panacea.GrpcClient, txBytes []byte) error {
	resp, err := grpcClient.BroadcastTx(txBytes)
	if err != nil {
		return err
	}

	if resp.TxResponse.Code != 0 {
		return fmt.Errorf("data delivery vote trasnsaction failed: %v", resp.TxResponse.RawLog)
	}

	log.Infof("MsgVoteDataDelivery transaction succeed. height(%v), hash(%s)", resp.TxResponse.Height, resp.TxResponse.TxHash)

	return nil
}
