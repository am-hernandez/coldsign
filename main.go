package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math/big"
	"os"
	"strings"

	"coldsign/hd"
	"coldsign/helpers"
	"coldsign/intent"
	"coldsign/logo"
	"coldsign/policy"
	"coldsign/qr"
	"coldsign/signer"
	"coldsign/tx"

	"github.com/ethereum/go-ethereum/common"
)

var Version = "dev"

var weiPerETH = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "help", "-h", "--help":
		printHelp()
	case "version", "-v", "--version":
		printVersion()
	case "sign":
		os.Exit(runSign(os.Args[2:]))
	case "addr":
		os.Exit(runAddr(os.Args[2:]))
	default:
		// Backward compatibility: coldsign <intent.json>
		if helpers.FileExists(cmd) {
			os.Exit(runSign(os.Args[1:]))
		}
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	logo.Print()
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  coldsign sign [flags] <intent.json>")
	fmt.Fprintln(os.Stderr, "  coldsign addr --index N [--qr]")
	fmt.Fprintln(os.Stderr, "  coldsign help")
	fmt.Fprintln(os.Stderr, "  coldsign version")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  sign     Review and sign transaction intents")
	fmt.Fprintln(os.Stderr, "  addr     Derive and display Ethereum addresses")
	fmt.Fprintln(os.Stderr, "  help     Show this help message")
	fmt.Fprintln(os.Stderr, "  version  Show version information")
}

func printVersion() {
	fmt.Println("coldsign", Version)
}

