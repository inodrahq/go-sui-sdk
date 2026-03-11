// Package passkey provides Passkey address derivation and authenticator assembly for Sui.
package passkey

import (
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/blake2b"

	"github.com/inodrahq/go-bcs"
	"github.com/inodrahq/go-sui-sdk/crypto"
)

// DeriveAddress derives a Passkey Sui address from a compressed secp256r1 public key.
// Blake2b-256(0x06 || compressed_pk_33bytes)
func DeriveAddress(compressedPublicKey []byte) (string, error) {
	if len(compressedPublicKey) != 33 {
		return "", fmt.Errorf("passkey: public key must be 33 bytes (compressed secp256r1), got %d", len(compressedPublicKey))
	}

	data := make([]byte, 0, 34)
	data = append(data, byte(crypto.PasskeyScheme))
	data = append(data, compressedPublicKey...)

	hash := blake2b.Sum256(data)
	return "0x" + fmt.Sprintf("%064x", hash), nil
}

// Assemble creates a Passkey authenticator from WebAuthn response data.
// Returns a base64-encoded authenticator with the Passkey flag prefix.
func Assemble(authenticatorData []byte, clientDataJSON string, userSignature string) (string, error) {
	e := bcs.NewEncoder()

	// authenticatorData: Vec<u8>
	e.WriteULEB128(uint32(len(authenticatorData)))
	e.WriteBytes(authenticatorData)

	// clientDataJson: BCS string (ULEB128 length + UTF-8)
	e.WriteULEB128(uint32(len(clientDataJSON)))
	e.WriteBytes([]byte(clientDataJSON))

	// userSignature: Vec<u8>
	sigBytes, err := base64.StdEncoding.DecodeString(userSignature)
	if err != nil {
		return "", fmt.Errorf("passkey: invalid base64 user signature: %w", err)
	}
	e.WriteULEB128(uint32(len(sigBytes)))
	e.WriteBytes(sigBytes)

	// Prepend flag
	result := make([]byte, 0, 1+len(e.Bytes()))
	result = append(result, byte(crypto.PasskeyScheme))
	result = append(result, e.Bytes()...)

	return base64.StdEncoding.EncodeToString(result), nil
}
