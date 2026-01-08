package policy

import (
	"fmt"
	"math/big"

	"coldsign/intent"
)

type Policy struct {
	AllowedChainIDs map[uint64]bool

	MaxFeePerGasWei        *big.Int
	MaxPriorityFeePerGasWei *big.Int
	MaxValueWei            *big.Int
}

func Default() *Policy {
	return &Policy{
		AllowedChainIDs: map[uint64]bool{
			1: true, // mainnet only (Phase 0)
		},

		// Conservative, adjustable later
		MaxFeePerGasWei:        big.NewInt(200_000_000_000), // 200 gwei
		MaxPriorityFeePerGasWei: big.NewInt(10_000_000_000), // 10 gwei
		MaxValueWei:            big.NewInt(0).Mul(big.NewInt(1000), big.NewInt(1e18)), // 1000 ETH
	}
}

func parseWei(s string) (*big.Int, error) {
	x, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return nil, fmt.Errorf("invalid wei value: %q", s)
	}
	return x, nil
}

func (p *Policy) Enforce(in *intent.EthSendIntent) error {
	if !p.AllowedChainIDs[in.ChainID] {
		return fmt.Errorf("chainId %d not allowed by policy", in.ChainID)
	}

	maxFee, _ := parseWei(in.MaxFeePerGasWei)
	if maxFee.Cmp(p.MaxFeePerGasWei) > 0 {
		return fmt.Errorf("maxFeePerGas exceeds policy limit")
	}

	maxPrio, _ := parseWei(in.MaxPriorityFeePerGasWei)
	if maxPrio.Cmp(p.MaxPriorityFeePerGasWei) > 0 {
		return fmt.Errorf("maxPriorityFeePerGas exceeds policy limit")
	}

	value, _ := parseWei(in.ValueWei)
	if value.Cmp(p.MaxValueWei) > 0 {
		return fmt.Errorf("value exceeds policy limit")
	}

	return nil
}
