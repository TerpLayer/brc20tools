package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/okx/go-wallet-sdk/coins/bitcoin/brc20"
)

func brc20Mint(from string, wif *btcutil.WIF, to string, ticker string, amount string, feerate int64, net *chaincfg.Params) (string, error) {
	commitPrivkey, err := btcec.NewPrivateKey()
	//test
	commitWfi, _ := btcutil.NewWIF(commitPrivkey, net, true)
	fmt.Println(commitWfi.String())
	//end test
	if err != nil {
		return "", err
	}
	contentType := "text/plain;charset=utf-8"
	body := []byte(fmt.Sprintf(`{"p":"brc-20","op":"%s","tick":"%s","amt":"%s"}`, "mint", ticker, amount))
	script, _ := brc20.CreateInscriptionScript(
		commitPrivkey,
		contentType,
		body)
	commitAddress, err := brc20.NewTapRootAddressWithScript(commitPrivkey, script, net)
	if err != nil {
		return "", err
	}
	// fmt.Println(to)
	const commitValue = int64(2000)
	commitTx, err := sendSatoshi(from, wif, commitAddress, commitValue, feerate, net)
	if err != nil {
		return "", err
	}

	//reveal
	inscription := brc20.NewInscription(contentType, body)
	builder := brc20.NewTxBuildV1(net)
	builder.AddInput(commitTx.TxHash().String(),
		0,
		hex.EncodeToString(commitPrivkey.Serialize()),
		commitAddress,
		strconv.Itoa(int(commitValue)),
		inscription,
	)
	const minValue = int64(546)
	builder.AddOutput(to, strconv.Itoa(int(minValue)))
	const TX_SIZE = int64(340)
	builder.AddOutput(from, strconv.Itoa(int(commitValue-TX_SIZE*feerate-minValue)))
	revealRaw, err := builder.Build()
	if err != nil {
		return "", nil
	}
	//print
	fmt.Println("privkey: ", hex.EncodeToString(commitPrivkey.Serialize()))
	var commitTxBuffer bytes.Buffer
	err = commitTx.Serialize(&commitTxBuffer)
	if err != nil {
		return "", nil
	}
	commitTxId, err := postTransaction(hex.EncodeToString(commitTxBuffer.Bytes()))
	if err != nil {
		return "", nil
	}
	fmt.Println("commit: ", commitTxId)
	revealTxId, err := postTransaction(revealRaw)
	if err != nil {
		return "", nil
	}
	fmt.Println("reveal: ", revealTxId)
	return fmt.Sprintf("%si%d", revealTxId, 0), nil
}
