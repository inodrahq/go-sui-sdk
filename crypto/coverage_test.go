package crypto_test

import (
	"encoding/base64"
	"testing"

	"github.com/inodrahq/go-sui-sdk/crypto"
	"github.com/inodrahq/go-sui-sdk/crypto/ed25519"
)

// --- SignatureScheme constants ---

func TestSignatureSchemeValues(t *testing.T) {
	if crypto.Ed25519Scheme != 0 {
		t.Errorf("Ed25519Scheme: expected 0, got %d", crypto.Ed25519Scheme)
	}
	if crypto.Secp256k1Scheme != 1 {
		t.Errorf("Secp256k1Scheme: expected 1, got %d", crypto.Secp256k1Scheme)
	}
	if crypto.Secp256r1Scheme != 2 {
		t.Errorf("Secp256r1Scheme: expected 2, got %d", crypto.Secp256r1Scheme)
	}
	if crypto.MultiSigScheme != 3 {
		t.Errorf("MultiSigScheme: expected 3, got %d", crypto.MultiSigScheme)
	}
	if crypto.ZkLoginScheme != 5 {
		t.Errorf("ZkLoginScheme: expected 5, got %d", crypto.ZkLoginScheme)
	}
	if crypto.PasskeyScheme != 6 {
		t.Errorf("PasskeyScheme: expected 6, got %d", crypto.PasskeyScheme)
	}
}

func TestSignatureSchemeStringUnknown(t *testing.T) {
	unknown := crypto.SignatureScheme(0x04)
	if unknown.String() != "Unknown" {
		t.Errorf("expected Unknown, got %s", unknown.String())
	}
	unknown2 := crypto.SignatureScheme(0xFF)
	if unknown2.String() != "Unknown" {
		t.Errorf("expected Unknown, got %s", unknown2.String())
	}
}

// --- IntentPrefix ---

func TestIntentPrefixCustomScope(t *testing.T) {
	// IntentTransactionData (scope=0)
	p := crypto.IntentPrefix(crypto.IntentTransactionData)
	if p != [3]byte{0, 0, 0} {
		t.Errorf("unexpected tx prefix: %v", p)
	}

	// IntentPersonalMessage (scope=3)
	p = crypto.IntentPrefix(crypto.IntentPersonalMessage)
	if p != [3]byte{3, 0, 0} {
		t.Errorf("unexpected personal msg prefix: %v", p)
	}

	// Arbitrary scope value
	p = crypto.IntentPrefix(crypto.IntentScope(7))
	if p != [3]byte{7, 0, 0} {
		t.Errorf("unexpected custom scope prefix: %v", p)
	}
}

// --- DeriveAddress ---

func TestDeriveAddressFormat(t *testing.T) {
	pubkey := make([]byte, 32)
	addr := crypto.DeriveAddress(0x00, pubkey)

	// Must be 0x + 64 hex chars = 66 chars total
	if len(addr) != 66 {
		t.Errorf("expected 66-char address, got %d", len(addr))
	}
	if addr[:2] != "0x" {
		t.Error("address should start with 0x")
	}
}

func TestDeriveAddressDifferentFlags(t *testing.T) {
	pubkey := make([]byte, 32)
	addr0 := crypto.DeriveAddress(0x00, pubkey)
	addr1 := crypto.DeriveAddress(0x01, pubkey)
	addr2 := crypto.DeriveAddress(0x02, pubkey)

	// Different flags must produce different addresses for the same pubkey
	if addr0 == addr1 {
		t.Error("Ed25519 and Secp256k1 flags should produce different addresses")
	}
	if addr0 == addr2 {
		t.Error("Ed25519 and Secp256r1 flags should produce different addresses")
	}
	if addr1 == addr2 {
		t.Error("Secp256k1 and Secp256r1 flags should produce different addresses")
	}
}

func TestDeriveAddressDifferentKeys(t *testing.T) {
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	key2[0] = 0x01

	addr1 := crypto.DeriveAddress(0x00, key1)
	addr2 := crypto.DeriveAddress(0x00, key2)

	if addr1 == addr2 {
		t.Error("different public keys should produce different addresses")
	}
}

// --- NormalizeAddress ---

