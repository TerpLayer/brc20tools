package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type getBalanceResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Total  int `json:"total"`
		Height int `json:"height"`
		Offset int `json:"offset"`
		Items  []struct {
			Ticker           string `json:"ticker"`
			OverallBalance   string `json:"overall_balance"`
			TransferBalance  string `json:"transfer_balance"`
			AvailableBalance string `json:"available_balance"`
		} `json:"items"`
	} `json:"data"`
}

func getAddressSummary(address string, auth string) (*getBalanceResponse, error) {
	url := fmt.Sprintf("https://testnet-api.merlinprotocol.org/apis/indexer/v1/address/%s/brc20/summary", address)
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", auth))

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	result := &getBalanceResponse{}
	err = json.Unmarshal(body, result)
	return result, err
}

// {
//     "code": 0,
//     "msg": "success",
//     "data": {
//         "total": 1,
//         "height": 2580802,
//         "offset": 0,
//         "inscriptions": [
//             {
//                 "data": {
//                     "p": "brc-20",
//                     "op": "transfer",
//                     "amt": "100",
//                     "tick": "qwpo"
//                 },
//                 "inscription_id": "b8a95aebae6e845f9bbaefcd8c818677286945cd9c90ce3bc096430f13424c6di0",
//                 "satoshi": 7766279631452241920,
//                 "confirmations": 43
//             }
//         ]
//     }
// }

type getTransferAbleInscriptionsResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Total        int `json:"total"`
		Height       int `json:"height"`
		Offset       int `json:"offset"`
		Inscriptions []struct {
			Data struct {
				P    string `json:"p"`
				Op   string `json:"op"`
				Amt  string `json:"amt"`
				Tick string `json:"tick"`
			} `json:"data"`
			InscriptionId string `json:"inscription_id"`
			Satoshi       int    `json:"satoshi"`
			Confirmations int    `json:"confirmations"`
		} `json:"inscriptions"`
	} `json:"data"`
}

func getTransferAbleInscriptions(address string, auth string) (*getTransferAbleInscriptionsResponse, error) {
	url := fmt.Sprintf("https://testnet-api.merlinprotocol.org/apis/indexer/v1/address/%s/brc20/qwpo/transferable-inscriptions", address)
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", auth))

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	result := &getTransferAbleInscriptionsResponse{}
	err = json.Unmarshal(body, result)
	return result, err
}
