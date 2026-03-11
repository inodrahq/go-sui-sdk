package wallet

import (
	"fmt"
	"strings"
)

const (
	bech32Charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"
	hrpPrivateKey = "suiprivkey"
)

// bech32Decode decodes a bech32-encoded string and returns the HRP and data bytes.
func bech32Decode(encoded string) (string, []byte, error) {
	encoded = strings.ToLower(encoded)

	sepIdx := strings.LastIndex(encoded, "1")
	if sepIdx < 1 {
		return "", nil, fmt.Errorf("bech32: no separator found")
	}

	hrp := encoded[:sepIdx]
	data := encoded[sepIdx+1:]

	if len(data) < 6 {
		return "", nil, fmt.Errorf("bech32: data too short")
	}

	// Decode base32 characters
	var values []int
	for _, c := range data {
		idx := strings.IndexRune(bech32Charset, c)
		if idx < 0 {
			return "", nil, fmt.Errorf("bech32: invalid character %c", c)
		}
		values = append(values, idx)
	}

	// Verify checksum
	if !bech32VerifyChecksum(hrp, values) {
		return "", nil, fmt.Errorf("bech32: invalid checksum")
	}

	// Strip checksum (last 6 values)
	values = values[:len(values)-6]

	// Convert from 5-bit to 8-bit groups
	bytes, err := convertBits(values, 5, 8, false)
	if err != nil {
		return "", nil, fmt.Errorf("bech32: %w", err)
	}

	return hrp, bytes, nil
}

// bech32Encode encodes data bytes with the given HRP into a bech32 string.
func bech32Encode(hrp string, data []byte) (string, error) {
	// Convert from 8-bit to 5-bit groups
	values, err := convertBits(byteSliceToInts(data), 8, 5, true)
	if err != nil {
		return "", fmt.Errorf("bech32: %w", err)
	}

	intValues := make([]int, len(values))
	for i, b := range values {
		intValues[i] = int(b)
	}

	// Create checksum
	checksum := bech32CreateChecksum(hrp, intValues)
	intValues = append(intValues, checksum...)

	// Encode to string
	var result strings.Builder
	result.WriteString(hrp)
	result.WriteByte('1')
	for _, v := range intValues {
		result.WriteByte(bech32Charset[v])
	}

	return result.String(), nil
}

func byteSliceToInts(data []byte) []int {
	result := make([]int, len(data))
	for i, b := range data {
		result[i] = int(b)
	}
	return result
}

func convertBits(data []int, fromBits, toBits int, pad bool) ([]byte, error) {
	acc := 0
	bits := 0
	maxV := (1 << toBits) - 1
	var result []byte

	for _, v := range data {
		if v < 0 || v>>fromBits != 0 {
			return nil, fmt.Errorf("invalid value: %d", v)
		}
		acc = (acc << fromBits) | v
		bits += fromBits
		for bits >= toBits {
			bits -= toBits
			result = append(result, byte((acc>>bits)&maxV))
		}
	}

	if pad {
		if bits > 0 {
			result = append(result, byte((acc<<(toBits-bits))&maxV))
		}
	} else if bits >= fromBits {
		return nil, fmt.Errorf("excess padding")
	} else if (acc<<(toBits-bits))&maxV != 0 {
		return nil, fmt.Errorf("non-zero padding")
	}

	return result, nil
}

func bech32Polymod(values []int) int {
	gen := []int{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}
	chk := 1
	for _, v := range values {
		b := chk >> 25
		chk = (chk&0x1ffffff)<<5 ^ v
		for i := 0; i < 5; i++ {
			if (b>>i)&1 == 1 {
				chk ^= gen[i]
			}
		}
	}
	return chk
}

func bech32HRPExpand(hrp string) []int {
	var result []int
	for _, c := range hrp {
		result = append(result, int(c>>5))
	}
	result = append(result, 0)
	for _, c := range hrp {
		result = append(result, int(c&31))
	}
	return result
}

func bech32VerifyChecksum(hrp string, data []int) bool {
	values := append(bech32HRPExpand(hrp), data...)
	return bech32Polymod(values) == 1
}

func bech32CreateChecksum(hrp string, data []int) []int {
	values := append(bech32HRPExpand(hrp), data...)
	values = append(values, 0, 0, 0, 0, 0, 0)
	polymod := bech32Polymod(values) ^ 1
	var checksum []int
	for i := 0; i < 6; i++ {
		checksum = append(checksum, (polymod>>(5*(5-i)))&31)
	}
	return checksum
}
