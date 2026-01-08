package hd

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip39"
)

// DeriveEthKey derives m/44'/60'/0'/0/index from a BIP-39 mnemonic.
// Passphrase is the optional BIP-39 passphrase ("25th word"), usually "".
func DeriveEthKey(mnemonic, passphrase string, index uint32) (*ecdsa.PrivateKey, common.Address, error) {
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, common.Address{}, fmt.Errorf("invalid mnemonic")
	}

	seed := bip39.NewSeed(mnemonic, passphrase)

	// hdkeychain expects network params; derivation math is the same.
	master, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("new master: %w", err)
	}

	// Path: m/44'/60'/0'/0/index
	// Hardened: add hdkeychain.HardenedKeyStart
	purpose, err := master.Derive(hdkeychain.HardenedKeyStart + 44)
	if err != nil {
		return nil, common.Address{}, err
	}
	coinType, err := purpose.Derive(hdkeychain.HardenedKeyStart + 60)
	if err != nil {
		return nil, common.Address{}, err
	}
	account, err := coinType.Derive(hdkeychain.HardenedKeyStart + 0)
	if err != nil {
		return nil, common.Address{}, err
	}
	change, err := account.Derive(0)
	if err != nil {
		return nil, common.Address{}, err
	}
	child, err := change.Derive(index)
	if err != nil {
		return nil, common.Address{}, err
	}

	ecPriv, err := child.ECPrivKey()
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("ec priv: %w", err)
	}

	// Convert to go-ethereum ecdsa key
	privBytes := ecPriv.Serialize()
	priv, err := crypto.ToECDSA(privBytes)
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("to ecdsa: %w", err)
	}

	addr := crypto.PubkeyToAddress(priv.PublicKey)
	return priv, addr, nil
}
