// Package txn provides BCS-serializable Sui transaction types.
package txn

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/inodrahq/go-bcs"
)

// SuiAddress is a 32-byte Sui address.
type SuiAddress [32]byte

// ParseAddress parses a hex-encoded Sui address (with or without 0x prefix).
func ParseAddress(s string) (SuiAddress, error) {
	s = strings.TrimPrefix(s, "0x")
	// Zero-pad to 64 hex chars
	if len(s) < 64 {
		s = strings.Repeat("0", 64-len(s)) + s
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return SuiAddress{}, fmt.Errorf("invalid address hex: %w", err)
	}
	if len(b) != 32 {
		return SuiAddress{}, fmt.Errorf("address must be 32 bytes, got %d", len(b))
	}
	var addr SuiAddress
	copy(addr[:], b)
	return addr, nil
}

// Hex returns the 0x-prefixed hex string.
func (a SuiAddress) Hex() string {
	return "0x" + hex.EncodeToString(a[:])
}

func (a SuiAddress) MarshalBCS(e *bcs.Encoder) error {
	return e.WriteBytes(a[:])
}

func (a *SuiAddress) UnmarshalBCS(d *bcs.Decoder) error {
	b, err := d.ReadBytes(32)
	if err != nil {
		return err
	}
	copy(a[:], b)
	return nil
}
