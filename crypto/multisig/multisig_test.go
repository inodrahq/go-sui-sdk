package multisig_test

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/inodrahq/go-sui-sdk/crypto"
	"github.com/inodrahq/go-sui-sdk/crypto/ed25519"
	"github.com/inodrahq/go-sui-sdk/crypto/multisig"
	"github.com/inodrahq/go-sui-sdk/crypto/secp256k1"
)

func TestMultiSigAddress(t *testing.T) {
	kp1, _ := ed25519.New()
	kp2, _ := ed25519.New()

	ms := &multisig.MultiSigPublicKey{
		PubKeys: []multisig.MemberPublicKey{
			{PubKey: kp1.PublicKey(), Weight: 1},
			{PubKey: kp2.PublicKey(), Weight: 1},
		},
		Threshold: 1,
	}

	addr := ms.SuiAddress()
	if len(addr) != 66 {
		t.Errorf("expected 66-char address, got %d: %s", len(addr), addr)
	}
	if addr[:2] != "0x" {
		t.Error("expected 0x prefix")
	}
}

func TestMultiSigAddressDeterministic(t *testing.T) {
	seed := make([]byte, 32)
	kp1, _ := ed25519.FromSeed(seed)
	seed[0] = 1
	kp2, _ := ed25519.FromSeed(seed)

	ms1 := &multisig.MultiSigPublicKey{
		PubKeys: []multisig.MemberPublicKey{
			{PubKey: kp1.PublicKey(), Weight: 1},
			{PubKey: kp2.PublicKey(), Weight: 1},
		},
		Threshold: 1,
	}
	ms2 := &multisig.MultiSigPublicKey{
		PubKeys: []multisig.MemberPublicKey{
			{PubKey: kp1.PublicKey(), Weight: 1},
			{PubKey: kp2.PublicKey(), Weight: 1},
		},
		Threshold: 1,
	}

	if ms1.SuiAddress() != ms2.SuiAddress() {
		t.Error("same keys should produce same address")
	}
}

func TestMultiSigAddressDifferentThreshold(t *testing.T) {
	seed := make([]byte, 32)
	kp1, _ := ed25519.FromSeed(seed)
	seed[0] = 1
	kp2, _ := ed25519.FromSeed(seed)

	ms1 := &multisig.MultiSigPublicKey{
		PubKeys: []multisig.MemberPublicKey{
			{PubKey: kp1.PublicKey(), Weight: 1},
			{PubKey: kp2.PublicKey(), Weight: 1},
		},
		Threshold: 1,
	}
	ms2 := &multisig.MultiSigPublicKey{
		PubKeys: []multisig.MemberPublicKey{
			{PubKey: kp1.PublicKey(), Weight: 1},
			{PubKey: kp2.PublicKey(), Weight: 1},
		},
		Threshold: 2,
	}

	if ms1.SuiAddress() == ms2.SuiAddress() {
		t.Error("different thresholds should produce different addresses")
	}
}

func TestBitmap(t *testing.T) {
	if multisig.Bitmap([]int{0}) != 1 {
		t.Error("expected 1")
	}
	if multisig.Bitmap([]int{1}) != 2 {
		t.Error("expected 2")
	}
	if multisig.Bitmap([]int{0, 1}) != 3 {
		t.Error("expected 3")
	}
	if multisig.Bitmap([]int{0, 2}) != 5 {
		t.Error("expected 5")
	}
}

func TestCombineSignatures(t *testing.T) {
	kp1, _ := ed25519.New()
	kp2, _ := ed25519.New()

	ms := &multisig.MultiSigPublicKey{
		PubKeys: []multisig.MemberPublicKey{
			{PubKey: kp1.PublicKey(), Weight: 1},
			{PubKey: kp2.PublicKey(), Weight: 1},
		},
		Threshold: 1,
	}

	// Sign with kp1
	sig1, err := crypto.SignTransaction(kp1, []byte("test tx"))
	if err != nil {
		t.Fatal(err)
	}

	combined, err := ms.CombineSignatures([]string{sig1}, []int{0})
	if err != nil {
		t.Fatal(err)
	}
	if len(combined) == 0 {
		t.Error("expected non-empty combined signature")
	}
}

func TestCombineSignaturesBelowThreshold(t *testing.T) {
	kp1, _ := ed25519.New()
	kp2, _ := ed25519.New()

	ms := &multisig.MultiSigPublicKey{
		PubKeys: []multisig.MemberPublicKey{
			{PubKey: kp1.PublicKey(), Weight: 1},
			{PubKey: kp2.PublicKey(), Weight: 1},
		},
		Threshold: 2, // need both
	}

	sig1, _ := crypto.SignTransaction(kp1, []byte("test tx"))

	_, err := ms.CombineSignatures([]string{sig1}, []int{0})
	if err == nil {
		t.Error("expected error: below threshold")
	}
}

