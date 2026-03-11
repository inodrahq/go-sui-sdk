package passkey_test

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/inodrahq/go-sui-sdk/crypto"
	"github.com/inodrahq/go-sui-sdk/crypto/passkey"
	"github.com/inodrahq/go-sui-sdk/crypto/secp256r1"
)

func TestDeriveAddress(t *testing.T) {
	kp, _ := secp256r1.New()
	addr, err := passkey.DeriveAddress(kp.PublicKey().Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(addr, "0x") {
		t.Error("expected 0x prefix")
	}
	if len(addr) != 66 {
		t.Errorf("expected 66-char address, got %d", len(addr))
	}
}

func TestDeriveAddressDiffersFromSecp256r1(t *testing.T) {
	kp, _ := secp256r1.New()
	passkeyAddr, _ := passkey.DeriveAddress(kp.PublicKey().Bytes())
	secp256r1Addr := kp.PublicKey().SuiAddress()
	if passkeyAddr == secp256r1Addr {
		t.Error("passkey and secp256r1 addresses should differ (different flags)")
	}
}

func TestDeriveAddressDeterministic(t *testing.T) {
	kp, _ := secp256r1.New()
	addr1, _ := passkey.DeriveAddress(kp.PublicKey().Bytes())
	addr2, _ := passkey.DeriveAddress(kp.PublicKey().Bytes())
	if addr1 != addr2 {
		t.Error("same key should produce same address")
	}
}

func TestDeriveAddressRejectsWrongSize(t *testing.T) {
	_, err := passkey.DeriveAddress(make([]byte, 32))
	if err == nil {
		t.Error("expected error for 32-byte key")
	}
}

func TestAssembleAuthenticator(t *testing.T) {
	kp, _ := secp256r1.New()
	sig, _ := crypto.SignTransaction(kp, []byte("test transaction"))

	authenticatorData := make([]byte, 37)
	clientDataJSON := `{"type":"webauthn.get","challenge":"test","origin":"https://example.com"}`

	result, err := passkey.Assemble(authenticatorData, clientDataJSON, sig)
	if err != nil {
		t.Fatal(err)
	}

	decoded, err := base64.StdEncoding.DecodeString(result)
	if err != nil {
		t.Fatal(err)
	}
	if decoded[0] != 0x06 {
		t.Errorf("expected passkey flag 0x06, got 0x%02x", decoded[0])
	}
}

func TestAssembleRejectsInvalidSignature(t *testing.T) {
	_, err := passkey.Assemble([]byte("auth-data"), "client-json", "!!!invalid!!!")
	if err == nil {
		t.Error("expected error for invalid base64")
	}
}
