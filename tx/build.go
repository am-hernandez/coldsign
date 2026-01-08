package tx

import (
	"fmt"
	"math/big"

	"coldsign/intent"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func parseWei(s string) (*big.Int, error) {
	x, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return nil, fmt.Errorf("invalid wei value: %q", s)
	}
	if x.Sign() < 0 {
		return nil, fmt.Errorf("negative wei not allowed: %q", s)
	}
	return x, nil
}

// BuildUnsignedEthSendTx builds a type-2 (EIP-1559) ETH transfer tx.
// Phase 0 invariants:
//   - gas = 21000
//   - data = empty
func BuildUnsignedEthSendTx(in *intent.EthSendIntent) (*types.Transaction, error) {
	to := common.HexToAddress(in.To)

	valueWei, err := parseWei(in.ValueWei)
	if err != nil {
		return nil, err
	}
	maxFeeWei, err := parseWei(in.MaxFeePerGasWei)
	if err != nil {
		return nil, err
	}
	maxPrioWei, err := parseWei(in.MaxPriorityFeePerGasWei)
	if err != nil {
		return nil, err
	}

	txData := &types.DynamicFeeTx{
		ChainID:   new(big.Int).SetUint64(in.ChainID),
		Nonce:     in.Nonce,
		To:        &to,
		Value:     valueWei,
		Gas:       21000,
		GasFeeCap: maxFeeWei,
		GasTipCap: maxPrioWei,
		Data:      []byte{},
	}

	return types.NewTx(txData), nil
}
