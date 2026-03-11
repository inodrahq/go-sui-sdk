// Package secp256k1 provides Secp256k1 keypair implementation for Sui.
package secp256k1

import (
	"fmt"

	dcrdsecp "github.com/decred/dcrd/dcrec/secp256k1/v4"
	dcrdecdsa "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"

	"github.com/inodrahq/go-sui-sdk/crypto"
)

// Keypair implements crypto.Keypair for Secp256k1.
type Keypair struct {
	privKey *dcrdsecp.PrivateKey
}

// New generates a new random Secp256k1 keypair.
func New() (*Keypair, error) {
	priv, err := dcrdsecp.GeneratePrivateKey()
	if err != nil {
		return nil, fmt.Errorf("secp256k1: generate key: %w", err)
	}
	return &Keypair{privKey: priv}, nil
}

// FromPrivateKeyBytes creates a Secp256k1 keypair from 32-byte raw private key.
func FromPrivateKeyBytes(key []byte) (*Keypair, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("secp256k1: private key must be 32 bytes, got %d", len(key))
	}
	priv := dcrdsecp.PrivKeyFromBytes(key)
	return &Keypair{privKey: priv}, nil
}

func (kp *Keypair) Scheme() crypto.SignatureScheme {
	return crypto.Secp256k1Scheme
}

func (kp *Keypair) PublicKey() crypto.PublicKey {
	return NewPublicKey(kp.privKey.PubKey())
}

// Sign signs the message hash (expected to be 32-byte Blake2b hash).
// Returns a 64-byte compact signature with low-S normalization (handled by dcrd).
func (kp *Keypair) Sign(hash []byte) ([]byte, error) {
	// dcrd's SignCompact returns [v || r || s] (65 bytes) — we strip v
	sig := dcrdecdsa.SignCompact(kp.privKey, hash, false)
	// sig[0] is recovery byte, sig[1:33] is R, sig[33:65] is S
	return sig[1:65], nil
}

// Seed returns the raw 32-byte private key.
func (kp *Keypair) Seed() []byte {
	return kp.privKey.Serialize()
}

// PrivateKeyBytes returns the raw 32-byte private key.
func (kp *Keypair) PrivateKeyBytes() []byte {
	return kp.privKey.Serialize()
}
