package main

import (
	"fmt"
	"os"

	"coldsign/intent"
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
}
