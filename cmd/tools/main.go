package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/joho/godotenv"
	"github.com/okx/go-wallet-sdk/coins/bitcoin"
	"github.com/urfave/cli/v3"
)

var NET = &chaincfg.TestNet3Params

const TICK = "qwpo"
const AMOUNT = "1000"
const BRC20AMOUNT = 546

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	cmd := &cli.Command{
		Commands: []*cli.Command{
			{
				Name:    "keys",
				Aliases: []string{"k"},
				Usage:   "print public keys",
				Action:  keys,
			},
			{
				Name:    "new",
				Aliases: []string{"n"},
				Usage:   "new a private key",
				Action:  newPrivateKey,
			},
			{
				Name:    "balance",
				Aliases: []string{"b"},
				Usage:   "print signer balance",
				Action:  printBalance,
			},
			{
				Name:    "mint",
				Aliases: []string{"m"},
				Usage:   fmt.Sprintf("mint %s to address", TICK),
				Action:  mint,
			},
			{
				Name:    "inscribe-transfer",
				Aliases: []string{"it"},
				Usage:   "inscribe a transfer inscription to mutilsig address",
				Action:  inscribeTransferFunc,
			},
			{
				Name:    "list-inscriptions",
				Aliases: []string{"li"},
				Usage:   "list all inscription on address",
				Action:  listInscriptions,
			},
			{
				Name:    "send-inscription",
				Aliases: []string{"s"},
				Usage:   "send inscription from multisig to address",
				Action:  sendInscription,
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func getWIFs() ([]*btcutil.WIF, error) {
	wif1, err := btcutil.DecodeWIF(os.Getenv("REDEEM_SERVICES"))
	if err != nil {
		return nil, err
	}
	wifs := make([]*btcutil.WIF, 0)
	wifs = append(wifs, wif1)

	wif2, err := btcutil.DecodeWIF(os.Getenv("TREASURY_SERVICES"))
	if err != nil {
		return nil, err
	}
	wifs = append(wifs, wif2)

	wif3, err := btcutil.DecodeWIF(os.Getenv("TREASURY_BACKUP"))
	if err != nil {
		return nil, err
	}
	wifs = append(wifs, wif3)
	return wifs, nil
}

func getMultiAddress(wifs []*btcutil.WIF) (string, []byte, error) {
	addressPubKeys := make([]*btcutil.AddressPubKey, 0)
	for _, wif := range wifs {
		addressPubKey, err := btcutil.NewAddressPubKey(wif.SerializePubKey(), NET)
		if err != nil {
			return "", nil, err
		}
		addressPubKeys = append(addressPubKeys, addressPubKey)
	}
	const nRequired = 2
	script, err := txscript.MultiSigScript(addressPubKeys, nRequired)
	if err != nil {
		return "", nil, err
	}
	// log.Printf("redeemScript: %s", hex.EncodeToString(script))
	addr, err := btcutil.NewAddressScriptHashFromHash(btcutil.Hash160(script), NET)
	if err != nil {
		return "", nil, err
	}
	return addr.EncodeAddress(), script, nil
}

func keys(ctx context.Context, cmd *cli.Command) error {
	wifs, err := getWIFs()
	if err != nil {
		return err
	}

	for i, wif := range wifs {
		address, err := bitcoin.PubKeyToAddr(wif.SerializePubKey(), bitcoin.SEGWIT_NATIVE, NET)
		if err != nil {
			return err
		}
		log.Printf("signer%d's address: %s", i, address)
	}
	multiAddress, _, err := getMultiAddress(wifs)
	if err != nil {
		return err
	}
	log.Printf("multi address: %s", multiAddress)
	return nil
}

func newPrivateKey(context.Context, *cli.Command) error {
	privkey, err := btcec.NewPrivateKey()
	if err != nil {
		return err
	}

	wif, err := btcutil.NewWIF(privkey, NET, true)
	if err != nil {
		return err
	}
	log.Printf("WIF: %v", wif.String())
	return nil
}

func printBalance(context.Context, *cli.Command) error {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Address(SegWit)", "Satoshi", fmt.Sprintf("BRC20(%s) available", TICK), fmt.Sprintf("BRC20(%s) transfer", TICK)})
	t.AppendSeparator()
	wifs, err := getWIFs()
	if err != nil {
		return err
	}
	for i, wif := range wifs {
		address, err := bitcoin.PubKeyToAddr(wif.SerializePubKey(), bitcoin.SEGWIT_NATIVE, NET)
		if err != nil {
			return err
		}
		getAddressResp, err := getAddress(address)
		if err != nil {
			return err
		}
		getAddressSummaryResp, err := getAddressSummary(address, os.Getenv("INDEXER_AUTH"))
		if err != nil {
			return err
		}
		if getAddressSummaryResp.Code != 0 {
			return fmt.Errorf("getAddressSummary code: %d", getAddressSummaryResp.Code)
		}
		tickerBalance := "0"
		transferBalance := "0"
		for _, item := range getAddressSummaryResp.Data.Items {
			if strings.ToLower(item.Ticker) == TICK {
				tickerBalance = item.AvailableBalance
				transferBalance = item.TransferBalance
			}
		}
		t.AppendRow([]interface{}{i, address, getAddressResp.ChainStats.FundedTxoSum, tickerBalance, transferBalance})
	}
	multiAddress, _, err := getMultiAddress(wifs)
	if err != nil {
		return err
	}
	getAddressResp, err := getAddress(multiAddress)
	if err != nil {
		return err
	}
	getAddressSummaryResp, err := getAddressSummary(multiAddress, os.Getenv("INDEXER_AUTH"))
	if err != nil {
		return err
	}
	if getAddressSummaryResp.Code != 0 {
		return fmt.Errorf("getAddressSummary code: %d", getAddressSummaryResp.Code)
	}
	tickerBalance := "0"
	transferBalance := "0"
	for _, item := range getAddressSummaryResp.Data.Items {
		if strings.ToLower(item.Ticker) == TICK {
			tickerBalance = item.AvailableBalance
			transferBalance = item.TransferBalance
		}
	}
	t.AppendRow([]interface{}{len(wifs), multiAddress, getAddressResp.ChainStats.FundedTxoSum, tickerBalance, transferBalance})
	t.Render()
	return nil
}

func mint(ctx context.Context, cli *cli.Command) error {
	toIndex := cli.Args().Get(0)
	index, err := strconv.ParseInt(toIndex, 10, 10)
	if err != nil {
		return err
	}
	wifs, err := getWIFs()
	if err != nil {
		return err
	}
	if index < 0 || index > int64(len(wifs)) {
		return fmt.Errorf("error to index: %s", toIndex)
	}
	to := ""
	if index == int64(len(wifs)) {
		to, _, err = getMultiAddress(wifs)
		if err != nil {
			return err
		}
	} else {
		to, err = bitcoin.PubKeyToAddr(wifs[index].SerializePubKey(), bitcoin.SEGWIT_NATIVE, NET)
		if err != nil {
			return err
		}
	}
	fmt.Printf("to: %v\n", to)
	from, err := bitcoin.PubKeyToAddr(wifs[1].SerializePubKey(), bitcoin.SEGWIT_NATIVE, NET)
	if err != nil {
		return err
	}
	feerate := int64(2)
	inscriptionId, err := brc20Mint(from, wifs[1], to, TICK, AMOUNT, feerate, NET)
	if err != nil {
		return err
	}
	fmt.Println("inscriptionId: ", inscriptionId)
	return nil
}

func inscribeTransferFunc(ctx context.Context, cli *cli.Command) error {
	toIndex := cli.Args().Get(0)
	index, err := strconv.ParseInt(toIndex, 10, 10)
	if err != nil {
		return err
	}
	wifs, err := getWIFs()
	if err != nil {
		return err
	}
	if index < 0 || index > int64(len(wifs)) {
		return fmt.Errorf("error to index: %s", toIndex)
	}
	to := ""
	if index == int64(len(wifs)) {
		to, _, err = getMultiAddress(wifs)
		if err != nil {
			return err
		}
	} else {
		to, err = bitcoin.PubKeyToAddr(wifs[index].SerializePubKey(), bitcoin.SEGWIT_NATIVE, NET)
		if err != nil {
			return err
		}
	}
	fmt.Printf("to: %v\n", to)
	from, err := bitcoin.PubKeyToAddr(wifs[1].SerializePubKey(), bitcoin.SEGWIT_NATIVE, NET)
	if err != nil {
		return err
	}
	feerate := int64(2)
	inscriptionId, err := inscribeTransfer(from, wifs[1], to, TICK, "100", feerate, NET)
	if err != nil {
		return err
	}
	fmt.Println("inscriptionId: ", inscriptionId)
	return nil
}

func listInscriptions(ctx context.Context, cli *cli.Command) error {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Address", "Ticker", "InscriptionId", "amount", "Confirmations"})
	t.AppendSeparator()
	wifs, err := getWIFs()
	if err != nil {
		return err
	}
	for i, wif := range wifs {
		address, err := bitcoin.PubKeyToAddr(wif.SerializePubKey(), bitcoin.SEGWIT_NATIVE, NET)
		if err != nil {
			return err
		}
		inscriptionRes, err := getTransferAbleInscriptions(address, os.Getenv("INDEXER_AUTH"))
		if err != nil {
			return err
		}
		if inscriptionRes.Code != 0 {
			return fmt.Errorf("getInscriptions code: %d", inscriptionRes.Code)
		}
		for _, item := range inscriptionRes.Data.Inscriptions {
			t.AppendRow([]interface{}{i, address, item.Data.Tick, item.InscriptionId, item.Data.Amt, item.Confirmations})
		}
	}

	multiAddress, _, err := getMultiAddress(wifs)
	if err != nil {
		return err
	}
	inscriptionRes, err := getTransferAbleInscriptions(multiAddress, os.Getenv("INDEXER_AUTH"))
	if err != nil {
		return err
	}
	if inscriptionRes.Code != 0 {
		return fmt.Errorf("getInscriptions code: %d", inscriptionRes.Code)
	}
	for _, item := range inscriptionRes.Data.Inscriptions {
		t.AppendRow([]interface{}{3, multiAddress, item.Data.Tick, item.InscriptionId, item.Data.Amt, item.Confirmations})
	}
	t.Render()
	return nil
}

func sendInscription(ctx context.Context, cli *cli.Command) error {
	toIndex := cli.Args().Get(0)
	index, err := strconv.ParseInt(toIndex, 10, 10)
	if err != nil {
		return err
	}
	wifs, err := getWIFs()
	if err != nil {
		return err
	}
	if index < 0 || index > int64(len(wifs)) {
		return fmt.Errorf("error to index: %s", toIndex)
	}
	to := ""
	if index == int64(len(wifs)) {
		to, _, err = getMultiAddress(wifs)
		if err != nil {
			return err
		}
	} else {
		to, err = bitcoin.PubKeyToAddr(wifs[index].SerializePubKey(), bitcoin.SEGWIT_NATIVE, NET)
		if err != nil {
			return err
		}
	}

	inscriptionId := cli.Args().Get(1)
	fmt.Printf("send %s to: %s\n", inscriptionId, to)
	fromMultiAddress, redeemScript, err := getMultiAddress(wifs)
	if err != nil {
		return err
	}
	gasWif := wifs[1]
	feeAddress, _ := bitcoin.PubKeyToAddr(gasWif.SerializePubKey(), bitcoin.SEGWIT_NATIVE, NET)
	const feerate = 3
	tx, err := createTx(fromMultiAddress, to, inscriptionId, feeAddress, feerate)
	if err != nil {
		return err
	}
	tx, err = signGasInput(tx, gasWif, 1)
	if err != nil {
		return err
	}
	tx, signature, err := signMultiInput(tx, redeemScript, wifs[0], 0)
	if err != nil {
		return err
	}
	tx, err = signMultiInputFinal(tx, redeemScript, wifs[1], 0, signature)
	if err != nil {
		return err
	}
	var buffer bytes.Buffer
	err = tx.Serialize(&buffer)
	if err != nil {
		return err
	}

	// fmt.Println(hex.EncodeToString(buffer.Bytes()))
	txId, err := postTransaction(hex.EncodeToString(buffer.Bytes()))
	if err != nil {
		return err
	}
	fmt.Println("txId: ", txId)
	return nil
}
