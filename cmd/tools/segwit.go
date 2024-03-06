package main

import (
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

func sendSatoshi(from string, wif *btcutil.WIF, to string, value int64, feerate int64, net *chaincfg.Params) (*wire.MsgTx, error) {
	inputUtxo, err := chooseMaxUtxo(from)
	if err != nil {
		return nil, err
	}
	inputHash, err := chainhash.NewHashFromStr(inputUtxo.Txid)
	if err != nil {
		return nil, err
	}
	txIn := wire.NewTxIn(wire.NewOutPoint(inputHash, uint32(inputUtxo.Vout)), nil, nil)
	tx := wire.NewMsgTx(2)
	tx.AddTxIn(txIn)
	//add to output
	decodedToAddr, err := btcutil.DecodeAddress(to, NET)
	if err != nil {
		return nil, err
	}
	toAddrByte, err := txscript.PayToAddrScript(decodedToAddr)
	if err != nil {
		return nil, err
	}
	txOut := wire.NewTxOut(value, toAddrByte)
	tx.AddTxOut(txOut)
	//add change output
	docodedChangeAddr, err := btcutil.DecodeAddress(from, NET)
	if err != nil {
		return nil, err
	}
	changeAddrByte, err := txscript.PayToAddrScript(docodedChangeAddr)
	if err != nil {
		return nil, err
	}
	txChangeOut := wire.NewTxOut(0, changeAddrByte)
	tx.AddTxOut(txChangeOut)
	fee := int64(tx.SerializeSize()) * feerate
	tx.TxOut[1].Value = int64(inputUtxo.Value) - fee - value
	preInput, err := getTransction(inputUtxo.Txid)
	if err != nil {
		return nil, err
	}
	fetcher := txscript.NewCannedPrevOutputFetcher(
		preInput.TxOut[inputUtxo.Vout].PkScript, preInput.TxOut[inputUtxo.Vout].Value,
	)
	sighashes := txscript.NewTxSigHashes(tx, fetcher)
	wit, err := txscript.WitnessSignature(
		tx,
		sighashes,
		0,
		preInput.TxOut[inputUtxo.Vout].Value,
		preInput.TxOut[inputUtxo.Vout].PkScript,
		txscript.SigHashAll,
		wif.PrivKey,
		true,
	)
	if err != nil {
		return nil, err
	}
	tx.TxIn[0].Witness = wit
	//test
	// var txRaw bytes.Buffer
	// err = tx.Serialize(&txRaw)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("wit: %v\n", tx.SerializeSize())
	// fmt.Printf("tx.TxHash().String(): %v\n", tx.TxHash().String())
	return tx, nil
}

func chooseMaxUtxo(from string) (*unspentUtxo, error) {
	utxos, err := getUnspentUtxo(from)
	if err != nil {
		return nil, err
	}
	var result *unspentUtxo
	for _, uu := range utxos {
		if result == nil {
			result = uu
			continue
		}
		if result.Value < uu.Value {
			result = uu
		}
	}
	return result, nil
}
