package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

func Test_Spent(t *testing.T) {
	rawTx := "020000000001013eda0c9b7d7dce9e5dbc6401494c04bf39872d40b34fe3ee509b4263fa5be80b0100000000ffffffff02e8030000000000001600145fba60e1ed8f336f68f51a4ad33821261e2dfa72ed9e000000000000160014e98933d4095bdf9a740e2bbef9427d9bc736300a02483045022100d3520ae5f9dc7a0cbd16ac5eaab0aa3618f35407ab850760c87b31500ce222ed022075f16f096aa258b5b794754985487d61307d21d7691095115b57b381d283eec5012103964bfe910fe53e59f56b177356f696689f58a6c124587f15bfe4d4130a1ae7bd00000000"
	// tx := wire.NewMsgTx(wire.TxVersion)
	// data, err := hex.DecodeString(rawTx)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// err = tx.Deserialize(bytes.NewReader(data))
	// if err != nil {
	// 	t.Fatal(err)
	// }
	wif, err := btcutil.DecodeWIF("cNGZ4vWwWD2SE9y6oQAbPs7miZxd4SayYh83Pe3bGWyWBqwhiLZH")
	if err != nil {
		t.Fatal(err)
	}
	tx := wire.NewMsgTx(wire.TxVersion)
	utxoHash, err := chainhash.NewHashFromStr("0be85bfa63429b50eee34fb3402d8739bf044c490164bc5d9ece7d7d9b0cda3e")
	if err != nil {
		t.Fatal(err)
	}
	txIn := wire.NewTxIn(wire.NewOutPoint(utxoHash, 1), nil, nil)
	tx.AddTxIn(txIn)

	//tb1qt7axpc0d3uek7684rf9dxwppyc0zm7njhwf6u4
	decodedAddr, err := btcutil.DecodeAddress("tb1qt7axpc0d3uek7684rf9dxwppyc0zm7njhwf6u4", NET)
	if err != nil {
		t.Fatal(err)
	}
	destinationAddrByte, err := txscript.PayToAddrScript(decodedAddr)
	if err != nil {
		t.Fatal(err)
	}
	// adding the destination address and the amount to the transaction
	txOut := wire.NewTxOut(1000, destinationAddrByte)
	tx.AddTxOut(txOut)

	decodedAddr2, err := btcutil.DecodeAddress("tb1qaxyn84qft00e5aqw9wl0jsnan0rnvvq2cvhrsh", NET)
	if err != nil {
		t.Fatal(err)
	}
	destinationAddrByte2, err := txscript.PayToAddrScript(decodedAddr2)
	if err != nil {
		t.Fatal(err)
	}
	// adding the destination address and the amount to the transaction
	txOut2 := wire.NewTxOut(1000, destinationAddrByte2)
	tx.AddTxOut(txOut2)
	pkScript, err := hex.DecodeString("0014e98933d4095bdf9a740e2bbef9427d9bc736300a")
	if err != nil {
		t.Fatal(err)
	}
	fetcher := txscript.NewCannedPrevOutputFetcher(
		pkScript, 41826,
	)
	// for idx, txin := range tx.TxIn {
	// 	fetcher.AddPrevOut(txin.PreviousOutPoint, &wire.TxOut{
	// 		Value:    int64(inputValues[idx]),
	// 		PkScript: prevPkScripts[idx],
	// 	})
	// }
	sighashes := txscript.NewTxSigHashes(tx, fetcher)
	wit, err := txscript.WitnessSignature(
		tx,
		sighashes,
		0,
		41826,
		pkScript,
		txscript.SigHashAll,
		wif.PrivKey,
		true,
	)
	if err != nil {
		t.Fatal(err)
	}
	tx.TxIn[0].Witness = wit
	txBuffer := make([]byte, 0)
	err = tx.Serialize(bytes.NewBuffer(txBuffer))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("hex.EncodeToString(txBuffer): %v\n", hex.EncodeToString(txBuffer))
	fmt.Printf("rawTx: %v\n", rawTx)

	// txscript.WitnessSignature(tx, )
}
