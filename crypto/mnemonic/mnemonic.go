// Package mnemonic provides BIP-39 mnemonic generation and SLIP-0010 key derivation.
package mnemonic

import (
	"fmt"

	"github.com/tyler-smith/go-bip39"
)

// Generate creates a new BIP-39 mnemonic with the given word count (12, 15, 18, 21, or 24).
func Generate(wordCount int) (string, error) {
	var bits int
	switch wordCount {
	case 12:
		bits = 128
	case 15:
		bits = 160
	case 18:
		bits = 192
	case 21:
		bits = 224
	case 24:
		bits = 256
	default:
		return "", fmt.Errorf("mnemonic: invalid word count %d (must be 12, 15, 18, 21, or 24)", wordCount)
	}
	entropy, err := bip39.NewEntropy(bits)
	if err != nil {
		return "", fmt.Errorf("mnemonic: generate entropy: %w", err)
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("mnemonic: create mnemonic: %w", err)
	}
	return mnemonic, nil
}

// Validate checks if a mnemonic phrase is valid.
func Validate(mnemonic string) bool {
	return bip39.IsMnemonicValid(mnemonic)
}

// ToSeed converts a mnemonic to a 64-byte seed with optional passphrase.
func ToSeed(mnemonic, passphrase string) []byte {
	return bip39.NewSeed(mnemonic, passphrase)
}
