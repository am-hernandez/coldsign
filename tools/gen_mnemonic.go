package main

import (
	"fmt"
	"os"
	"time"

	"github.com/tyler-smith/go-bip39"
)

func main() {
	if len(os.Args) != 2 || os.Args[1] != "--print" {
		fmt.Fprintln(os.Stderr, "Usage: gen_mnemonic --print")
		os.Exit(1)
	}

	entropy, err := bip39.NewEntropy(256) // 24 words
	if err != nil {
		panic(err)
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		panic(err)
	}

	fmt.Fprintln(os.Stderr, "\033[31m!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	fmt.Fprintln(os.Stderr, "!!!  SECURITY WARNING – READ CAREFULLY      !!!")
	fmt.Fprintln(os.Stderr, "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	fmt.Fprintln(os.Stderr, "YOU ARE ABOUT TO GENERATE A WALLET MNEMONIC.")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "THIS WILL BE PRINTED TO STDOUT.")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "IF YOU ARE NOT ALONE, YOUR TERMINAL MAY BE")
	fmt.Fprintln(os.Stderr, "LOGGED, OR THIS SCREEN MAY BE RECORDED:")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "EXIT NOW AND REGENERATE IN A PRIVATE SESSION.")
	fmt.Fprintln(os.Stderr, "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\033[0m")

	time.Sleep(1500 * time.Millisecond)

	fmt.Println(mnemonic)
}
