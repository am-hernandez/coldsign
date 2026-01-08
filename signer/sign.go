package signer

import (
	"crypto/ecdsa"
	"encoding/hex"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
)

type Result struct {
	RawTxHex string
	TxHash   string
}

func SignEIP1559Tx(tx *types.Transaction, chainID uint64, priv *ecdsa.PrivateKey) (*Result, error) {
	signer := types.LatestSignerForChainID(new(big.Int).SetUint64(chainID))

	signedTx, err := types.SignTx(tx, signer, priv)
	if err != nil {
		return nil, err
	}

	raw, err := signedTx.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return &Result{
		RawTxHex: "0x" + hex.EncodeToString(raw),
		TxHash:   signedTx.Hash().Hex(),
	}, nil
}
