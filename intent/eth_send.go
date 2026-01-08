package intent

import (
	"encoding/json"
	"fmt"
)

type FromRef struct {
	Type  string `json:"type"`  // "bip44_index"
	Index uint32 `json:"index"`
}

type EthSendIntent struct {
	V                      int     `json:"v"`
	Kind                   string  `json:"kind"` // must be "ETH_SEND"
	ChainID                uint64  `json:"chainId"`
	From                   FromRef `json:"from"`
	To                     string  `json:"to"`
	ValueWei               string  `json:"valueWei"`
	Nonce                  uint64  `json:"nonce"`
	MaxFeePerGasWei        string  `json:"maxFeePerGasWei"`
	MaxPriorityFeePerGasWei string `json:"maxPriorityFeePerGasWei"`
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

	return &in, nil
}
