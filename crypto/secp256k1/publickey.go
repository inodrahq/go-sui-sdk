package secp256k1

import (
	dcrdsecp "github.com/decred/dcrd/dcrec/secp256k1/v4"
	dcrdecdsa "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"

	"github.com/inodrahq/go-sui-sdk/crypto"
)

// PublicKey implements crypto.PublicKey for Secp256k1 (33-byte compressed).
type PublicKey struct {
	key *dcrdsecp.PublicKey
}

// NewPublicKey creates a Secp256k1 public key from a dcrd public key.
func NewPublicKey(key *dcrdsecp.PublicKey) *PublicKey {
	return &PublicKey{key: key}
}

// NewPublicKeyFromBytes creates a Secp256k1 public key from 33-byte compressed bytes.
func NewPublicKeyFromBytes(b []byte) (*PublicKey, error) {
	key, err := dcrdsecp.ParsePubKey(b)
	if err != nil {
		return nil, err
	}
	return &PublicKey{key: key}, nil
}

func (pk *PublicKey) Scheme() crypto.SignatureScheme {
	return crypto.Secp256k1Scheme
}

func (pk *PublicKey) Bytes() []byte {
	return pk.key.SerializeCompressed()
}

func (pk *PublicKey) Flag() byte {
	return byte(crypto.Secp256k1Scheme)
}

func (pk *PublicKey) SuiAddress() string {
	return crypto.DeriveAddress(pk.Flag(), pk.Bytes())
}

// Verify verifies a 64-byte compact signature against a message hash.
func (pk *PublicKey) Verify(hash, sig []byte) bool {
	if len(sig) != 64 {
		return false
	}
	// Parse compact signature
	r := new(dcrdsecp.ModNScalar)
	r.SetByteSlice(sig[:32])
	s := new(dcrdsecp.ModNScalar)
	s.SetByteSlice(sig[32:])
	dcrdsig := dcrdecdsa.NewSignature(r, s)
	return dcrdsig.Verify(hash, pk.key)
}
