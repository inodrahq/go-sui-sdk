// Package ed25519 provides Ed25519 keypair implementation for Sui.
package ed25519

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"

	"github.com/inodrahq/go-sui-sdk/crypto"
)

// Keypair implements the crypto.Keypair interface for Ed25519.
type Keypair struct {
	privKey ed25519.PrivateKey
}

// New generates a new random Ed25519 keypair.
func New() (*Keypair, error) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("ed25519: generate key: %w", err)
	}
	return &Keypair{privKey: priv}, nil
}

// FromSeed creates an Ed25519 keypair from a 32-byte seed.
func FromSeed(seed []byte) (*Keypair, error) {
	if len(seed) != ed25519.SeedSize {
		return nil, fmt.Errorf("ed25519: seed must be %d bytes, got %d", ed25519.SeedSize, len(seed))
	}
	priv := ed25519.NewKeyFromSeed(seed)
	return &Keypair{privKey: priv}, nil
}

// FromPrivateKey creates an Ed25519 keypair from a 64-byte private key.
func FromPrivateKey(privKey []byte) (*Keypair, error) {
	if len(privKey) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("ed25519: private key must be %d bytes, got %d", ed25519.PrivateKeySize, len(privKey))
	}
	return &Keypair{privKey: ed25519.PrivateKey(privKey)}, nil
}

func (kp *Keypair) Scheme() crypto.SignatureScheme {
	return crypto.Ed25519Scheme
}

func (kp *Keypair) PublicKey() crypto.PublicKey {
	pub := kp.privKey.Public().(ed25519.PublicKey)
	return NewPublicKey(pub)
}

// Sign signs the message (expected to be pre-hashed Blake2b digest).
func (kp *Keypair) Sign(msg []byte) ([]byte, error) {
	return ed25519.Sign(kp.privKey, msg), nil
}

// Seed returns the 32-byte seed of the private key.
func (kp *Keypair) Seed() []byte {
	return kp.privKey.Seed()
}