func runSign(args []string) int {
	fs := flag.NewFlagSet("sign", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	qrFlag := fs.Bool("qr", false, "print signed raw tx as terminal QR (to stderr)")
	intentStdin := fs.Bool("intent-stdin", false, "read intent from stdin (JSON or coldintent:v1:<base64url>)")
	signFlag := fs.Bool("sign", false, "authorize signing (otherwise only review)")
	yesFlag := fs.Bool("yes", false, "skip interactive confirmation (DANGEROUS)")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	var rawInput []byte
	var err error

	if *intentStdin {
		fmt.Fprintln(os.Stderr, "READY: waiting for intent on stdin (JSON or coldintent:v1:...)")
		fmt.Fprintln(os.Stderr, "Tip: zbarcam --raw | coldsign sign --intent-stdin ...")

		reader := bufio.NewReader(os.Stdin)
		line, readErr := reader.ReadString('\n')
		if readErr != nil && readErr != io.EOF {
			fmt.Fprintln(os.Stderr, "stdin read error:", readErr)
			return 1
		}
		rawInput = []byte(strings.TrimSpace(line))
		if len(rawInput) == 0 {
			fmt.Fprintln(os.Stderr, "stdin error: no intent provided")
			return 1
		}
	} else {
		if fs.NArg() != 1 {
			fmt.Fprintln(os.Stderr, "usage: coldsign sign [flags] <intent.json>")
			return 2
		}
		rawInput, err = os.ReadFile(fs.Arg(0))
		if err != nil {
			fmt.Fprintln(os.Stderr, "read error:", err)
			return 1
		}
	}

	decodedJSON, err := intent.DecodeEnvelopeOrJSON(string(rawInput))
	if err != nil {
		fmt.Fprintln(os.Stderr, "intent decode error:", err)
		return 1
	}

	in, err := intent.ParseEthSend(decodedJSON)
	if err != nil {
		fmt.Fprintln(os.Stderr, "intent error:", err)
		return 1
	}

	// Validate addresses before proceeding
	if !common.IsHexAddress(in.To) {
		fmt.Fprintln(os.Stderr, "intent error: invalid to address")
		return 1
	}
	if !common.IsHexAddress(in.FromAddress) {
		fmt.Fprintln(os.Stderr, "intent error: invalid fromAddress")
		return 1
	}

	fmt.Println("")
	fmt.Println(helpers.Separator("SIGNING REVIEW (ETH_SEND)"))
	fmt.Printf("Chain:   %d\n", in.ChainID)
	fmt.Printf("From:    %s\n", in.FromAddress)
	fmt.Printf("To:      %s\n", in.To)
	fmt.Printf("Nonce:   %d\n", in.Nonce)

	amtEth, err := helpers.FormatETH(in.ValueWei)
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid valueWei:", err)
		return 1
	}
	fmt.Printf("Amount:  %s ETH  (%s wei)\n", amtEth, in.ValueWei)

	maxGwei, err := helpers.FormatGwei(in.MaxFeePerGasWei)
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid maxFeePerGasWei:", err)
		return 1
	}
	tipGwei, err := helpers.FormatGwei(in.MaxPriorityFeePerGasWei)
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid maxPriorityFeePerGasWei:", err)
		return 1
	}
	fmt.Printf("Fees:    max=%s gwei, tip=%s gwei\n", maxGwei, tipGwei)

	// Worst-case: 21000 * maxFeePerGasWei
	mfWei, ok := new(big.Int).SetString(in.MaxFeePerGasWei, 10)
	if !ok {
		fmt.Fprintln(os.Stderr, "invalid maxFeePerGasWei:", in.MaxFeePerGasWei)
		return 1
	}

	worstWei := new(big.Int).Mul(big.NewInt(21000), mfWei)
	worstEth, err := helpers.FormatETH6(worstWei.String())
	if err != nil {
		fmt.Fprintln(os.Stderr, "fee cap format error:", err)
		return 1
	}

	fmt.Printf("Fee cap: ~%s ETH worst-case\n", worstEth)

	fmt.Println(helpers.Separator(""))

	if err := policy.Default().Enforce(in); err != nil {
		fmt.Fprintln(os.Stderr, "policy violation:", err)
		return 1
	}

	fmt.Println("Policy check: OK")

	if !*signFlag {
		fmt.Println("NOT SIGNED: pass --sign to authorize signing")
		return 2
	}

	if !*yesFlag {
		tty, err := os.Open("/dev/tty")
		if err != nil {
			fmt.Fprintln(os.Stderr, "no TTY available; re-run with --yes")
			return 2
		}
		defer tty.Close()

		to := common.HexToAddress(in.To).Hex()
		// Generate confirmation code once and reuse for display and validation
		first := strings.ToLower(to[2:6])
		last := strings.ToLower(to[len(to)-4:])
		code := fmt.Sprintf("%s %s", first, last)

		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, helpers.Separator("CONFIRM SIGNING"))
		fmt.Fprintln(os.Stderr, "Destination address:")
		fmt.Fprintln(os.Stderr, to)
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Re-type the destination address fragment exactly as shown:")
		fmt.Fprintln(os.Stderr, code)
		fmt.Fprint(os.Stderr, "> ")

		resp, _ := bufio.NewReader(tty).ReadString('\n')
		got := strings.Fields(strings.ToLower(resp))
		want := strings.Fields(code) // ["1111","1111"]
		if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
			fmt.Fprintln(os.Stderr, "Canceled.")
			return 0
		}
	}

	mnemonic, err := helpers.ReadHiddenLineFromTTY(
		"ENTER MNEMONIC (space-separated BIP-39 words; hidden)",
		false,
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "mnemonic error:", err)
		return 1
	}
	defer helpers.ZeroString(&mnemonic)

	passphrase, err := helpers.ReadHiddenLineFromTTY(
		"ENTER PASSPHRASE (optional; hidden)",
		true,
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "passphrase error:", err)
		return 1
	}
	defer helpers.ZeroString(&passphrase)

	privKey, addr, err := hd.DeriveEthKey(mnemonic, passphrase, in.From.Index)
	if err != nil {
		fmt.Fprintln(os.Stderr, "hd derive error:", err)
		return 1
	}

	if addr.Hex() != common.HexToAddress(in.FromAddress).Hex() {
		fmt.Fprintln(os.Stderr, "fromAddress mismatch")
		return 1
	} else {
		fmt.Println("From address verified:", addr.Hex())
	}

	unsignedTx, err := tx.BuildUnsignedEthSendTx(in)
	if err != nil {
		fmt.Fprintln(os.Stderr, "tx build error:", err)
		return 1
	}

	signed, err := signer.SignEIP1559Tx(unsignedTx, in.ChainID, privKey)
	if err != nil {
		fmt.Fprintln(os.Stderr, "sign error:", err)
		return 1
	}

	fmt.Println(helpers.Separator(""))
	fmt.Println("Signed tx hash:", signed.TxHash)
	fmt.Println("Signed raw tx hex:", signed.RawTxHex)

	if *qrFlag {
		fmt.Fprintln(os.Stderr, "\n--- SIGNED RAW TX QR ---")
		qr.PrintToTerminal(signed.RawTxHex)
	}

	fmt.Println("DONE: signed transaction ready for broadcast")
	return 0
}

func runAddr(args []string) int {
	fs := flag.NewFlagSet("addr", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	index := fs.Int("index", -1, "BIP-44 address index")
	qrFlag := fs.Bool("qr", false, "print address as terminal QR")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *index < 0 {
		fmt.Fprintln(os.Stderr, "usage: coldsign addr --index N [--qr]")
		return 2
	}

	mnemonic, err := helpers.ReadHiddenLineFromTTY("ENTER MNEMONIC (hidden)", false)
	if err != nil {
		return 1
	}
	defer helpers.ZeroString(&mnemonic)

	passphrase, err := helpers.ReadHiddenLineFromTTY("ENTER PASSPHRASE (optional)", true)
	if err != nil {
		return 1
	}
	defer helpers.ZeroString(&passphrase)

	_, addr, err := hd.DeriveEthKey(mnemonic, passphrase, uint32(*index))
	if err != nil {
		fmt.Fprintln(os.Stderr, "derive error:", err)
		return 1
	}

	fmt.Println(addr.Hex())
	if *qrFlag {
		fmt.Fprintln(os.Stderr, "\n--- ADDRESS QR ---")
		qr.PrintToTerminal(addr.Hex())
	}
	return 0
}
