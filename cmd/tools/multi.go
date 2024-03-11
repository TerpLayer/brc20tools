package main

import (
	"fmt"
	"strconv"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

const INSCRIPTION_ID_LEN = 66

func createTx(fromMultiAddress string, to string, inscriptionId string, feeFrom string, feerate int64) (*wire.MsgTx, error) {
	if len(inscriptionId) != INSCRIPTION_ID_LEN {
		return nil, fmt.Errorf("error inscription format")
	}
	inscriptionTxId := inscriptionId[:64]
	inscriptionN, err := strconv.Atoi(inscriptionId[65:])
	if err != nil {
		return nil, err
	}
	inputHash, err := chainhash.NewHashFromStr(inscriptionTxId)
	if err != nil {
		return nil, err
	}
	txIn := wire.NewTxIn(wire.NewOutPoint(inputHash, uint32(inscriptionN)), nil, nil)
	tx := wire.NewMsgTx(1)
	tx.AddTxIn(txIn)

	feeInputUtxo, err := chooseMaxUtxo(feeFrom)
	if err != nil {
		return nil, err
	}
	feeInputHash, err := chainhash.NewHashFromStr(feeInputUtxo.Txid)
	if err != nil {
		return nil, err
	}
	feeTxIn := wire.NewTxIn(wire.NewOutPoint(feeInputHash, uint32(feeInputUtxo.Vout)), nil, nil)
	tx.AddTxIn(feeTxIn)

	decodedToAddr, err := btcutil.DecodeAddress(to, NET)
	if err != nil {
		return nil, err
	}
	toAddrByte, err := txscript.PayToAddrScript(decodedToAddr)
	if err != nil {
		return nil, err
	}
	txOut := wire.NewTxOut(BRC20AMOUNT, toAddrByte)
	tx.AddTxOut(txOut)

	docodedChangeAddr, err := btcutil.DecodeAddress(feeFrom, NET)
	if err != nil {
		return nil, err
	}
	changeAddrByte, err := txscript.PayToAddrScript(docodedChangeAddr)
	if err != nil {
		return nil, err
	}
	txChangeOut := wire.NewTxOut(0, changeAddrByte)
	tx.AddTxOut(txChangeOut)

	// fee := int64(tx.SerializeSize()) * feerate
	fee := 437 * feerate
	tx.TxOut[1].Value = int64(feeInputUtxo.Value) + BRC20AMOUNT - fee - BRC20AMOUNT
	return tx, nil
}

func signGasInput(tx *wire.MsgTx, wif *btcutil.WIF, idx int) (*wire.MsgTx, error) {
	inputUtxo := tx.TxIn[idx].PreviousOutPoint
	preInput, err := getTransction(inputUtxo.Hash.String())
	if err != nil {
		return nil, err
	}
	fetcher := txscript.NewCannedPrevOutputFetcher(
		preInput.TxOut[inputUtxo.Index].PkScript, preInput.TxOut[inputUtxo.Index].Value,
	)
	sighashes := txscript.NewTxSigHashes(tx, fetcher)
	wit, err := txscript.WitnessSignature(
		tx,
		sighashes,
		idx,
		preInput.TxOut[inputUtxo.Index].Value,
		preInput.TxOut[inputUtxo.Index].PkScript,
		txscript.SigHashAll,
		wif.PrivKey,
		true,
	)
	if err != nil {
		return nil, err
	}
	tx.TxIn[idx].Witness = wit
	return tx, nil
}

func signMultiInput(tx *wire.MsgTx, redeemScript []byte, wif *btcutil.WIF, idx int) (*wire.MsgTx, []byte, error) {
	signature, err := txscript.RawTxInSignature(tx, idx, redeemScript, txscript.SigHashAll, wif.PrivKey)
	if err != nil {
		return nil, nil, err
	}
	return tx, signature, nil
}

func signMultiInputFinal(tx *wire.MsgTx, redeemScript []byte, wif *btcutil.WIF, idx int, signature []byte) (*wire.MsgTx, error) {
	signatureFinal, err := txscript.RawTxInSignature(tx, idx, redeemScript, txscript.SigHashAll, wif.PrivKey)
	if err != nil {
		return nil, err
	}
	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_FALSE).AddData(signature).AddData(signatureFinal).AddData(redeemScript)
	signatureScript, err := builder.Script()
	if err != nil {
		return nil, err
	}
	tx.TxIn[idx].SignatureScript = signatureScript
	return tx, nil
}
