// Package secp256r1 provides Secp256r1 (P-256) keypair implementation for Sui.
package secp256r1

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/inodrahq/go-sui-sdk/crypto"
)

// Keypair implements crypto.Keypair for Secp256r1 (P-256).
type Keypair struct {
	privKey *ecdsa.PrivateKey
}

// New generates a new random Secp256r1 keypair.
func New() (*Keypair, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("secp256r1: generate key: %w", err)
	}
	return &Keypair{privKey: priv}, nil
}

// FromPrivateKeyBytes creates a Secp256r1 keypair from 32-byte raw private key.
func FromPrivateKeyBytes(key []byte) (*Keypair, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("secp256r1: private key must be 32 bytes, got %d", len(key))
	}
	curve := elliptic.P256()
	d := new(big.Int).SetBytes(key)
	x, y := curve.ScalarBaseMult(d.Bytes())
	priv := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: curve,
			X:     x,
			Y:     y,
		},
		D: d,
	}
	return &Keypair{privKey: priv}, nil
}

func (kp *Keypair) Scheme() crypto.SignatureScheme {
	return crypto.Secp256r1Scheme
}

func (kp *Keypair) PublicKey() crypto.PublicKey {
	return NewPublicKey(&kp.privKey.PublicKey)
}

// Sign signs the message hash (expected to be 32-byte Blake2b hash).
// Returns a 64-byte compact signature with low-S normalization.
func (kp *Keypair) Sign(hash []byte) ([]byte, error) {
	r, s, err := ecdsa.Sign(rand.Reader, kp.privKey, hash)
	if err != nil {
		return nil, fmt.Errorf("secp256r1: sign: %w", err)
	}

	// Normalize to low-S
	s = normalizeLowS(s, kp.privKey.Curve)

	// Convert to 64-byte compact format [R(32) || S(32)]
	sig := make([]byte, 64)
	rBytes := r.Bytes()
	sBytes := s.Bytes()
	copy(sig[32-len(rBytes):32], rBytes)
	copy(sig[64-len(sBytes):64], sBytes)
	return sig, nil
}

// Seed returns the raw 32-byte private key.
func (kp *Keypair) Seed() []byte {
	b := kp.privKey.D.Bytes()
	out := make([]byte, 32)
	copy(out[32-len(b):], b)
	return out
}

// PrivateKeyBytes returns the raw 32-byte private key.
func (kp *Keypair) PrivateKeyBytes() []byte {
	return kp.Seed()
}
