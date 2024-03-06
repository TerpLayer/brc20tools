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
