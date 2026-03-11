package ed25519_test

import (
	"testing"

	"github.com/inodrahq/go-sui-sdk/crypto"
	"github.com/inodrahq/go-sui-sdk/crypto/ed25519"
)

func TestNewKeypair(t *testing.T) {
	kp, err := ed25519.New()
	if err != nil {
		t.Fatal(err)
	}
	if kp.Scheme() != crypto.Ed25519Scheme {
		t.Error("expected Ed25519 scheme")
	}
	pk := kp.PublicKey()
	if len(pk.Bytes()) != 32 {
		t.Errorf("expected 32-byte public key, got %d", len(pk.Bytes()))
	}
	if pk.Flag() != 0x00 {
		t.Errorf("expected flag 0x00, got %02x", pk.Flag())
	}
	if pk.Scheme() != crypto.Ed25519Scheme {
		t.Error("expected Ed25519 scheme on public key")
	}
}

func TestFromSeed(t *testing.T) {
	seed := make([]byte, 32)
	for i := range seed {
		seed[i] = byte(i)
	}
	kp, err := ed25519.FromSeed(seed)
	if err != nil {
		t.Fatal(err)
	}
	if len(kp.Seed()) != 32 {
		t.Error("expected 32-byte seed")
	}
	for i, b := range kp.Seed() {
		if b != byte(i) {
			t.Errorf("seed mismatch at %d", i)
			break
		}
	}
}

func TestFromSeedInvalidLength(t *testing.T) {
	_, err := ed25519.FromSeed(make([]byte, 16))
	if err == nil {
		t.Error("expected error for wrong seed length")
	}
}

func TestFromPrivateKeyInvalidLength(t *testing.T) {
	_, err := ed25519.FromPrivateKey(make([]byte, 16))
	if err == nil {
		t.Error("expected error for wrong private key length")
	}
}

func TestSignAndVerify(t *testing.T) {
	kp, err := ed25519.New()
	if err != nil {
		t.Fatal(err)
	}
	msg := []byte("hello sui")
	sig, err := kp.Sign(msg)
	if err != nil {
		t.Fatal(err)
	}
	if len(sig) != 64 {
		t.Errorf("expected 64-byte signature, got %d", len(sig))
	}
	pk := kp.PublicKey()
	if !pk.Verify(msg, sig) {
		t.Error("signature verification failed")
	}
	// Wrong message should fail
	if pk.Verify([]byte("wrong"), sig) {
		t.Error("expected verification to fail for wrong message")
	}
}

func TestVerifyInvalidSigLength(t *testing.T) {
	kp, _ := ed25519.New()
	pk := kp.PublicKey()
	if pk.Verify([]byte("test"), []byte("short")) {
		t.Error("expected false for invalid sig length")
	}
}

func TestSuiAddressFormat(t *testing.T) {
	kp, _ := ed25519.New()
	addr := kp.PublicKey().SuiAddress()
	if len(addr) != 66 { // "0x" + 64 hex chars
		t.Errorf("expected 66-char address, got %d: %s", len(addr), addr)
	}
	if addr[:2] != "0x" {
		t.Error("address should start with 0x")
	}
}

func TestDeterministicAddress(t *testing.T) {
	seed := make([]byte, 32)
	kp1, _ := ed25519.FromSeed(seed)
	kp2, _ := ed25519.FromSeed(seed)
	if kp1.PublicKey().SuiAddress() != kp2.PublicKey().SuiAddress() {
		t.Error("same seed should produce same address")
	}
}
