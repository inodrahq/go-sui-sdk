package mnemonic_test

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/inodrahq/go-sui-sdk/crypto/mnemonic"
)

// --------------- Generate: all valid word counts ---------------

func TestGenerate15Words(t *testing.T) {
	m, err := mnemonic.Generate(15)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(strings.Fields(m)); n != 15 {
		t.Errorf("expected 15 words, got %d", n)
	}
	if !mnemonic.Validate(m) {
		t.Error("generated mnemonic should be valid")
	}
}

func TestGenerate18Words(t *testing.T) {
	m, err := mnemonic.Generate(18)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(strings.Fields(m)); n != 18 {
		t.Errorf("expected 18 words, got %d", n)
	}
	if !mnemonic.Validate(m) {
		t.Error("generated mnemonic should be valid")
	}
}

func TestGenerate21Words(t *testing.T) {
	m, err := mnemonic.Generate(21)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(strings.Fields(m)); n != 21 {
		t.Errorf("expected 21 words, got %d", n)
	}
	if !mnemonic.Validate(m) {
		t.Error("generated mnemonic should be valid")
	}
}

// --------------- Generate: invalid word counts ---------------

func TestGenerateInvalidWordCounts(t *testing.T) {
	for _, wc := range []int{0, 1, 7, 11, 13, 16, 23, 25, 100, -1} {
		_, err := mnemonic.Generate(wc)
		if err == nil {
			t.Errorf("expected error for word count %d", wc)
		}
	}
}

// --------------- Validate: edge cases ---------------

func TestValidateEmptyString(t *testing.T) {
	if mnemonic.Validate("") {
		t.Error("empty string should be invalid")
	}
}

func TestValidateSingleWord(t *testing.T) {
	if mnemonic.Validate("abandon") {
		t.Error("single word should be invalid")
	}
}

func TestValidateWrongChecksum(t *testing.T) {
	// Valid structure but wrong last word (checksum mismatch)
	if mnemonic.Validate("abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon") {
		t.Error("wrong checksum should be invalid")
	}
}

func TestValidate15WordMnemonic(t *testing.T) {
	// Generate a valid 15-word mnemonic and verify it validates
	m, err := mnemonic.Generate(15)
	if err != nil {
		t.Fatal(err)
	}
	if !mnemonic.Validate(m) {
		t.Error("generated 15-word mnemonic should pass validation")
	}
}

func TestValidate24WordMnemonic(t *testing.T) {
	m := "zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo vote"
	if !mnemonic.Validate(m) {
		t.Error("valid 24-word mnemonic should pass validation")
	}
}

// --------------- ToSeed: deterministic ---------------

func TestToSeedDeterministic(t *testing.T) {
	m := "zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo vote"
	seed1 := mnemonic.ToSeed(m, "test")
	seed2 := mnemonic.ToSeed(m, "test")
	if hex.EncodeToString(seed1) != hex.EncodeToString(seed2) {
		t.Error("ToSeed should be deterministic")
	}
	if len(seed1) != 64 {
		t.Errorf("expected 64-byte seed, got %d", len(seed1))
	}
}

func TestToSeedDifferentMnemonics(t *testing.T) {
	m1 := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	m2 := "zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo wrong"
	s1 := mnemonic.ToSeed(m1, "")
	s2 := mnemonic.ToSeed(m2, "")
	if hex.EncodeToString(s1) == hex.EncodeToString(s2) {
		t.Error("different mnemonics should produce different seeds")
	}
}

// --------------- DeriveEd25519: deeper paths ---------------

func TestDeriveEd25519DeeperPath(t *testing.T) {
	m := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	seed := mnemonic.ToSeed(m, "")
	key, err := mnemonic.DeriveEd25519(seed, "m/44'/784'/1'/0'/0'")
	if err != nil {
		t.Fatal(err)
	}
	if len(key) != 32 {
		t.Fatalf("expected 32-byte key, got %d", len(key))
	}
	// Must differ from account 0
	key0, _ := mnemonic.DeriveEd25519(seed, "m/44'/784'/0'/0'/0'")
	if hex.EncodeToString(key) == hex.EncodeToString(key0) {
		t.Error("different account indices should produce different keys")
	}
}

