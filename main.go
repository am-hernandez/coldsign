package main

import (
	"fmt"
	"os"

	"coldsign/intent"
	"coldsign/policy"
	"coldsign/hd"
  "strings"

	"github.com/ethereum/go-ethereum/common"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("usage: coldsign <intent.json>")
		os.Exit(1)
	}

	path := os.Args[1]
	b, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("read error:", err)
		os.Exit(1)
	}

	in, err := intent.ParseEthSend(b)
	if err != nil {
		fmt.Println("intent error:", err)
		os.Exit(1)
	}

	fmt.Println("---- TX SUMMARY (ETH_SEND) ----")
	fmt.Println("ChainID:", in.ChainID)
	fmt.Println("From index:", in.From.Index)
	fmt.Println("To:", in.To)
	fmt.Println("Value (wei):", in.ValueWei)
	fmt.Println("Nonce:", in.Nonce)
	fmt.Println("MaxFeePerGas (wei):", in.MaxFeePerGasWei)
	fmt.Println("MaxPriorityFeePerGas (wei):", in.MaxPriorityFeePerGasWei)
	fmt.Println("------------------------------")

	p := policy.Default()
	if err := p.Enforce(in); err != nil {
		fmt.Println("policy violation:", err)
		os.Exit(1)
	}

	fmt.Println("Policy check: OK")

	mn := strings.TrimSpace(os.Getenv("COLD_MNEMONIC"))
	if mn == "" {
		fmt.Println("error: set COLD_MNEMONIC (BIP-39 mnemonic) in environment")
		os.Exit(1)
	}
	pass := os.Getenv("COLD_PASSPHRASE") // optional

	_, derivedAddr, err := hd.DeriveEthKey(mn, pass, in.From.Index)
	if err != nil {
		fmt.Println("hd derive error:", err)
		os.Exit(1)
	}

	if derivedAddr.Hex() != common.HexToAddress(in.FromAddress).Hex() {
		fmt.Println("error: derived fromAddress does not match intent")
		fmt.Println("derived:", derivedAddr.Hex())
		fmt.Println("intent: ", common.HexToAddress(in.FromAddress).Hex())
		os.Exit(1)
	}

	fmt.Println("From address verified:", derivedAddr.Hex())

}
