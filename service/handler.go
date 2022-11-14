package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/medibloc/panacea-doracle/crypto"
	log "github.com/sirupsen/logrus"

	datadealtypes "github.com/medibloc/panacea-core/v2/x/datadeal/types"
)

type SellDataReq struct {
	SellerAddress string `json:"seller_address"`
	DealID        uint64 `json:"deal_id"`
	DataHash      string `json:"data_hash"`
	EncryptedData string `json:"encrypted_data"`
}

func (s *Service) ValidateData(w http.ResponseWriter, r *http.Request) {
	// ***************** 1. validate data *****************
	var reqBody SellDataReq

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	queryClient := s.QueryClient()

	deal, err := queryClient.GetDeal(reqBody.DealID)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	encryptedDataBz, err := hex.DecodeString(reqBody.EncryptedData)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, "fail to decode encrypted data", http.StatusBadRequest)
		return
	}

	sellerAcc, err := queryClient.GetAccount(reqBody.SellerAddress)
	if err != nil {
		log.Errorf(err.Error())
		http.Error(w, "wrong seller address", http.StatusBadRequest)
		return
	}
	sellerPubKeyBytes := sellerAcc.GetPubKey().Bytes()

	oraclePrivKey := s.OraclePrivKey()

	sellerPubKey, err := btcec.ParsePubKey(sellerPubKeyBytes, btcec.S256())
	if err != nil {
		log.Error(err.Error())
		http.Error(w, "wrong seller pub key", http.StatusBadRequest)
		return
	}

	decryptSharedKey := crypto.DeriveSharedKey(oraclePrivKey, sellerPubKey, crypto.KDFSHA256)

	decryptedData, err := crypto.DecryptWithAES256(decryptSharedKey, deal.Nonce, encryptedDataBz)
	if err != nil {
		log.Errorf(err.Error())
		http.Error(w, "fail to decrypt", http.StatusBadRequest)
		return
	}

	if !compareDataHash(reqBody.DataHash, decryptedData) {
		http.Error(w, "hash mismatch", http.StatusBadRequest)
		return
	}

	// skip schema validation
	//if err := validation.ValidateJSONSchemata(decryptedData, deal.DataSchema); err != nil {
	//	http.Error(w, "invalid data schema", http.StatusBadRequest)
	//	return
	//}

	// ***************** 2. re-encrypt data *****************

	buyerAccount, err := queryClient.GetAccount(deal.BuyerAddress)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, "wrong buyer", http.StatusInternalServerError)
		return
	}

	buyerPubKeyBytes := buyerAccount.GetPubKey().Bytes()
	buyerPubKey, err := btcec.ParsePubKey(buyerPubKeyBytes, btcec.S256())
	if err != nil {
		log.Error(err.Error())
		http.Error(w, "wrong buyer pub key", http.StatusInternalServerError)
		return
	}

	encryptSharedKey := crypto.DeriveSharedKey(oraclePrivKey, buyerPubKey, crypto.KDFSHA256)

	encryptDataWithBuyerKey, err := crypto.EncryptWithAES256(encryptSharedKey, deal.Nonce, decryptedData)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, "failed to encrypt data", http.StatusInternalServerError)
		return
	}
	log.Infof("data is encrypted for buyer succesfully: %v", string(encryptDataWithBuyerKey))

	//// ipfs.add (decrypted data) & get CID
	//deliveredCid, err := s.Ipfs().Add(encryptDataWithBuyerKey)
	//if err != nil {
	//	http.Error(w, "failed to store data", http.StatusInternalServerError)
	//}
	tempDeliveredCID := "deliveredCID"

	uniqueID := s.EnclaveInfo().UniqueIDHex()

	// ***************** 3. sign *****************
	unsignedDataCert := &datadealtypes.UnsignedDataCert{
		OracleUniqueId: uniqueID,
		OracleAddress:  s.oracleAccount.GetAddress(),
		SellerAddress:  reqBody.SellerAddress,
		DealId:         reqBody.DealID,
		DataHash:       reqBody.DataHash,
		DeliveredCid:   tempDeliveredCID,
	}

	key := secp256k1.PrivKey{
		Key: oraclePrivKey.Serialize(),
	}

	marshaledDataCert, err := json.Marshal(unsignedDataCert)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, "fail to marshal unsignedDataCert", http.StatusInternalServerError)
		return
	}

	sig, err := key.Sign(marshaledDataCert)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, "fail to sign", http.StatusInternalServerError)
		return
	}

	payload := datadealtypes.MsgSellData{
		UnsignedDataCert: unsignedDataCert,
		Signature:        sig,
	}
	log.Infof("%v", payload)

	marshaledPayload, err := json.Marshal(payload)
	if err != nil {
		log.Error(err.Error())
		http.Error(w, "fail to marshal payload", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(marshaledPayload)
}

func compareDataHash(dataHash string, decryptedData []byte) bool {
	decryptedDataHash := sha256.Sum256(decryptedData)
	decryptedDataHashStr := hex.EncodeToString(decryptedDataHash[:])

	return decryptedDataHashStr == dataHash
}
