package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"coldsign/hd"
	"coldsign/intent"
	"coldsign/policy"
	"coldsign/qr"
	"coldsign/signer"
	"coldsign/tx"

	"github.com/ethereum/go-ethereum/common"
)

func main() {
	qrFlag := flag.Bool("qr", false, "print signed raw tx as terminal QR (to stderr)")
	intentStdin := flag.Bool("intent-stdin", false, "read intent from stdin (JSON or coldintent:v1:<base64url>)")
	flag.Parse()

	var rawInput []byte
	var err error

	if *intentStdin {
		fmt.Fprintln(os.Stderr, "READY: waiting for intent on stdin (JSON or coldintent:v1:...)")
		fmt.Fprintln(os.Stderr, "Tip: zbarcam --raw | coldsign --intent-stdin ...")

		// Read everything from stdin (ephemeral)
		reader := bufio.NewReader(os.Stdin)
		line, readErr := reader.ReadString('\n')
		if readErr != nil && readErr != io.EOF {
			fmt.Println("stdin read error:", readErr)
			os.Exit(1)
		}
		rawInput = []byte(strings.TrimSpace(line))
		if len(rawInput) == 0 {
			fmt.Println("stdin error: no intent provided")
			os.Exit(1)
		}

	} else {
		if flag.NArg() != 1 {
			fmt.Println("usage: coldsign [-qr] [--intent-stdin] <intent.json>")
			fmt.Println("  - for stdin mode:  zbarcam --raw | coldsign --intent-stdin")
			os.Exit(1)
		}
		path := flag.Arg(0)
		rawInput, err = os.ReadFile(path)
		if err != nil {
			fmt.Println("read error:", err)
			os.Exit(1)
		}
	}

	// Decode envelope if needed (or pass through raw JSON)
	decodedJSON, err := intent.DecodeEnvelopeOrJSON(string(rawInput))
	if err != nil {
		fmt.Println("intent decode error:", err)
		os.Exit(1)
	}

	// Parse + validate the intent JSON
	in, err := intent.ParseEthSend(decodedJSON)
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

	privKey, derivedAddr, err := hd.DeriveEthKey(mn, pass, in.From.Index)
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

	unsignedTx, err := tx.BuildUnsignedEthSendTx(in)
	if err != nil {
		fmt.Println("tx build error:", err)
		os.Exit(1)
	}

	fmt.Println("Unsigned tx type:", unsignedTx.Type())
	fmt.Println("Unsigned tx hash (pre-sign):", unsignedTx.Hash().Hex())

	signed, err := signer.SignEIP1559Tx(unsignedTx, in.ChainID, privKey)
	if err != nil {
		fmt.Println("sign error:", err)
		os.Exit(1)
	}

	fmt.Println("Signed tx hash:", signed.TxHash)
	fmt.Println("Signed raw tx hex:", signed.RawTxHex)

	if *qrFlag {
		fmt.Fprintln(os.Stderr, "\n--- SIGNED RAW TX QR (scan on online machine) ---")
		qr.PrintToTerminal(signed.RawTxHex)
	}
}
