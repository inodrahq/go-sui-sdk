package crypto_test

import (
	"testing"

	"github.com/inodrahq/go-sui-sdk/crypto"
)

func TestSignatureSchemeString(t *testing.T) {
	tests := []struct {
		scheme crypto.SignatureScheme
		name   string
	}{
		{crypto.Ed25519Scheme, "Ed25519"},
		{crypto.Secp256k1Scheme, "Secp256k1"},
		{crypto.Secp256r1Scheme, "Secp256r1"},
		{crypto.MultiSigScheme, "MultiSig"},
		{crypto.ZkLoginScheme, "zkLogin"},
		{crypto.PasskeyScheme, "Passkey"},
		{crypto.SignatureScheme(0xFF), "Unknown"},
	}
	for _, tt := range tests {
		if tt.scheme.String() != tt.name {
			t.Errorf("expected %s, got %s", tt.name, tt.scheme.String())
		}
	}
}

func TestIntentPrefix(t *testing.T) {
	txPrefix := crypto.IntentPrefix(crypto.IntentTransactionData)
	if txPrefix != [3]byte{0, 0, 0} {
		t.Errorf("unexpected tx intent prefix: %v", txPrefix)
	}

	msgPrefix := crypto.IntentPrefix(crypto.IntentPersonalMessage)
	if msgPrefix != [3]byte{3, 0, 0} {
		t.Errorf("unexpected personal message intent prefix: %v", msgPrefix)
	}
}
