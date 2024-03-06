package main

import (
	"context"
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

// func getSegWitAddress(wif *btcutil.WIF) string {
// 	witnessProg := btcutil.Hash160(wif.SerializePubKey())
// 	addressWitnessPubKeyHash, _ := btcutil.NewAddressWitnessPubKeyHash(witnessProg, NET)
// 	return addressWitnessPubKeyHash.EncodeAddress()
// }

func getMultiAddress(wifs []*btcutil.WIF) (string, error) {
	addressPubKeys := make([]*btcutil.AddressPubKey, 0)
	for _, wif := range wifs {
		addressPubKey, err := btcutil.NewAddressPubKey(wif.SerializePubKey(), NET)
		if err != nil {
			return "", err
		}
		addressPubKeys = append(addressPubKeys, addressPubKey)
	}
	const nRequired = 2
	script, err := txscript.MultiSigScript(addressPubKeys, nRequired)
	if err != nil {
		return "", err
	}
	// log.Printf("redeemScript: %s", hex.EncodeToString(script))
	addr, err := btcutil.NewAddressScriptHashFromHash(btcutil.Hash160(script), NET)
	if err != nil {
		return "", err
	}
	return addr.EncodeAddress(), nil
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
	multiAddress, err := getMultiAddress(wifs)
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
	multiAddress, err := getMultiAddress(wifs)
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
		to, err = getMultiAddress(wifs)
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
