package mnemonic_test

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/inodrahq/go-sui-sdk/crypto/mnemonic"
)

func TestGenerate12Words(t *testing.T) {
	m, err := mnemonic.Generate(12)
	if err != nil {
		t.Fatal(err)
	}
	words := strings.Fields(m)
	if len(words) != 12 {
		t.Errorf("expected 12 words, got %d", len(words))
	}
	if !mnemonic.Validate(m) {
		t.Error("generated mnemonic should be valid")
	}
}

func TestGenerate24Words(t *testing.T) {
	m, err := mnemonic.Generate(24)
	if err != nil {
		t.Fatal(err)
	}
	words := strings.Fields(m)
	if len(words) != 24 {
		t.Errorf("expected 24 words, got %d", len(words))
	}
}

func TestGenerateInvalidWordCount(t *testing.T) {
	_, err := mnemonic.Generate(13)
	if err == nil {
		t.Error("expected error for invalid word count")
	}
}

func TestValidate(t *testing.T) {
	if !mnemonic.Validate("abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about") {
		t.Error("expected valid")
	}
	if mnemonic.Validate("not a valid mnemonic phrase at all") {
		t.Error("expected invalid")
	}
}

func TestToSeed(t *testing.T) {
	m := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	seed := mnemonic.ToSeed(m, "")
	if len(seed) != 64 {
		t.Fatalf("expected 64-byte seed, got %d", len(seed))
	}
	expected := "5eb00bbddcf069084889a8ab9155568165f5c453ccb85e70811aaed6f6da5fc19a5ac40b389cd370d086206dec8aa6c43daea6690f20ad3d8d48b2d2ce9e38e4"
	if hex.EncodeToString(seed) != expected {
		t.Errorf("seed mismatch:\n  got: %s\n  exp: %s", hex.EncodeToString(seed), expected)
	}
}

func TestToSeedWithPassphrase(t *testing.T) {
	m := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	seed1 := mnemonic.ToSeed(m, "")
	seed2 := mnemonic.ToSeed(m, "my secret passphrase")
	if hex.EncodeToString(seed1) == hex.EncodeToString(seed2) {
		t.Error("different passphrases should produce different seeds")
	}
}

func TestDeriveEd25519(t *testing.T) {
	m := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	seed := mnemonic.ToSeed(m, "")
	key, err := mnemonic.DeriveEd25519(seed, "m/44'/784'/0'/0'/0'")
	if err != nil {
		t.Fatal(err)
	}
	if len(key) != 32 {
		t.Fatalf("expected 32-byte key, got %d", len(key))
	}
}

func TestDeriveEd25519Deterministic(t *testing.T) {
	m := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	seed := mnemonic.ToSeed(m, "")
	key1, _ := mnemonic.DeriveEd25519(seed, "m/44'/784'/0'/0'/0'")
	key2, _ := mnemonic.DeriveEd25519(seed, "m/44'/784'/0'/0'/0'")
	if hex.EncodeToString(key1) != hex.EncodeToString(key2) {
		t.Error("derivation should be deterministic")
	}
}

func TestDeriveEd25519DifferentPaths(t *testing.T) {
	m := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	seed := mnemonic.ToSeed(m, "")
	key1, _ := mnemonic.DeriveEd25519(seed, "m/44'/784'/0'/0'/0'")
	key2, _ := mnemonic.DeriveEd25519(seed, "m/44'/784'/0'/0'/1'")
	if hex.EncodeToString(key1) == hex.EncodeToString(key2) {
		t.Error("different paths should produce different keys")
	}
}

func TestDeriveEd25519InvalidPath(t *testing.T) {
	seed := make([]byte, 64)
	_, err := mnemonic.DeriveEd25519(seed, "m/44/784/0/0/0") // non-hardened
	if err == nil {
		t.Error("expected error for non-hardened Ed25519 path")
	}
}

func TestDeriveEd25519InvalidPathComponent(t *testing.T) {
	seed := make([]byte, 64)
	_, err := mnemonic.DeriveEd25519(seed, "m/abc'/784'/0'/0'/0'")
	if err == nil {
		t.Error("expected error for invalid path component")
	}
}

func TestDeriveEd25519EmptyPath(t *testing.T) {
	seed := make([]byte, 64)
	key, err := mnemonic.DeriveEd25519(seed, "m/")
	if err != nil {
		t.Fatal(err)
	}
	// Should return master key
	if len(key) != 32 {
		t.Errorf("expected 32-byte key, got %d", len(key))
	}
}
