package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func die(msg string, err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, msg+":", err)
	} else {
		fmt.Fprintln(os.Stderr, msg)
	}
	os.Exit(1)
}

func readInput(arg string) (string, error) {
	if arg != "" {
		return arg, nil
	}
	b, err := os.ReadFile("/dev/stdin")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func main() {
	var arg string
	if len(os.Args) == 2 {
		arg = os.Args[1]
	} else if len(os.Args) > 2 {
		die("usage: decode_rawtx <0xRAW_TX_HEX>  (or pipe via stdin)", nil)
	}

	rawHex, err := readInput(arg)
	if err != nil {
		die("read input", err)
	}

	rawHex = strings.TrimSpace(rawHex)
	rawHex = strings.Trim(rawHex, "\"") // tolerate JSON-quoted strings
	rawHex = strings.TrimSpace(rawHex)

	if strings.HasPrefix(rawHex, "0x") {
		rawHex = rawHex[2:]
	}
	rawHex = strings.TrimSpace(rawHex)

	rawBytes, err := hex.DecodeString(rawHex)
	if err != nil {
		die("hex decode", err)
	}

	var tx types.Transaction
	if err := tx.UnmarshalBinary(rawBytes); err != nil {
		die("tx decode (UnmarshalBinary)", err)
	}

	chainID := tx.ChainId()
	signer := types.LatestSignerForChainID(chainID)

	from, err := types.Sender(signer, &tx)
	if err != nil {
		die("recover sender", err)
	}

	// Print decoded fields
	fmt.Println("---- DECODED TX ----")
	fmt.Println("Type:", tx.Type())
	fmt.Println("Hash:", tx.Hash().Hex())
	fmt.Println("ChainID:", chainID.String())
	fmt.Println("From:", from.Hex())

	to := tx.To()
	if to == nil {
		fmt.Println("To: <contract creation>")
	} else {
		fmt.Println("To:", to.Hex())
	}

	fmt.Println("Nonce:", tx.Nonce())
	fmt.Println("ValueWei:", tx.Value().String())
	fmt.Println("GasLimit:", tx.Gas())

	// Fee fields (EIP-1559 aware)
	if tx.Type() == 2 {
		fmt.Println("MaxFeePerGasWei:", tx.GasFeeCap().String())
		fmt.Println("MaxPriorityFeePerGasWei:", tx.GasTipCap().String())
	} else {
		fmt.Println("GasPriceWei:", tx.GasPrice().String())
	}

	data := tx.Data()
	fmt.Println("DataLen:", len(data))
	if len(data) > 0 {
		// Print first bytes only (avoid huge spam)
		n := len(data)
		if n > 32 {
			n = 32
		}
		fmt.Println("DataPrefix:", "0x"+hex.EncodeToString(data[:n]))
	}

	// Quick invariant checks for Phase 0 ETH_SEND
	fmt.Println("---- QUICK CHECKS ----")
	fmt.Println("Gas == 21000:", tx.Gas() == 21000)
	fmt.Println("Data empty:", len(data) == 0)

	// Ensure To is non-zero address if present
	if to != nil {
		fmt.Println("To != 0x0:", *to != (common.Address{}))
	}

	// Sanity: re-encode and compare bytes (roundtrip)
	reb, err := tx.MarshalBinary()
	if err != nil {
		die("marshal binary", err)
	}
	fmt.Println("Roundtrip equal:", bytes.Equal(rawBytes, reb))
}
