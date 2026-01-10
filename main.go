package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math/big"
	"os"
	"strings"

	"coldsign/helpers"
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

	signFlag := flag.Bool("sign", false, "authorize signing (otherwise only review)")
	yesFlag := flag.Bool("yes", false, "skip interactive confirmation (DANGEROUS)")
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

	fmt.Println("")
	fmt.Println("========== SIGNING REVIEW (ETH_SEND) ==========")
	fmt.Printf("Chain:   %d\n", in.ChainID)
	fmt.Printf("From:    %s\n", in.FromAddress)
	fmt.Printf("To:      %s\n", in.To)
	fmt.Printf("Nonce:   %d\n", in.Nonce)

	// Amount display: ETH + wei
	valWei, ok := new(big.Int).SetString(in.ValueWei, 10)
	if !ok {
		fmt.Println("intent error: invalid valueWei")
		os.Exit(1)
	}
	valEth := new(big.Rat).SetFrac(valWei, big.NewInt(1e18))
	fmt.Printf("Amount:  %s ETH  (%s wei)\n", valEth.FloatString(18), in.ValueWei)

	// Fees: show gwei + wei
	mfWei, ok := new(big.Int).SetString(in.MaxFeePerGasWei, 10)
	if !ok {
		fmt.Println("intent error: invalid maxFeePerGasWei")
		os.Exit(1)
	}
	mpWei, ok := new(big.Int).SetString(in.MaxPriorityFeePerGasWei, 10)
	if !ok {
		fmt.Println("intent error: invalid maxPriorityFeePerGasWei")
		os.Exit(1)
	}

	gwei := big.NewInt(1e9)
	mfGwei := new(big.Rat).SetFrac(mfWei, gwei)
	mpGwei := new(big.Rat).SetFrac(mpWei, gwei)

	fmt.Printf("Fees:    max=%s gwei, tip=%s gwei\n",
		mfGwei.FloatString(2), mpGwei.FloatString(2))

	// Worst-case fee estimate for ETH send (gas=21000)
	gas := big.NewInt(21000)
	worstFeeWei := new(big.Int).Mul(gas, mfWei)
	worstFeeEth := new(big.Rat).SetFrac(worstFeeWei, big.NewInt(1e18))
	fmt.Printf("Fee cap: ~%s ETH worst-case (21000 * maxFeePerGas)\n", worstFeeEth.FloatString(6))

	fmt.Println("===============================================")

	p := policy.Default()
	if err := p.Enforce(in); err != nil {
		fmt.Println("policy violation:", err)
		os.Exit(1)
	}

	fmt.Println("Policy check: OK")

	mnemonic, err := helpers.ReadHiddenLineFromTTY(
		"ENTER MNEMONIC (space-separated BIP-39 words; input hidden)",
		false,
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "mnemonic error:", err)
		os.Exit(1)
	}

	passphrase, err := helpers.ReadHiddenLineFromTTY(
		"ENTER PASSPHRASE (optional, press ENTER to skip; input hidden)",
		true,
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "passphrase error:", err)
		os.Exit(1)
	}

	privKey, derivedAddr, err := hd.DeriveEthKey(
		mnemonic,
		passphrase,
		in.From.Index,
	)
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

	// Default: do not sign unless explicitly authorized
	if !*signFlag {
		fmt.Println("NOT SIGNED: pass --sign to authorize signing")
		os.Exit(2)
	}

	// If authorized to sign, require interactive confirmation unless --yes flag is provided.
	// IMPORTANT: read confirmation from /dev/tty (not stdin) so piping (zbarcam | coldsign) works.
	if !*yesFlag {
		tty, err := os.Open("/dev/tty")
		if err != nil {
			fmt.Fprintln(os.Stderr, "no TTY available for confirmation; re-run with --yes to sign without prompt")
			os.Exit(2)
		}
		defer tty.Close()

		to := common.HexToAddress(in.To).Hex()
		suffix := strings.ToLower(to[len(to)-6:])

		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "CONFIRM SIGNING")
		fmt.Fprintf(os.Stderr, "Type the last 6 hex chars of the TO address (...%s) and press ENTER to sign.\n", suffix)
		fmt.Fprint(os.Stderr, "> ")

		reader := bufio.NewReader(tty)
		resp, _ := reader.ReadString('\n')
		resp = strings.TrimSpace(strings.ToLower(resp))

		if resp != suffix {
			fmt.Fprintln(os.Stderr, "Canceled (suffix mismatch).")
			os.Exit(0)
		}
	}

	unsignedTx, err := tx.BuildUnsignedEthSendTx(in)
	if err != nil {
		fmt.Println("tx build error:", err)
		os.Exit(1)
	}

	fmt.Println("===============================================")
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

	// Best-effort memory hygiene: clear secrets AFTER signing
	helpers.ZeroString(&mnemonic)
	helpers.ZeroString(&passphrase)

	fmt.Println("DONE: signed transaction ready for broadcast")
}
