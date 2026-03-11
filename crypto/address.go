package crypto

import (
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/crypto/blake2b"
)

// DeriveAddress computes a Sui address from a signature flag and public key bytes.
// address = "0x" + hex(Blake2b256(flag || pubkey_bytes))
func DeriveAddress(flag byte, pubkeyBytes []byte) string {
	data := make([]byte, 1+len(pubkeyBytes))
	data[0] = flag
	copy(data[1:], pubkeyBytes)
	hash := blake2b.Sum256(data)
	return "0x" + hex.EncodeToString(hash[:])
}

// NormalizeAddress normalizes a Sui address to lowercase with 0x prefix and zero-padding to 64 hex chars.
func NormalizeAddress(addr string) (string, error) {
	addr = strings.TrimPrefix(strings.ToLower(addr), "0x")
	if len(addr) > 64 {
		return "", fmt.Errorf("address too long: %d hex chars", len(addr))
	}
	// Zero-pad to 64 hex chars
	addr = strings.Repeat("0", 64-len(addr)) + addr
	// Validate hex
	if _, err := hex.DecodeString(addr); err != nil {
		return "", fmt.Errorf("invalid hex address: %w", err)
	}
	return "0x" + addr, nil
}

// ValidateAddress checks if a string is a valid Sui address.
func ValidateAddress(addr string) bool {
	_, err := NormalizeAddress(addr)
	return err == nil
}
