package base58

import (
	"fmt"
	"math/big"
)

const alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

var bigZero = big.NewInt(0)
var big58 = big.NewInt(58)

// Decode decodes a base58-encoded string.
func Decode(s string) ([]byte, error) {
	if len(s) == 0 {
		return []byte{}, nil
	}

	// Count leading '1's (base58 zero)
	var leadingZeros int
	for _, c := range s {
		if c != '1' {
			break
		}
		leadingZeros++
	}

	n := new(big.Int)
	for _, c := range s {
		idx := -1
		for i, a := range alphabet {
			if a == c {
				idx = i
				break
			}
		}
		if idx < 0 {
			return nil, fmt.Errorf("base58: invalid character %q", c)
		}
		n.Mul(n, big58)
		n.Add(n, big.NewInt(int64(idx)))
	}

	b := n.Bytes()
	result := make([]byte, leadingZeros+len(b))
	copy(result[leadingZeros:], b)
	return result, nil
}
