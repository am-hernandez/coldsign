package intent

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type FromRef struct {
	Type  string `json:"type"` // "bip44_index"
	Index uint32 `json:"index"`
}

type EthSendIntent struct {
	V                       int     `json:"v"`
	Kind                    string  `json:"kind"` // must be "ETH_SEND"
	ChainID                 uint64  `json:"chainId"`
	From                    FromRef `json:"from"`
	To                      string  `json:"to"`
	ValueWei                string  `json:"valueWei"`
	Nonce                   uint64  `json:"nonce"`
	MaxFeePerGasWei         string  `json:"maxFeePerGasWei"`
	MaxPriorityFeePerGasWei string  `json:"maxPriorityFeePerGasWei"`
}

func ParseEthSend(b []byte) (*EthSendIntent, error) {
	var in EthSendIntent
	if err := json.Unmarshal(b, &in); err != nil {
		return nil, err
	}

	if in.V != 1 {
		return nil, fmt.Errorf("unsupported intent version: %d", in.V)
	}
	if in.Kind != "ETH_SEND" {
		return nil, fmt.Errorf("unsupported intent kind: %s", in.Kind)
	}
	if in.From.Type != "bip44_index" {
		return nil, fmt.Errorf("unsupported from.type: %s", in.From.Type)
	}

	if err := in.Validate(); err != nil {
		return nil, err
	}

	return &in, nil
}

func parseUintDecimal(s string) (*big.Int, error) {
	x, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return nil, fmt.Errorf("invalid decimal integer: %q", s)
	}
	if x.Sign() < 0 {
		return nil, fmt.Errorf("negative value not allowed: %q", s)
	}
	return x, nil
}

func (in *EthSendIntent) Validate() error {
	// Address checks
	if !common.IsHexAddress(in.To) {
		return fmt.Errorf("invalid to address: %s", in.To)
	}
	to := common.HexToAddress(in.To)
	if to == (common.Address{}) {
		return fmt.Errorf("to address must not be zero address")
	}

	// Numeric checks
	if _, err := parseUintDecimal(in.ValueWei); err != nil {
		return fmt.Errorf("valueWei: %w", err)
	}
	if _, err := parseUintDecimal(in.MaxFeePerGasWei); err != nil {
		return fmt.Errorf("maxFeePerGasWei: %w", err)
	}
	if _, err := parseUintDecimal(in.MaxPriorityFeePerGasWei); err != nil {
		return fmt.Errorf("maxPriorityFeePerGasWei: %w", err)
	}

	return nil
}