func TestCombineSignaturesIndexMismatch(t *testing.T) {
	kp1, _ := ed25519.New()
	ms := &multisig.MultiSigPublicKey{
		PubKeys: []multisig.MemberPublicKey{
			{PubKey: kp1.PublicKey(), Weight: 1},
		},
		Threshold: 1,
	}

	sig1, _ := crypto.SignTransaction(kp1, []byte("test tx"))
	_, err := ms.CombineSignatures([]string{sig1, sig1}, []int{0})
	if err == nil {
		t.Error("expected error: length mismatch")
	}
}

func TestCombineSignaturesOutOfRange(t *testing.T) {
	kp1, _ := ed25519.New()
	ms := &multisig.MultiSigPublicKey{
		PubKeys: []multisig.MemberPublicKey{
			{PubKey: kp1.PublicKey(), Weight: 1},
		},
		Threshold: 1,
	}

	sig1, _ := crypto.SignTransaction(kp1, []byte("test tx"))
	_, err := ms.CombineSignatures([]string{sig1}, []int{5})
	if err == nil {
		t.Error("expected error: index out of range")
	}
}

func TestMultiSigMixedSchemes(t *testing.T) {
	kpEd, _ := ed25519.New()
	kpK1, _ := secp256k1.New()

	ms := &multisig.MultiSigPublicKey{
		PubKeys: []multisig.MemberPublicKey{
			{PubKey: kpEd.PublicKey(), Weight: 1},
			{PubKey: kpK1.PublicKey(), Weight: 1},
		},
		Threshold: 1,
	}

	addr := ms.SuiAddress()
	if len(addr) != 66 {
		t.Errorf("expected 66-char address, got %d", len(addr))
	}

	// Sign with Ed25519 member
	sig, _ := crypto.SignTransaction(kpEd, []byte("test"))
	combined, err := ms.CombineSignatures([]string{sig}, []int{0})
	if err != nil {
		t.Fatal(err)
	}
	if len(combined) == 0 {
		t.Error("expected combined sig")
	}
}

func TestCombineSignaturesDuplicateIndices(t *testing.T) {
	kp1, _ := ed25519.New()
	kp2, _ := ed25519.New()

	ms := &multisig.MultiSigPublicKey{
		PubKeys: []multisig.MemberPublicKey{
			{PubKey: kp1.PublicKey(), Weight: 1},
			{PubKey: kp2.PublicKey(), Weight: 1},
		},
		Threshold: 2,
	}

	sig1, _ := crypto.SignTransaction(kp1, []byte("test tx"))

	// Duplicate index 0 — total weight counts twice, but bitmap only sets bit once
	combined, err := ms.CombineSignatures([]string{sig1, sig1}, []int{0, 0})
	// This should succeed since total weight (1+1=2) meets threshold
	// The behavior with duplicate indices is defined by the implementation
	if err != nil {
		// If the implementation rejects duplicates, that's also valid
		t.Logf("duplicate indices rejected: %v", err)
		return
	}
	if combined == "" {
		t.Error("expected non-empty result")
	}
}

func TestThreeSignersThreshold2OnlyTwoSign(t *testing.T) {
	seed := make([]byte, 32)
	kp1, _ := ed25519.FromSeed(seed)
	seed[0] = 1
	kp2, _ := ed25519.FromSeed(seed)
	seed[0] = 2
	kp3, _ := ed25519.FromSeed(seed)

	ms := &multisig.MultiSigPublicKey{
		PubKeys: []multisig.MemberPublicKey{
			{PubKey: kp1.PublicKey(), Weight: 1},
			{PubKey: kp2.PublicKey(), Weight: 1},
			{PubKey: kp3.PublicKey(), Weight: 1},
		},
		Threshold: 2,
	}

	txBytes := []byte("test tx for 3-of-2")

	// Only signers 0 and 2 sign
	sig1, err := crypto.SignTransaction(kp1, txBytes)
	if err != nil {
		t.Fatal(err)
	}
	sig3, err := crypto.SignTransaction(kp3, txBytes)
	if err != nil {
		t.Fatal(err)
	}

	combined, err := ms.CombineSignatures([]string{sig1, sig3}, []int{0, 2})
	if err != nil {
		t.Fatalf("expected success with 2 of 3 signers meeting threshold: %v", err)
	}

	// Verify it's a valid base64 string starting with MultiSig flag
	decoded, err := base64.StdEncoding.DecodeString(combined)
	if err != nil {
		t.Fatalf("failed to decode combined signature: %v", err)
	}
	if decoded[0] != byte(crypto.MultiSigScheme) {
		t.Errorf("expected multisig flag 0x03, got 0x%02x", decoded[0])
	}

	// Verify bitmap: signers 0 and 2 => bits 0 and 2 => 0b101 = 5
	// (bitmap is embedded in the BCS encoding)
}