func TestNormalizeAddressShortHex(t *testing.T) {
	got, err := crypto.NormalizeAddress("0x2")
	if err != nil {
		t.Fatal(err)
	}
	expected := "0x0000000000000000000000000000000000000000000000000000000000000002"
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestNormalizeAddressNoPrefix(t *testing.T) {
	got, err := crypto.NormalizeAddress("abc")
	if err != nil {
		t.Fatal(err)
	}
	expected := "0x0000000000000000000000000000000000000000000000000000000000000abc"
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestNormalizeAddressUpperCase(t *testing.T) {
	got, err := crypto.NormalizeAddress("0xABCDEF")
	if err != nil {
		t.Fatal(err)
	}
	expected := "0x0000000000000000000000000000000000000000000000000000000000abcdef"
	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestNormalizeAddressFullLength(t *testing.T) {
	full := "0x" + "aa" + "00000000000000000000000000000000000000000000000000000000000000"
	got, err := crypto.NormalizeAddress(full)
	if err != nil {
		t.Fatal(err)
	}
	if got != full {
		t.Errorf("expected %s, got %s", full, got)
	}
}

func TestNormalizeAddressTooLong(t *testing.T) {
	tooLong := "0x" + "a" + "0000000000000000000000000000000000000000000000000000000000000000"
	_, err := crypto.NormalizeAddress(tooLong)
	if err == nil {
		t.Error("expected error for too-long address")
	}
}

func TestNormalizeAddressBadHex(t *testing.T) {
	_, err := crypto.NormalizeAddress("0xZZZZ")
	if err == nil {
		t.Error("expected error for invalid hex")
	}
}

// --- ValidateAddress ---

func TestValidateAddressValid(t *testing.T) {
	valids := []string{
		"0x0000000000000000000000000000000000000000000000000000000000000001",
		"0x1",
		"0x2",
		"0xabcdef",
		"ff",
	}
	for _, addr := range valids {
		if !crypto.ValidateAddress(addr) {
			t.Errorf("expected %q to be valid", addr)
		}
	}
}

func TestValidateAddressInvalid(t *testing.T) {
	invalids := []string{
		"0xZZZZ",
		"0xGGGG",
		"0x" + "a" + "0000000000000000000000000000000000000000000000000000000000000000", // 65 hex chars = too long
	}
	for _, addr := range invalids {
		if crypto.ValidateAddress(addr) {
			t.Errorf("expected %q to be invalid", addr)
		}
	}
}

// --- SignTransaction and SignPersonalMessage ---

func TestSignTransaction(t *testing.T) {
	seed := make([]byte, 32)
	seed[0] = 0x42
	kp, err := ed25519.FromSeed(seed)
	if err != nil {
		t.Fatal(err)
	}

	txBytes := []byte("test transaction data")
	sig, err := crypto.SignTransaction(kp, txBytes)
	if err != nil {
		t.Fatal(err)
	}

	// Decode the base64 signature
	raw, err := base64.StdEncoding.DecodeString(sig)
	if err != nil {
		t.Fatalf("invalid base64: %v", err)
	}

	// Sui signature format: flag(1) + ed25519_sig(64) + pubkey(32) = 97 bytes
	if len(raw) != 97 {
		t.Fatalf("expected 97 bytes, got %d", len(raw))
	}

	// First byte is flag = 0x00 for Ed25519
	if raw[0] != 0x00 {
		t.Errorf("expected flag 0x00, got 0x%02x", raw[0])
	}

	// Last 32 bytes should be the public key
	pk := kp.PublicKey()
	pkBytes := pk.Bytes()
	for i := 0; i < 32; i++ {
		if raw[65+i] != pkBytes[i] {
			t.Errorf("pubkey mismatch at byte %d: expected 0x%02x, got 0x%02x", i, pkBytes[i], raw[65+i])
			break
		}
	}
}

func TestSignPersonalMessage(t *testing.T) {
	seed := make([]byte, 32)
	seed[0] = 0x42
	kp, err := ed25519.FromSeed(seed)
	if err != nil {
		t.Fatal(err)
	}

	msg := []byte("hello world")
	sig, err := crypto.SignPersonalMessage(kp, msg)
	if err != nil {
		t.Fatal(err)
	}

	raw, err := base64.StdEncoding.DecodeString(sig)
	if err != nil {
		t.Fatalf("invalid base64: %v", err)
	}

	// 97 bytes for Ed25519
	if len(raw) != 97 {
		t.Fatalf("expected 97 bytes, got %d", len(raw))
	}

	if raw[0] != 0x00 {
		t.Errorf("expected flag 0x00, got 0x%02x", raw[0])
	}
}

func TestSignTransactionAndPersonalMessageDiffer(t *testing.T) {
	seed := make([]byte, 32)
	seed[0] = 0x42
	kp, err := ed25519.FromSeed(seed)
	if err != nil {
		t.Fatal(err)
	}

	data := []byte("same data")
	txSig, err := crypto.SignTransaction(kp, data)
	if err != nil {
		t.Fatal(err)
	}

	msgSig, err := crypto.SignPersonalMessage(kp, data)
	if err != nil {
		t.Fatal(err)
	}

	// Different intents should produce different signatures for the same data
	if txSig == msgSig {
		t.Error("transaction and personal message signatures should differ for the same data")
	}
}

func TestSignTransactionDeterministic(t *testing.T) {
	seed := make([]byte, 32)
	seed[0] = 0x42
	kp, err := ed25519.FromSeed(seed)
	if err != nil {
		t.Fatal(err)
	}

	txBytes := []byte("deterministic test")
	sig1, err := crypto.SignTransaction(kp, txBytes)
	if err != nil {
		t.Fatal(err)
	}
	sig2, err := crypto.SignTransaction(kp, txBytes)
	if err != nil {
		t.Fatal(err)
	}

	// Ed25519 signatures are deterministic
	if sig1 != sig2 {
		t.Error("signing the same data twice should produce the same signature")
	}
}

// --- Error types ---

func TestErrorMessages(t *testing.T) {
	if crypto.ErrInvalidPublicKey.Error() != "crypto: invalid public key" {
		t.Errorf("unexpected error message: %s", crypto.ErrInvalidPublicKey.Error())
	}
	if crypto.ErrInvalidSignature.Error() != "crypto: invalid signature" {
		t.Errorf("unexpected error message: %s", crypto.ErrInvalidSignature.Error())
	}
	if crypto.ErrUnsupportedScheme.Error() != "crypto: unsupported signature scheme" {
		t.Errorf("unexpected error message: %s", crypto.ErrUnsupportedScheme.Error())
	}
}