func TestDeriveEd25519Account2(t *testing.T) {
	m := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	seed := mnemonic.ToSeed(m, "")
	key, err := mnemonic.DeriveEd25519(seed, "m/44'/784'/2'/0'/0'")
	if err != nil {
		t.Fatal(err)
	}
	if len(key) != 32 {
		t.Fatalf("expected 32-byte key, got %d", len(key))
	}
}

// --------------- DeriveEd25519: invalid paths ---------------

func TestDeriveEd25519NoMPrefix(t *testing.T) {
	seed := make([]byte, 64)
	// Without m/ prefix, parsePath trims "m/" so "44'/784'/0'/0'/0'" is valid.
	// But "44'/784'/0'/0'/0'" without m/ should still work (parsePath does TrimPrefix).
	// Test truly bad paths:
	_, err := mnemonic.DeriveEd25519(seed, "m/44'/abc'/0'/0'/0'")
	if err == nil {
		t.Error("expected error for non-numeric path component")
	}
}

func TestDeriveEd25519NegativeIndex(t *testing.T) {
	seed := make([]byte, 64)
	_, err := mnemonic.DeriveEd25519(seed, "m/-1'/784'/0'/0'/0'")
	if err == nil {
		t.Error("expected error for negative index")
	}
}

func TestDeriveEd25519LargeIndex(t *testing.T) {
	seed := make([]byte, 64)
	// 2^31 = 2147483648, exceeds uint31 max
	_, err := mnemonic.DeriveEd25519(seed, "m/2147483648'/784'/0'/0'/0'")
	if err == nil {
		t.Error("expected error for index exceeding 31-bit range")
	}
}

func TestDeriveEd25519MixedHardenedNonHardened(t *testing.T) {
	seed := make([]byte, 64)
	// Non-hardened 0 in middle of otherwise hardened path
	_, err := mnemonic.DeriveEd25519(seed, "m/44'/784'/0/0'/0'")
	if err == nil {
		t.Error("expected error for non-hardened segment in Ed25519 path")
	}
}

func TestDeriveEd25519SingleHardenedSegment(t *testing.T) {
	seed := make([]byte, 64)
	key, err := mnemonic.DeriveEd25519(seed, "m/0'")
	if err != nil {
		t.Fatal(err)
	}
	if len(key) != 32 {
		t.Fatalf("expected 32-byte key, got %d", len(key))
	}
}

// --------------- DeriveBIP32: hardened path ---------------

func TestDeriveBIP32HardenedPath(t *testing.T) {
	m := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	seed := mnemonic.ToSeed(m, "")
	key, err := mnemonic.DeriveBIP32(seed, "m/44'/784'/0'/0'/0'")
	if err != nil {
		t.Fatal(err)
	}
	if len(key) != 32 {
		t.Fatalf("expected 32-byte key, got %d", len(key))
	}
}

func TestDeriveBIP32Deterministic(t *testing.T) {
	m := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	seed := mnemonic.ToSeed(m, "")
	key1, err := mnemonic.DeriveBIP32(seed, "m/44'/784'/0'/0'/0'")
	if err != nil {
		t.Fatal(err)
	}
	key2, err := mnemonic.DeriveBIP32(seed, "m/44'/784'/0'/0'/0'")
	if err != nil {
		t.Fatal(err)
	}
	if hex.EncodeToString(key1) != hex.EncodeToString(key2) {
		t.Error("BIP32 derivation should be deterministic")
	}
}

func TestDeriveBIP32DifferentPaths(t *testing.T) {
	m := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	seed := mnemonic.ToSeed(m, "")
	key1, _ := mnemonic.DeriveBIP32(seed, "m/44'/784'/0'/0'/0'")
	key2, _ := mnemonic.DeriveBIP32(seed, "m/44'/784'/0'/0'/1'")
	if hex.EncodeToString(key1) == hex.EncodeToString(key2) {
		t.Error("different BIP32 paths should produce different keys")
	}
}

