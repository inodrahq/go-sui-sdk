package ed25519

import (
	stded25519 "crypto/ed25519"

	"github.com/inodrahq/go-sui-sdk/crypto"
)

// PublicKey implements crypto.PublicKey for Ed25519.
type PublicKey struct {
	key stded25519.PublicKey
}

// NewPublicKey creates an Ed25519 public key from raw bytes.
func NewPublicKey(key stded25519.PublicKey) *PublicKey {
	return &PublicKey{key: key}
}

func (pk *PublicKey) Scheme() crypto.SignatureScheme {
	return crypto.Ed25519Scheme
}

func (pk *PublicKey) Bytes() []byte {
	return pk.key
}

func (pk *PublicKey) Flag() byte {
	return byte(crypto.Ed25519Scheme)
}

func (pk *PublicKey) SuiAddress() string {
	return crypto.DeriveAddress(pk.Flag(), pk.key)
}

// Verify verifies an Ed25519 signature against a message.
func (pk *PublicKey) Verify(msg, sig []byte) bool {
	if len(sig) != stded25519.SignatureSize {
		return false
	}
	return stded25519.Verify(pk.key, msg, sig)
}
