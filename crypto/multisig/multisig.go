// Package multisig provides MultiSig support for Sui.
package multisig

import (
	"encoding/base64"
	"fmt"
	"sort"

	"golang.org/x/crypto/blake2b"

	"github.com/inodrahq/go-bcs"
	"github.com/inodrahq/go-sui-sdk/crypto"
)

// MultiSigPublicKey represents a multisig public key with threshold and member weights.
type MultiSigPublicKey struct {
	PubKeys   []MemberPublicKey
	Threshold uint16
}

// MemberPublicKey is a public key with its weight in the multisig.
type MemberPublicKey struct {
	PubKey crypto.PublicKey
	Weight uint8
}

// SuiAddress derives the multisig address: Blake2b256(MultiSigFlag || threshold_le16 || pk1_bcs || w1 || pk2_bcs || w2 || ...).
func (ms *MultiSigPublicKey) SuiAddress() string {
	e := bcs.NewEncoder()

	// Flag byte
	e.WriteByte(byte(crypto.MultiSigScheme))

	// Threshold as u16 LE
	e.WriteU16(ms.Threshold)

	// Each member: scheme_flag(1) || raw_pk_bytes(variable) || weight(1)
	for _, m := range ms.PubKeys {
		pk := m.PubKey
		e.WriteByte(pk.Flag())
		e.WriteBytes(pk.Bytes())
		e.WriteByte(m.Weight)
	}

	hash := blake2b.Sum256(e.Bytes())
	return "0x" + fmt.Sprintf("%064x", hash)
}

// Bitmap computes the bitmap for the given signer indices (0-based).
// The bitmap is a u16 where bit i is set if signer i participated.
func Bitmap(indices []int) uint16 {
	var bitmap uint16
	for _, i := range indices {
		bitmap |= 1 << uint(i)
	}
	return bitmap
}

// CombineSignatures creates a multisig serialized signature from individual Sui signatures.
// signerIndices maps each signature to its position in the PubKeys array.
func (ms *MultiSigPublicKey) CombineSignatures(signatures []string, signerIndices []int) (string, error) {
	if len(signatures) != len(signerIndices) {
		return "", fmt.Errorf("multisig: signatures and indices length mismatch")
	}

	// Validate threshold
	var totalWeight uint16
	for _, idx := range signerIndices {
		if idx < 0 || idx >= len(ms.PubKeys) {
			return "", fmt.Errorf("multisig: signer index %d out of range", idx)
		}
		totalWeight += uint16(ms.PubKeys[idx].Weight)
	}
	if totalWeight < ms.Threshold {
		return "", fmt.Errorf("multisig: total weight %d below threshold %d", totalWeight, ms.Threshold)
	}

	// Sort by signer index for deterministic encoding
	type sigEntry struct {
		sig   string
		index int
	}
	entries := make([]sigEntry, len(signatures))
	for i := range signatures {
		entries[i] = sigEntry{sig: signatures[i], index: signerIndices[i]}
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].index < entries[j].index
	})

	bitmap := Bitmap(signerIndices)

	// Build multisig BCS structure
	e := bcs.NewEncoder()

	// MultiSig flag
	e.WriteByte(byte(crypto.MultiSigScheme))

	// Vec<CompressedSignature> — ULEB128 length + each signature
	e.WriteULEB128(uint32(len(entries)))
	for _, entry := range entries {
		sigBytes, err := base64.StdEncoding.DecodeString(entry.sig)
		if err != nil {
			return "", fmt.Errorf("multisig: decode signature: %w", err)
		}
		if len(sigBytes) < 1 {
			return "", fmt.Errorf("multisig: empty signature")
		}

		// CompressedSignature is an enum: Ed25519([u8;64])=0, Secp256k1([u8;64])=1, Secp256r1([u8;64])=2
		scheme := sigBytes[0]
		var rawSig []byte
		switch crypto.SignatureScheme(scheme) {
		case crypto.Ed25519Scheme:
			e.WriteULEB128(0) // Ed25519 variant
			rawSig = sigBytes[1:65]
		case crypto.Secp256k1Scheme:
			e.WriteULEB128(1)
			rawSig = sigBytes[1:65]
		case crypto.Secp256r1Scheme:
			e.WriteULEB128(2)
			rawSig = sigBytes[1:65]
		default:
			return "", fmt.Errorf("multisig: unsupported scheme %d", scheme)
		}
		// Fixed-size array (no length prefix)
		e.WriteBytes(rawSig)
	}

	// bitmap as u16 LE
	e.WriteU16(bitmap)

	// MultiSigPublicKey: Vec<(PublicKey, u8)> + threshold
	e.WriteULEB128(uint32(len(ms.PubKeys)))
	for _, m := range ms.PubKeys {
		pk := m.PubKey
		// PublicKey enum: Ed25519([u8;32])=0, Secp256k1([u8;33])=1, Secp256r1([u8;33])=2
		switch pk.Scheme() {
		case crypto.Ed25519Scheme:
			e.WriteULEB128(0)
			e.WriteBytes(pk.Bytes()) // fixed 32 bytes
		case crypto.Secp256k1Scheme:
			e.WriteULEB128(1)
			e.WriteBytes(pk.Bytes()) // fixed 33 bytes
		case crypto.Secp256r1Scheme:
			e.WriteULEB128(2)
			e.WriteBytes(pk.Bytes()) // fixed 33 bytes
		}
		e.WriteByte(m.Weight)
	}
	e.WriteU16(ms.Threshold)

	return base64.StdEncoding.EncodeToString(e.Bytes()), nil
}
