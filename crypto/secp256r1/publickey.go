package secp256r1

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"math/big"

	"github.com/inodrahq/go-sui-sdk/crypto"
)

// PublicKey implements crypto.PublicKey for Secp256r1 (33-byte compressed).
type PublicKey struct {
	key *ecdsa.PublicKey
}

// NewPublicKey creates a Secp256r1 public key.
func NewPublicKey(key *ecdsa.PublicKey) *PublicKey {
	return &PublicKey{key: key}
}

// NewPublicKeyFromBytes creates a Secp256r1 public key from 33-byte compressed bytes.
func NewPublicKeyFromBytes(b []byte) (*PublicKey, error) {
	x, y := elliptic.UnmarshalCompressed(elliptic.P256(), b)
	if x == nil {
		return nil, crypto.ErrInvalidPublicKey
	}
	return &PublicKey{key: &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}}, nil
}

func (pk *PublicKey) Scheme() crypto.SignatureScheme {
	return crypto.Secp256r1Scheme
}

// Bytes returns the 33-byte compressed public key.
func (pk *PublicKey) Bytes() []byte {
	return elliptic.MarshalCompressed(pk.key.Curve, pk.key.X, pk.key.Y)
}

func (pk *PublicKey) Flag() byte {
	return byte(crypto.Secp256r1Scheme)
}

func (pk *PublicKey) SuiAddress() string {
	return crypto.DeriveAddress(pk.Flag(), pk.Bytes())
}

// Verify verifies a 64-byte compact signature against a message hash.
func (pk *PublicKey) Verify(hash, sig []byte) bool {
	if len(sig) != 64 {
		return false
	}
	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:])
	return ecdsa.Verify(pk.key, hash, r, s)
}
