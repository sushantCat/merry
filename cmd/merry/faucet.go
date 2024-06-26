package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/catalogfi/blockchain/localnet"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

func Faucet(state *State) *cobra.Command {
	var (
		to string
	)
	var cmd = &cobra.Command{
		Use:   "faucet",
		Short: "Generate and send bitcoin to given address",
		RunE: func(c *cobra.Command, args []string) error {
			if !state.Running {
				return fmt.Errorf("merry is not running")
			}

			_, err := btcutil.DecodeAddress(to, &chaincfg.RegressionNetParams)
			if err != nil {
				if len(to) == 42 {
					to = to[2:]
				}
				if len(to) == 40 {
					_, err := hex.DecodeString(to)
					if err != nil {
						return fmt.Errorf("to is not an ethereum or a bitcoin regtest address: %s", to)
					}
					return fundEVM(to)
				}
				return fmt.Errorf("to is not an ethereum or a bitcoin regtest address: %s", to)
			}

			return fundBTC(to)
		},
	}
	cmd.Flags().StringVar(&to, "to", "", "user should pass the address they needs to be funded")
	return cmd
}

func fundEVM(to string) error {
	ethAmount, _ := new(big.Int).SetString("1000000000000000000", 10)
	wbtcAmount, _ := new(big.Int).SetString("100000000", 10)
	tx, err := localnet.EVMWallet().Send(context.Background(), localnet.ETH(), common.HexToAddress(to), ethAmount)
	if err != nil {
		return fmt.Errorf("failed to send eth: %v", err)
	}
	fmt.Printf("Successfully sent %v ETH on Ethereum Localnet at: %s\n", ethAmount, tx.Hash().Hex())
	tx2, err := localnet.EVMWallet().Send(context.Background(), localnet.WBTC(), common.HexToAddress(to), wbtcAmount)
	if err != nil {
		return fmt.Errorf("failed to send eth: %v", err)
	}
	fmt.Printf("Successfully sent %v WBTC on Ethereum Localnet at: %s\n", wbtcAmount, tx2.Hash().Hex())
	tx3, err := localnet.EVMWallet().Send(context.Background(), localnet.ArbitrumETH(), common.HexToAddress(to), ethAmount)
	if err != nil {
		return fmt.Errorf("failed to send eth: %v", err)
	}
	fmt.Printf("Successfully sent %v ETH on Arbitrum Localnet at: %s\n", wbtcAmount, tx3.Hash().Hex())
	tx4, err := localnet.EVMWallet().Send(context.Background(), localnet.ArbitrumWBTC(), common.HexToAddress(to), wbtcAmount)
	if err != nil {
		return fmt.Errorf("failed to send eth: %v", err)
	}
	fmt.Printf("Successfully sent %v WBTC on Arbitrum Localnet at: %s\n", wbtcAmount, tx4.Hash().Hex())
	return nil
}

func fundBTC(to string) error {
	payload, err := json.Marshal(map[string]string{
		"address": to,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal address: %v", err)
	}

	res, err := http.Post("http://127.0.0.1:3000/faucet", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to get funds from faucet: %v", err)
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return errors.New(string(data))
	}
	var dat map[string]string
	if err := json.Unmarshal([]byte(data), &dat); err != nil {
		return errors.New("internal error, please try again")
	}
	if dat["txId"] == "" {
		return errors.New("not successful")
	}
	fmt.Println("Successfully submitted at http://localhost:5050/tx/" + dat["txId"])
	return nil
}