func TestDeriveBIP32DifferentFromEd25519(t *testing.T) {
	m := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	seed := mnemonic.ToSeed(m, "")
	keyBIP32, _ := mnemonic.DeriveBIP32(seed, "m/44'/784'/0'/0'/0'")
	keyEd, _ := mnemonic.DeriveEd25519(seed, "m/44'/784'/0'/0'/0'")
	if hex.EncodeToString(keyBIP32) == hex.EncodeToString(keyEd) {
		t.Error("BIP32 and Ed25519 derivation should produce different keys")
	}
}

func TestDeriveBIP32EmptyPath(t *testing.T) {
	seed := make([]byte, 64)
	key, err := mnemonic.DeriveBIP32(seed, "m/")
	if err != nil {
		t.Fatal(err)
	}
	if len(key) != 32 {
		t.Fatalf("expected 32-byte master key, got %d", len(key))
	}
}

func TestDeriveBIP32SingleSegment(t *testing.T) {
	seed := make([]byte, 64)
	key, err := mnemonic.DeriveBIP32(seed, "m/44'")
	if err != nil {
		t.Fatal(err)
	}
	if len(key) != 32 {
		t.Fatalf("expected 32-byte key, got %d", len(key))
	}
}

// --------------- DeriveBIP32: error cases ---------------

func TestDeriveBIP32NonHardenedErrors(t *testing.T) {
	seed := make([]byte, 64)
	_, err := mnemonic.DeriveBIP32(seed, "m/44/784/0/0/0")
	if err == nil {
		t.Error("expected error for non-hardened BIP32 path")
	}
}

func TestDeriveBIP32MixedNonHardened(t *testing.T) {
	seed := make([]byte, 64)
	// First segment hardened, second non-hardened
	_, err := mnemonic.DeriveBIP32(seed, "m/44'/0")
	if err == nil {
		t.Error("expected error for non-hardened segment in BIP32 path")
	}
}

func TestDeriveBIP32InvalidPathComponent(t *testing.T) {
	seed := make([]byte, 64)
	_, err := mnemonic.DeriveBIP32(seed, "m/xyz'/0'/0'")
	if err == nil {
		t.Error("expected error for invalid path component")
	}
}

func TestDeriveBIP32NegativeIndex(t *testing.T) {
	seed := make([]byte, 64)
	_, err := mnemonic.DeriveBIP32(seed, "m/-1'/0'/0'")
	if err == nil {
		t.Error("expected error for negative index")
	}
}

func TestDeriveBIP32OverflowIndex(t *testing.T) {
	seed := make([]byte, 64)
	_, err := mnemonic.DeriveBIP32(seed, "m/2147483648'/0'/0'")
	if err == nil {
		t.Error("expected error for overflowing index")
	}
}

// --------------- parsePath via DeriveEd25519 (parsePath is unexported) ---------------

func TestParsePathSpecialCharacters(t *testing.T) {
	seed := make([]byte, 64)
	_, err := mnemonic.DeriveEd25519(seed, "m/44'/784'/@/0'/0'")
	if err == nil {
		t.Error("expected error for special character in path")
	}
}

func TestParsePathEmptySegment(t *testing.T) {
	seed := make([]byte, 64)
	_, err := mnemonic.DeriveEd25519(seed, "m/44'//0'/0'/0'")
	if err == nil {
		t.Error("expected error for empty segment in path")
	}
}

func TestParsePathTrailingSlash(t *testing.T) {
	seed := make([]byte, 64)
	// "m/44'/" splits to ["44'", ""] -- empty last segment should error
	_, err := mnemonic.DeriveEd25519(seed, "m/44'/")
	if err == nil {
		t.Error("expected error for trailing slash in path")
	}
}

func TestParsePathJustM(t *testing.T) {
	seed := make([]byte, 64)
	// "m/" with TrimPrefix becomes "" -> returns nil segments -> master key
	key, err := mnemonic.DeriveEd25519(seed, "m/")
	if err != nil {
		t.Fatal(err)
	}
	if len(key) != 32 {
		t.Errorf("expected 32-byte master key, got %d", len(key))
	}
}
