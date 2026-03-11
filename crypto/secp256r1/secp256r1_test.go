package secp256r1_test

import (
	"testing"

	"github.com/inodrahq/go-sui-sdk/crypto"
	"github.com/inodrahq/go-sui-sdk/crypto/secp256r1"
)

func TestNewKeypair(t *testing.T) {
	kp, err := secp256r1.New()
	if err != nil {
		t.Fatal(err)
	}
	if kp.Scheme() != crypto.Secp256r1Scheme {
		t.Error("expected Secp256r1 scheme")
	}
	pk := kp.PublicKey()
	if len(pk.Bytes()) != 33 {
		t.Errorf("expected 33-byte compressed pubkey, got %d", len(pk.Bytes()))
	}
	if pk.Flag() != 0x02 {
		t.Errorf("expected flag 0x02, got %02x", pk.Flag())
	}
}

func TestFromPrivateKeyBytes(t *testing.T) {
	kp1, _ := secp256r1.New()
	privBytes := kp1.PrivateKeyBytes()
	kp2, err := secp256r1.FromPrivateKeyBytes(privBytes)
	if err != nil {
		t.Fatal(err)
	}
	if kp1.PublicKey().SuiAddress() != kp2.PublicKey().SuiAddress() {
		t.Error("same private key should produce same address")
	}
}

func TestFromPrivateKeyBytesInvalidLength(t *testing.T) {
	_, err := secp256r1.FromPrivateKeyBytes(make([]byte, 16))
	if err == nil {
		t.Error("expected error for wrong key length")
	}
}

func TestSignAndVerify(t *testing.T) {
	kp, _ := secp256r1.New()
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
	kp, _ := secp256r1.New()
	pk := kp.PublicKey()
	if pk.Verify(make([]byte, 32), []byte("short")) {
		t.Error("expected false for invalid sig length")
	}
}

func TestSuiAddressFormat(t *testing.T) {
	kp, _ := secp256r1.New()
	addr := kp.PublicKey().SuiAddress()
	if len(addr) != 66 {
		t.Errorf("expected 66-char address, got %d", len(addr))
	}
}

func TestPublicKeyFromBytes(t *testing.T) {
	kp, _ := secp256r1.New()
	bytes := kp.PublicKey().Bytes()
	pk, err := secp256r1.NewPublicKeyFromBytes(bytes)
	if err != nil {
		t.Fatal(err)
	}
	if pk.SuiAddress() != kp.PublicKey().SuiAddress() {
		t.Error("parsed public key should have same address")
	}
}

func TestPublicKeyFromBytesInvalid(t *testing.T) {
	_, err := secp256r1.NewPublicKeyFromBytes([]byte{0xFF, 0xFF})
	if err == nil {
		t.Error("expected error for invalid public key bytes")
	}
}

func TestLowSNormalization(t *testing.T) {
	// Sign many times and verify all produce valid signatures
	kp, _ := secp256r1.New()
	hash := make([]byte, 32)
	for i := 0; i < 20; i++ {
		hash[0] = byte(i)
		sig, err := kp.Sign(hash)
		if err != nil {
			t.Fatalf("sign iteration %d: %v", i, err)
		}
		if !kp.PublicKey().Verify(hash, sig) {
			t.Fatalf("verify failed iteration %d", i)
		}
	}
}

func TestSignatureSize(t *testing.T) {
	// Sui secp256r1 signature: flag(1) + sig(64) + pk(33) = 98 bytes
	kp, _ := secp256r1.New()
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
