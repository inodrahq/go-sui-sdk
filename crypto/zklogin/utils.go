package zklogin

import (
	"fmt"
	"math/big"
	"strings"
)

// GenAddressSeed generates the address seed from JWT claims and salt.
// salt, name, value, aud are as in the TypeScript SDK's genAddressSeed.
func GenAddressSeed(salt, name, value, aud string) (string, error) {
	hashedName, err := HashASCIIStrToField(name)
	if err != nil {
		return "", err
	}
	hashedValue, err := HashASCIIStrToField(value)
	if err != nil {
		return "", err
	}
	hashedAud, err := HashASCIIStrToField(aud)
	if err != nil {
		return "", err
	}
	hashedSalt, err := PoseidonHash([]string{salt})
	if err != nil {
		return "", err
	}

	return PoseidonHash([]string{hashedName, hashedValue, hashedAud, hashedSalt})
}

// HashASCIIStrToField hashes an ASCII string to a BN254 field element.
// Characters are packed into 31-byte chunks (252 bits) and hashed with Poseidon.
func HashASCIIStrToField(s string) (string, error) {
	packed := packASCII(s)

	if len(packed) == 0 {
		return PoseidonHash([]string{"0"})
	}
	return PoseidonHash(packed)
}

// NormalizeIssuer ensures the issuer has an https:// prefix.
func NormalizeIssuer(iss string) string {
	if !strings.HasPrefix(iss, "https://") && !strings.HasPrefix(iss, "http://") {
		return "https://" + iss
	}
	return iss
}

// packASCII packs an ASCII string into BN254 field elements (31 bytes per element).
func packASCII(s string) []string {
	var packed []string
	for i := 0; i < len(s); i += 31 {
		end := i + 31
		if end > len(s) {
			end = len(s)
		}
		chunk := s[i:end]
		val := new(big.Int)
		for _, b := range []byte(chunk) {
			val.Mul(val, big.NewInt(256))
			val.Add(val, big.NewInt(int64(b)))
		}
		packed = append(packed, val.String())
	}
	return packed
}

// AddressSeedToBytes converts address seed (decimal string) to 32 big-endian bytes.
func AddressSeedToBytes(seed string) ([]byte, error) {
	val, ok := new(big.Int).SetString(seed, 10)
	if !ok {
		return nil, fmt.Errorf("zklogin: invalid address seed %q", seed)
	}
	b := val.Bytes() // big-endian
	if len(b) > 32 {
		return nil, fmt.Errorf("zklogin: address seed too large")
	}
	// Zero-pad to 32 bytes
	result := make([]byte, 32)
	copy(result[32-len(b):], b)
	return result, nil
}
