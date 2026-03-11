package secp256r1

import (
	"crypto/elliptic"
	"math/big"
)

// normalizeLowS ensures S is in the lower half of the curve order.
// This is required by Sui for secp256r1 signatures.
func normalizeLowS(s *big.Int, curve elliptic.Curve) *big.Int {
	halfOrder := new(big.Int).Rsh(curve.Params().N, 1)
	if s.Cmp(halfOrder) > 0 {
		s = new(big.Int).Sub(curve.Params().N, s)
	}
	return s
}
