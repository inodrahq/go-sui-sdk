package wallet_test

import (
	"testing"

	"github.com/inodrahq/go-sui-sdk/crypto"
	"github.com/inodrahq/go-sui-sdk/wallet"
)

// Sui TypeScript SDK test vectors (mnemonic → address)
func TestWalletFromMnemonicVectors(t *testing.T) {
	tests := []struct {
		name     string
		mnemonic string
		address  string
	}{
		{
			"vector1",
			"film crazy soon outside stand loop subway crumble thrive popular green nuclear struggle pistol arm wife phrase warfare march wheat nephew ask sunny firm",
			"0xa2d14fad60c56049ecf75246a481934691214ce413e6a8ae2fe6834c173a6133",
		},
		{
			"vector2",
			"require decline left thought grid priority false tiny gasp angle royal system attack beef setup reward aunt skill wasp tray vital bounce inflict level",
			"0x1ada6e6f3f3e4055096f606c746690f1108fcc2ca479055cc434a3e1d3f758aa",
		},
		{
			"vector3",
			"organ crash swim stick traffic remember army arctic mesh slice swear summer police vast chaos cradle squirrel hood useless evidence pet hub soap lake",
			"0xe69e896ca10f5a77732769803cc2b5707f0ab9d4407afb5e4b4464b89769af14",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, err := wallet.FromMnemonic(tt.mnemonic)
			if err != nil {
				t.Fatal(err)
			}
			if w.Address() != tt.address {
				t.Errorf("address mismatch:\n  got: %s\n  exp: %s", w.Address(), tt.address)
			}
		})
	}
}

func TestWalletNew(t *testing.T) {
	w, err := wallet.New()
	if err != nil {
		t.Fatal(err)
	}
	if w.Scheme() != crypto.Ed25519Scheme {
		t.Error("default scheme should be Ed25519")
	}
	addr := w.Address()
	if len(addr) != 66 {
		t.Errorf("expected 66-char address, got %d", len(addr))
	}
}

func TestWalletInvalidMnemonic(t *testing.T) {
	_, err := wallet.FromMnemonic("invalid mnemonic phrase")
	if err == nil {
		t.Error("expected error for invalid mnemonic")
	}
}

func TestWalletUnsupportedScheme(t *testing.T) {
	_, err := wallet.New(wallet.WithScheme(crypto.MultiSigScheme))
	if err == nil {
		t.Error("expected error for unsupported scheme")
	}
}

func TestWalletSecp256k1(t *testing.T) {
	w, err := wallet.New(wallet.WithScheme(crypto.Secp256k1Scheme))
	if err != nil {
		t.Fatal(err)
	}
	if w.Scheme() != crypto.Secp256k1Scheme {
		t.Error("expected Secp256k1")
	}
	if len(w.Address()) != 66 {
		t.Errorf("expected 66-char address, got %d", len(w.Address()))
	}
}

func TestWalletSecp256r1(t *testing.T) {
	w, err := wallet.New(wallet.WithScheme(crypto.Secp256r1Scheme))
	if err != nil {
		t.Fatal(err)
	}
	if w.Scheme() != crypto.Secp256r1Scheme {
		t.Error("expected Secp256r1")
	}
}

func TestWalletSignTransaction(t *testing.T) {
	w, _ := wallet.New()
	sig, err := w.SignTransaction([]byte("fake tx bytes"))
	if err != nil {
		t.Fatal(err)
	}
	if len(sig) == 0 {
		t.Error("expected non-empty signature")
	}
}

func TestWalletSignPersonalMessage(t *testing.T) {
	w, _ := wallet.New()
	sig, err := w.SignPersonalMessage([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if len(sig) == 0 {
		t.Error("expected non-empty signature")
	}
}

func TestWalletSignRaw(t *testing.T) {
	w, _ := wallet.New()
	sig, err := w.SignRaw([]byte("test"))
	if err != nil {
		t.Fatal(err)
	}
	// Ed25519: 1 (flag) + 64 (sig) + 32 (pk) = 97 bytes
	if len(sig) != 97 {
		t.Errorf("expected 97-byte raw signature, got %d", len(sig))
	}
	if sig[0] != 0x00 {
		t.Errorf("expected flag 0x00, got %02x", sig[0])
	}
}

func TestWalletSignRawBase64(t *testing.T) {
	w, _ := wallet.New()
	sig, err := w.SignRawBase64([]byte("test"))
	if err != nil {
		t.Fatal(err)
	}
	if len(sig) == 0 {
		t.Error("expected non-empty base64 signature")
	}
}

func TestWalletKeypair(t *testing.T) {
	w, _ := wallet.New()
	kp := w.Keypair()
	if kp == nil {
		t.Error("expected non-nil keypair")
	}
}

func TestWalletWithPath(t *testing.T) {
	m := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	w1, _ := wallet.FromMnemonic(m, wallet.WithPath("m/44'/784'/0'/0'/0'"))
	w2, _ := wallet.FromMnemonic(m, wallet.WithPath("m/44'/784'/0'/0'/1'"))
	if w1.Address() == w2.Address() {
		t.Error("different paths should produce different addresses")
	}
}

func TestWalletMnemonicUnsupportedScheme(t *testing.T) {
	m := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	_, err := wallet.FromMnemonic(m, wallet.WithScheme(crypto.MultiSigScheme))
	if err == nil {
		t.Error("expected error for unsupported scheme in mnemonic derivation")
	}
}

func TestWalletMnemonicSecp256k1(t *testing.T) {
	m := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	w, err := wallet.FromMnemonic(m, wallet.WithScheme(crypto.Secp256k1Scheme))
	if err != nil {
		t.Fatal(err)
	}
	if w.Scheme() != crypto.Secp256k1Scheme {
		t.Error("expected Secp256k1")
	}
}

func TestWalletMnemonicSecp256r1(t *testing.T) {
	m := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	w, err := wallet.FromMnemonic(m, wallet.WithScheme(crypto.Secp256r1Scheme))
	if err != nil {
		t.Fatal(err)
	}
	if w.Scheme() != crypto.Secp256r1Scheme {
		t.Error("expected Secp256r1")
	}
}
