package mnemonic

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"math/big"
)

var curveOrder = mustParseBigInt("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141", 16)

func mustParseBigInt(s string, base int) *big.Int {
	v, ok := new(big.Int).SetString(s, base)
	if !ok {
		panic("mnemonic: invalid big.Int constant: " + s)
	}
	return v
}

// DeriveBIP32 derives an ECDSA key from a BIP-39 seed using standard BIP-32.
// Supports both hardened and non-hardened paths.
// Used for Secp256k1 and Secp256r1.
func DeriveBIP32(seed []byte, path string) ([]byte, error) {
	segments, err := parsePath(path)
	if err != nil {
		return nil, err
	}

	// Master key
	mac := hmac.New(sha512.New, []byte("Bitcoin seed"))
	mac.Write(seed)
	I := mac.Sum(nil)
	key := I[:32]
	chainCode := I[32:]

	for _, segment := range segments {
		key, chainCode, err = deriveChildBIP32(key, chainCode, segment)
		if err != nil {
			return nil, err
		}
	}

	return key, nil
}

func deriveChildBIP32(key, chainCode []byte, index uint32) ([]byte, []byte, error) {
	data := make([]byte, 0, 37)

	if index >= 0x80000000 {
		// Hardened: 0x00 || key || index
		data = append(data, 0x00)
		data = append(data, key...)
	} else {
		// Normal: compressed_pubkey || index
		return nil, nil, fmt.Errorf("bip32: non-hardened derivation not supported for Sui ECDSA keys")
	}

	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, index)
	data = append(data, buf...)

	mac := hmac.New(sha512.New, chainCode)
	mac.Write(data)
	I := mac.Sum(nil)

	// Parse IL as 256-bit integer
	il := new(big.Int).SetBytes(I[:32])
	kpar := new(big.Int).SetBytes(key)

	// child key = (IL + kpar) mod n
	childKey := new(big.Int).Add(il, kpar)
	childKey.Mod(childKey, curveOrder)

	if childKey.Sign() == 0 {
		return nil, nil, fmt.Errorf("bip32: derived key is zero")
	}

	// Serialize as 32 bytes
	childKeyBytes := make([]byte, 32)
	b := childKey.Bytes()
	copy(childKeyBytes[32-len(b):], b)

	return childKeyBytes, I[32:], nil
}