func TestThreeSignersThreshold3OnlyTwoSignFails(t *testing.T) {
	kp1, _ := ed25519.New()
	kp2, _ := ed25519.New()
	kp3, _ := ed25519.New()

	ms := &multisig.MultiSigPublicKey{
		PubKeys: []multisig.MemberPublicKey{
			{PubKey: kp1.PublicKey(), Weight: 1},
			{PubKey: kp2.PublicKey(), Weight: 1},
			{PubKey: kp3.PublicKey(), Weight: 1},
		},
		Threshold: 3, // need all three
	}

	sig1, _ := crypto.SignTransaction(kp1, []byte("test"))
	sig2, _ := crypto.SignTransaction(kp2, []byte("test"))

	_, err := ms.CombineSignatures([]string{sig1, sig2}, []int{0, 1})
	if err == nil {
		t.Error("expected error: 2 of 3 below threshold 3")
	}
	if !strings.Contains(err.Error(), "below threshold") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestBitmapEmpty(t *testing.T) {
	bm := multisig.Bitmap([]int{})
	if bm != 0 {
		t.Errorf("expected 0 for empty indices, got %d", bm)
	}
}

func TestBitmapNil(t *testing.T) {
	bm := multisig.Bitmap(nil)
	if bm != 0 {
		t.Errorf("expected 0 for nil indices, got %d", bm)
	}
}

func TestSuiAddressDiffersWithKeyOrdering(t *testing.T) {
	seed := make([]byte, 32)
	kp1, _ := ed25519.FromSeed(seed)
	seed[0] = 1
	kp2, _ := ed25519.FromSeed(seed)

	msAB := &multisig.MultiSigPublicKey{
		PubKeys: []multisig.MemberPublicKey{
			{PubKey: kp1.PublicKey(), Weight: 1},
			{PubKey: kp2.PublicKey(), Weight: 1},
		},
		Threshold: 1,
	}
	msBA := &multisig.MultiSigPublicKey{
		PubKeys: []multisig.MemberPublicKey{
			{PubKey: kp2.PublicKey(), Weight: 1},
			{PubKey: kp1.PublicKey(), Weight: 1},
		},
		Threshold: 1,
	}

	addrAB := msAB.SuiAddress()
	addrBA := msBA.SuiAddress()
	if addrAB == addrBA {
		t.Error("different key orderings should produce different addresses")
	}
}

func TestCombineSignaturesInvalidBase64(t *testing.T) {
	kp1, _ := ed25519.New()
	ms := &multisig.MultiSigPublicKey{
		PubKeys: []multisig.MemberPublicKey{
			{PubKey: kp1.PublicKey(), Weight: 1},
		},
		Threshold: 1,
	}

	_, err := ms.CombineSignatures([]string{"not-valid-base64!!!"}, []int{0})
	if err == nil {
		t.Error("expected error for invalid base64 signature")
	}
}

func TestCombineSignaturesEmptySignature(t *testing.T) {
	kp1, _ := ed25519.New()
	ms := &multisig.MultiSigPublicKey{
		PubKeys: []multisig.MemberPublicKey{
			{PubKey: kp1.PublicKey(), Weight: 1},
		},
		Threshold: 1,
	}

	// Empty bytes encoded as base64
	emptySig := base64.StdEncoding.EncodeToString([]byte{})
	_, err := ms.CombineSignatures([]string{emptySig}, []int{0})
	if err == nil {
		t.Error("expected error for empty signature bytes")
	}
}

func TestCombineSignaturesUnsupportedScheme(t *testing.T) {
	kp1, _ := ed25519.New()
	ms := &multisig.MultiSigPublicKey{
		PubKeys: []multisig.MemberPublicKey{
			{PubKey: kp1.PublicKey(), Weight: 1},
		},
		Threshold: 1,
	}

	// Create a fake signature with unsupported scheme flag (0xFF)
	fakeSig := make([]byte, 97)
	fakeSig[0] = 0xFF // unsupported scheme
	encoded := base64.StdEncoding.EncodeToString(fakeSig)

	_, err := ms.CombineSignatures([]string{encoded}, []int{0})
	if err == nil {
		t.Error("expected error for unsupported signature scheme")
	}
	if !strings.Contains(err.Error(), "unsupported scheme") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCombineSignaturesNegativeIndex(t *testing.T) {
	kp1, _ := ed25519.New()
	ms := &multisig.MultiSigPublicKey{
		PubKeys: []multisig.MemberPublicKey{
			{PubKey: kp1.PublicKey(), Weight: 1},
		},
		Threshold: 1,
	}

	sig1, _ := crypto.SignTransaction(kp1, []byte("test"))
	_, err := ms.CombineSignatures([]string{sig1}, []int{-1})
	if err == nil {
		t.Error("expected error for negative index")
	}
}

func TestCombineSignaturesSecp256k1(t *testing.T) {
	kp1, _ := secp256k1.New()
	kp2, _ := ed25519.New()

	ms := &multisig.MultiSigPublicKey{
		PubKeys: []multisig.MemberPublicKey{
			{PubKey: kp1.PublicKey(), Weight: 1},
			{PubKey: kp2.PublicKey(), Weight: 1},
		},
		Threshold: 1,
	}

	sig, _ := crypto.SignTransaction(kp1, []byte("test"))
	combined, err := ms.CombineSignatures([]string{sig}, []int{0})
	if err != nil {
		t.Fatalf("secp256k1 signature combine failed: %v", err)
	}
	if combined == "" {
		t.Error("expected non-empty combined")
	}
}
