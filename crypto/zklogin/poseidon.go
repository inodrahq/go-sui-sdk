// Package zklogin provides zkLogin address derivation and authenticator assembly for Sui.
package zklogin

import (
	"fmt"
	"math/big"

	"github.com/iden3/go-iden3-crypto/poseidon"
)

// PoseidonHash computes a Poseidon hash over BN254 field elements.
// Inputs are decimal strings representing field elements.
func PoseidonHash(inputs []string) (string, error) {
	if len(inputs) < 1 || len(inputs) > 16 {
		return "", fmt.Errorf("zklogin: poseidon supports 1-16 inputs, got %d", len(inputs))
	}

	bigInputs := make([]*big.Int, len(inputs))
	for i, s := range inputs {
		val, ok := new(big.Int).SetString(s, 10)
		if !ok {
			return "", fmt.Errorf("zklogin: invalid decimal input %q", s)
		}
		bigInputs[i] = val
	}

	result, err := poseidon.Hash(bigInputs)
	if err != nil {
		return "", fmt.Errorf("zklogin: poseidon hash: %w", err)
	}

	return result.String(), nil
}
