package secp256k1_test

import (
	"testing"

	"github.com/inodrahq/go-sui-sdk/crypto"
	"github.com/inodrahq/go-sui-sdk/crypto/secp256k1"
)

func TestNewKeypair(t *testing.T) {
	kp, err := secp256k1.New()
	if err != nil {
		t.Fatal(err)
	}
	if kp.Scheme() != crypto.Secp256k1Scheme {
		t.Error("expected Secp256k1 scheme")
	}
	pk := kp.PublicKey()
	if len(pk.Bytes()) != 33 {
		t.Errorf("expected 33-byte compressed pubkey, got %d", len(pk.Bytes()))
	}
	if pk.Flag() != 0x01 {
		t.Errorf("expected flag 0x01, got %02x", pk.Flag())
	}
}

func TestFromPrivateKeyBytes(t *testing.T) {
	kp1, _ := secp256k1.New()
	privBytes := kp1.PrivateKeyBytes()
	kp2, err := secp256k1.FromPrivateKeyBytes(privBytes)
	if err != nil {
		t.Fatal(err)
	}
	if kp1.PublicKey().SuiAddress() != kp2.PublicKey().SuiAddress() {
		t.Error("same private key should produce same address")
	}
}

func TestFromPrivateKeyBytesInvalidLength(t *testing.T) {
	_, err := secp256k1.FromPrivateKeyBytes(make([]byte, 16))
	if err == nil {
		t.Error("expected error for wrong key length")
	}
}

func TestSignAndVerify(t *testing.T) {
	kp, _ := secp256k1.New()
	hash := make([]byte, 32)
	hash[0] = 0x42
	sig, err := kp.Sign(hash)
	if err != nil {
		t.Fatal(err)
	}
	if len(sig) != 64 {
		t.Errorf("expected 64-byte signature, got %d", len(sig))
	}
	pk := kp.PublicKey()
	if !pk.Verify(hash, sig) {
		t.Error("signature verification failed")
	}
	// Wrong hash
	wrongHash := make([]byte, 32)
	wrongHash[0] = 0xFF
	if pk.Verify(wrongHash, sig) {
		t.Error("expected verification to fail for wrong hash")
	}
}

func TestVerifyInvalidSigLength(t *testing.T) {
	kp, _ := secp256k1.New()
	pk := kp.PublicKey()
	if pk.Verify(make([]byte, 32), []byte("short")) {
		t.Error("expected false for invalid sig length")
	}
}

func TestSuiAddressFormat(t *testing.T) {
	kp, _ := secp256k1.New()
	addr := kp.PublicKey().SuiAddress()
	if len(addr) != 66 {
		t.Errorf("expected 66-char address, got %d", len(addr))
	}
	if addr[:2] != "0x" {
		t.Error("address should start with 0x")
	}
}

func TestPublicKeyFromBytes(t *testing.T) {
	kp, _ := secp256k1.New()
	bytes := kp.PublicKey().Bytes()
	pk, err := secp256k1.NewPublicKeyFromBytes(bytes)
	if err != nil {
		t.Fatal(err)
	}
	if pk.SuiAddress() != kp.PublicKey().SuiAddress() {
		t.Error("parsed public key should have same address")
	}
}

func TestPublicKeyFromBytesInvalid(t *testing.T) {
	_, err := secp256k1.NewPublicKeyFromBytes([]byte{0xFF, 0xFF})
	if err == nil {
		t.Error("expected error for invalid public key bytes")
	}
}

func TestSignatureSize(t *testing.T) {
	// Sui secp256k1 signature: flag(1) + sig(64) + pk(33) = 98 bytes
	kp, _ := secp256k1.New()
	hash := make([]byte, 32)
	sig, _ := kp.Sign(hash)
	pk := kp.PublicKey()
	suiSig := make([]byte, 1+len(sig)+len(pk.Bytes()))
	suiSig[0] = pk.Flag()
	copy(suiSig[1:], sig)
	copy(suiSig[1+len(sig):], pk.Bytes())
	if len(suiSig) != 98 {
		t.Errorf("expected 98-byte Sui signature, got %d", len(suiSig))
	}
}
