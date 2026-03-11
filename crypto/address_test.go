package crypto_test

import (
	"testing"

	"github.com/inodrahq/go-sui-sdk/crypto"
)

func TestDeriveAddress(t *testing.T) {
	// Zero public key with Ed25519 flag
	pubkey := make([]byte, 32)
	addr := crypto.DeriveAddress(0x00, pubkey)
	if len(addr) != 66 {
		t.Errorf("expected 66-char address, got %d", len(addr))
	}
	if addr[:2] != "0x" {
		t.Error("address should start with 0x")
	}
}

func TestNormalizeAddress(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"0x1", "0x0000000000000000000000000000000000000000000000000000000000000001", false},
		{"0x0000000000000000000000000000000000000000000000000000000000000001", "0x0000000000000000000000000000000000000000000000000000000000000001", false},
		{"1", "0x0000000000000000000000000000000000000000000000000000000000000001", false},
		{"0xABCD", "0x000000000000000000000000000000000000000000000000000000000000abcd", false},
		{"0x" + "ff" + "0000000000000000000000000000000000000000000000000000000000000001", "", true}, // too long
		{"0xGGGG", "", true}, // invalid hex
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := crypto.NormalizeAddress(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, got)
			}
		})
	}
}

func TestValidateAddress(t *testing.T) {
	if !crypto.ValidateAddress("0x0000000000000000000000000000000000000000000000000000000000000001") {
		t.Error("expected valid")
	}
	if !crypto.ValidateAddress("0x1") {
		t.Error("short address should be valid")
	}
	if crypto.ValidateAddress("0xZZZZ") {
		t.Error("expected invalid")
	}
}
