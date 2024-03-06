package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/btcsuite/btcd/wire"
)

type getAddressResponse struct {
	Address    string `json:"address"`
	ChainStats struct {
		FundedTxoCount int `json:"funded_txo_count"`
		FundedTxoSum   int `json:"funded_txo_sum"`
		SpentTxoCount  int `json:"spent_txo_count"`
		SpentTxoSum    int `json:"spent_txo_sum"`
		TxCount        int `json:"tx_count"`
	} `json:"chain_stats"`
	MempoolStats struct {
		FundedTxoCount int `json:"funded_txo_count"`
		FundedTxoSum   int `json:"funded_txo_sum"`
		SpentTxoCount  int `json:"spent_txo_count"`
		SpentTxoSum    int `json:"spent_txo_sum"`
		TxCount        int `json:"tx_count"`
	} `json:"mempool_stats"`
}

func getAddress(address string) (*getAddressResponse, error) {
	url := fmt.Sprintf("https://mempool.space/testnet/api/address/%s", address)
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	result := &getAddressResponse{}
	err = json.Unmarshal(body, result)
	return result, err
}

// {
//     "txid": "0be85bfa63429b50eee34fb3402d8739bf044c490164bc5d9ece7d7d9b0cda3e",
//     "vout": 1,
//     "status": {
//         "confirmed": true,
//         "block_height": 2580617,
//         "block_hash": "00000000380efa2b321f30fc61024f5473b046a57fb540f2e3a7d27be4335ec0",
//         "block_time": 1709643054
//     },
//     "value": 41826
// }

type unspentUtxo struct {
	Txid   string `json:"txid"`
	Vout   int    `json:"vout"`
	Status struct {
		Confirmed   bool   `json:"confirmed"`
		BlockHeight int    `json:"block_height"`
		BlockHash   string `json:"block_hash"`
		BlockTime   int    `json:"block_time"`
	} `json:"status"`
	Value int `json:"value"`
}

func getUnspentUtxo(address string) ([]*unspentUtxo, error) {
	url := fmt.Sprintf("https://mempool.space/testnet/api/address/%s/utxo", address)
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	result := make([]*unspentUtxo, 0)
	err = json.Unmarshal(body, &result)
	return result, err
}

func getRawTransaction(txid string) (string, error) {
	url := fmt.Sprintf("https://mempool.space/testnet/api/tx/%s/hex", txid)
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return "", err
	}
	res, err := client.Do(req)
	if err != nil {

		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func getTransction(txid string) (*wire.MsgTx, error) {
	raw, err := getRawTransaction(txid)
	if err != nil {
		return nil, err
	}
	tx := &wire.MsgTx{}
	data, err := hex.DecodeString(raw)
	if err != nil {
		return nil, err
	}
	err = tx.Deserialize(bytes.NewReader(data))
	return tx, err
}

func postTransaction(raw string) (string, error) {
	url := "https://mempool.space/testnet/api/tx"
	method := "POST"
	payload := strings.NewReader(raw)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "text/plain")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(body), err
}
