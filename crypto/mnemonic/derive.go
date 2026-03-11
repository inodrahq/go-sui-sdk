package mnemonic

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

// SLIP-0010 derivation for Ed25519 (hardened only).
// Path format: m/44'/784'/0'/0'/0'

// DeriveEd25519 derives an Ed25519 seed from a BIP-39 seed using SLIP-0010.
// The path must use hardened indices only (e.g., "m/44'/784'/0'/0'/0'").
func DeriveEd25519(seed []byte, path string) ([]byte, error) {
	segments, err := parsePath(path)
	if err != nil {
		return nil, err
	}

	// Master key
	mac := hmac.New(sha512.New, []byte("ed25519 seed"))
	mac.Write(seed)
	I := mac.Sum(nil)
	key := I[:32]
	chainCode := I[32:]

	// Derive child keys
	for _, segment := range segments {
		key, chainCode, err = deriveChildEd25519(key, chainCode, segment)
		if err != nil {
			return nil, err
		}
	}

	return key, nil
}

func deriveChildEd25519(key, chainCode []byte, index uint32) ([]byte, []byte, error) {
	// SLIP-0010: Ed25519 only supports hardened derivation
	if index < 0x80000000 {
		return nil, nil, fmt.Errorf("slip0010: Ed25519 requires hardened derivation (index >= 0x80000000)")
	}

	data := make([]byte, 1+32+4)
	data[0] = 0x00
	copy(data[1:33], key)
	binary.BigEndian.PutUint32(data[33:], index)

	mac := hmac.New(sha512.New, chainCode)
	mac.Write(data)
	I := mac.Sum(nil)

	return I[:32], I[32:], nil
}

func parsePath(path string) ([]uint32, error) {
	path = strings.TrimPrefix(path, "m/")
	if path == "" {
		return nil, nil
	}
	parts := strings.Split(path, "/")
	indices := make([]uint32, len(parts))
	for i, part := range parts {
		hardened := strings.HasSuffix(part, "'")
		if hardened {
			part = part[:len(part)-1]
		}
		n, err := strconv.ParseUint(part, 10, 31)
		if err != nil {
			return nil, fmt.Errorf("slip0010: invalid path component %q: %w", parts[i], err)
		}
		idx := uint32(n)
		if hardened {
			idx += 0x80000000
		}
		indices[i] = idx
	}
	return indices, nil
}
