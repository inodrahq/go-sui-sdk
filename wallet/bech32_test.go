package wallet

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/inodrahq/go-sui-sdk/crypto"
	"github.com/inodrahq/go-sui-sdk/crypto/ed25519"
	"github.com/inodrahq/go-sui-sdk/crypto/secp256k1"
	"github.com/inodrahq/go-sui-sdk/crypto/secp256r1"
)

// Golden vectors verified against the PHP SDK.
func TestFromPrivateKeyEd25519(t *testing.T) {
	w, err := FromPrivateKey("suiprivkey1qp3lh8yeq3ypuv8sllt9ugmzu9c5t8xqyy3e0u0v9eghrrcsd26lsv86lfh")
	if err != nil {
		t.Fatal(err)
	}
	if w.Scheme() != crypto.Ed25519Scheme {
		t.Errorf("expected Ed25519, got %s", w.Scheme())
	}
	if w.Address() != "0xb443907dcff69d5f287fca41ace800b2e55e4100da0d214482bbcc60c46ce69b" {
		t.Errorf("address mismatch: %s", w.Address())
	}
	pubHex := hex.EncodeToString(w.Keypair().PublicKey().Bytes())
	if pubHex != "61bb800fe8ed296b7b15e8ffc4e06ba1aef9cb848b8468f0fa194927df4ea581" {
		t.Errorf("pubkey mismatch: %s", pubHex)
	}
}

func TestFromPrivateKeyEd25519Second(t *testing.T) {
	w, err := FromPrivateKey("suiprivkey1qqw6yeljc07m9vn4nlfsnsy7h7m0fzdckrlnl677dacycs7yjldlwn90fcg")
	if err != nil {
		t.Fatal(err)
	}
	if w.Address() != "0xa47ade8eae2ed6f689d709112bba4637fd0c8c82aea60b32a43c676e5b034481" {
		t.Errorf("address mismatch: %s", w.Address())
	}
	pubHex := hex.EncodeToString(w.Keypair().PublicKey().Bytes())
	if pubHex != "79c91460013eae285372871624c45802e56ef99802f36a43c1b81ea533dc3bda" {
		t.Errorf("pubkey mismatch: %s", pubHex)
	}
}

func TestFromPrivateKeyRoundtripEd25519(t *testing.T) {
	// Create from known key, export, reimport
	w1, err := FromPrivateKey("suiprivkey1qp3lh8yeq3ypuv8sllt9ugmzu9c5t8xqyy3e0u0v9eghrrcsd26lsv86lfh")
	if err != nil {
		t.Fatal(err)
	}
	exported, err := w1.PrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	if exported != "suiprivkey1qp3lh8yeq3ypuv8sllt9ugmzu9c5t8xqyy3e0u0v9eghrrcsd26lsv86lfh" {
		t.Errorf("exported key mismatch: %s", exported)
	}
	w2, err := FromPrivateKey(exported)
	if err != nil {
		t.Fatal(err)
	}
	if w1.Address() != w2.Address() {
		t.Errorf("address mismatch after roundtrip: %s != %s", w1.Address(), w2.Address())
	}
}

func TestFromPrivateKeyRoundtripNewWallet(t *testing.T) {
	// Generate new wallet, export, reimport
	w1, err := New()
	if err != nil {
		t.Fatal(err)
	}
	exported, err := w1.PrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	w2, err := FromPrivateKey(exported)
	if err != nil {
		t.Fatal(err)
	}
	if w1.Address() != w2.Address() {
		t.Errorf("address mismatch: %s != %s", w1.Address(), w2.Address())
	}
	if !bytes.Equal(w1.Keypair().PublicKey().Bytes(), w2.Keypair().PublicKey().Bytes()) {
		t.Error("public keys differ after roundtrip")
	}
}

func TestFromPrivateKeyRoundtripSecp256k1(t *testing.T) {
	kp, err := secp256k1.New()
	if err != nil {
		t.Fatal(err)
	}
	w1 := FromKeypair(kp)
	exported, err := w1.PrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	w2, err := FromPrivateKey(exported)
	if err != nil {
		t.Fatal(err)
	}
	if w1.Address() != w2.Address() {
		t.Errorf("address mismatch: %s != %s", w1.Address(), w2.Address())
	}
	if w2.Scheme() != crypto.Secp256k1Scheme {
		t.Errorf("expected Secp256k1, got %s", w2.Scheme())
	}
}

func TestFromPrivateKeyRoundtripSecp256r1(t *testing.T) {
	kp, err := secp256r1.New()
	if err != nil {
		t.Fatal(err)
	}
	w1 := FromKeypair(kp)
	exported, err := w1.PrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	w2, err := FromPrivateKey(exported)
	if err != nil {
		t.Fatal(err)
	}
	if w1.Address() != w2.Address() {
		t.Errorf("address mismatch: %s != %s", w1.Address(), w2.Address())
	}
	if w2.Scheme() != crypto.Secp256r1Scheme {
		t.Errorf("expected Secp256r1, got %s", w2.Scheme())
	}
}

func TestFromPrivateKeySeedPreservation(t *testing.T) {
	// Verify the seed bytes survive the roundtrip
	seed := make([]byte, 32)
	for i := range seed {
		seed[i] = byte(i)
	}
	kp, err := ed25519.FromSeed(seed)
	if err != nil {
		t.Fatal(err)
	}
	w1 := FromKeypair(kp)
	exported, err := w1.PrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	w2, err := FromPrivateKey(exported)
	if err != nil {
		t.Fatal(err)
	}
	seed2 := w2.Keypair().Seed()
	if !bytes.Equal(seed, seed2) {
		t.Errorf("seed mismatch:\n  got:  %x\n  want: %x", seed2, seed)
	}
}

func TestFromPrivateKeyInvalidHRP(t *testing.T) {
	// Valid bech32 but wrong HRP
	_, err := FromPrivateKey("bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t4")
	if err == nil {
		t.Fatal("expected error for wrong HRP")
	}
}

func TestFromPrivateKeyInvalidBech32(t *testing.T) {
	_, err := FromPrivateKey("suiprivkey1notvalidbech32data!!!")
	if err == nil {
		t.Fatal("expected error for invalid bech32")
	}
}

func TestFromPrivateKeyEmptyString(t *testing.T) {
	_, err := FromPrivateKey("")
	if err == nil {
		t.Fatal("expected error for empty string")
	}
}

func TestFromPrivateKeyNoSeparator(t *testing.T) {
	_, err := FromPrivateKey("noseparatorhere")
	if err == nil {
		t.Fatal("expected error for no separator")
	}
}

func TestPrivateKeyExportFormat(t *testing.T) {
	w, err := FromPrivateKey("suiprivkey1qp3lh8yeq3ypuv8sllt9ugmzu9c5t8xqyy3e0u0v9eghrrcsd26lsv86lfh")
	if err != nil {
		t.Fatal(err)
	}
	pk, err := w.PrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	// Must start with suiprivkey1
	if len(pk) < 11 || pk[:11] != "suiprivkey1" {
		t.Errorf("invalid format: %s", pk)
	}
}

func TestBech32EncodeDecodeRoundtrip(t *testing.T) {
	data := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}
	encoded, err := bech32Encode("test", data)
	if err != nil {
		t.Fatal(err)
	}
	hrp, decoded, err := bech32Decode(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if hrp != "test" {
		t.Errorf("HRP mismatch: %s", hrp)
	}
	if !bytes.Equal(data, decoded) {
		t.Errorf("data mismatch:\n  got:  %x\n  want: %x", decoded, data)
	}
}

func TestBech32DecodeInvalidChecksum(t *testing.T) {
	// Take a valid encoding and corrupt the last character
	data := []byte{0x00}
	encoded, err := bech32Encode("test", data)
	if err != nil {
		t.Fatal(err)
	}
	// Flip last char
	corrupted := encoded[:len(encoded)-1] + "q"
	if corrupted == encoded {
		corrupted = encoded[:len(encoded)-1] + "p"
	}
	_, _, err = bech32Decode(corrupted)
	if err == nil {
		t.Fatal("expected checksum error")
	}
}

func TestBech32RoundtripVariousLengths(t *testing.T) {
	for _, n := range []int{1, 16, 20, 32, 33, 64} {
		data := make([]byte, n)
		for i := range data {
			data[i] = byte(i * 7)
		}
		encoded, err := bech32Encode("sui", data)
		if err != nil {
			t.Fatalf("encode len=%d: %v", n, err)
		}
		_, decoded, err := bech32Decode(encoded)
		if err != nil {
			t.Fatalf("decode len=%d: %v", n, err)
		}
		if !bytes.Equal(data, decoded) {
			t.Fatalf("roundtrip failed for len=%d", n)
		}
	}
}

func TestFromPrivateKeyAllSchemesFromSeed(t *testing.T) {
	seed := make([]byte, 32)
	for i := range seed {
		seed[i] = byte(i + 42)
	}

	tests := []struct {
		name   string
		scheme crypto.SignatureScheme
		flag   byte
	}{
		{"Ed25519", crypto.Ed25519Scheme, 0x00},
		{"Secp256k1", crypto.Secp256k1Scheme, 0x01},
		{"Secp256r1", crypto.Secp256r1Scheme, 0x02},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Manually encode seed with flag
			data := make([]byte, 33)
			data[0] = tt.flag
			copy(data[1:], seed)
			encoded, err := bech32Encode(hrpPrivateKey, data)
			if err != nil {
				t.Fatal(err)
			}
			w, err := FromPrivateKey(encoded)
			if err != nil {
				t.Fatal(err)
			}
			if w.Scheme() != tt.scheme {
				t.Errorf("expected %s, got %s", tt.scheme, w.Scheme())
			}
			// Verify seed matches
			gotSeed := w.Keypair().Seed()
			if !bytes.Equal(seed, gotSeed) {
				t.Errorf("seed mismatch")
			}
			// Export and reimport
			exported, err := w.PrivateKey()
			if err != nil {
				t.Fatal(err)
			}
			w2, err := FromPrivateKey(exported)
			if err != nil {
				t.Fatal(err)
			}
			if w.Address() != w2.Address() {
				t.Errorf("address mismatch after roundtrip")
			}
		})
	}
}

func TestFromPrivateKeyUnsupportedScheme(t *testing.T) {
	// Manually encode with flag 0x05 (zkLogin — not supported for import)
	data := make([]byte, 33)
	data[0] = 0x05
	encoded, err := bech32Encode(hrpPrivateKey, data)
	if err != nil {
		t.Fatal(err)
	}
	_, err = FromPrivateKey(encoded)
	if err == nil {
		t.Fatal("expected error for unsupported scheme")
	}
}

func TestFromPrivateKeyTooShort(t *testing.T) {
	// Encode just a flag byte with no seed
	data := []byte{0x00}
	encoded, err := bech32Encode(hrpPrivateKey, data)
	if err != nil {
		t.Fatal(err)
	}
	_, err = FromPrivateKey(encoded)
	if err == nil {
		t.Fatal("expected error for too-short key")
	}
}

func TestMultipleWalletsUniqueAddresses(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 10; i++ {
		w, err := New()
		if err != nil {
			t.Fatal(err)
		}
		pk, err := w.PrivateKey()
		if err != nil {
			t.Fatal(err)
		}
		w2, err := FromPrivateKey(pk)
		if err != nil {
			t.Fatal(err)
		}
		if w.Address() != w2.Address() {
			t.Fatalf("roundtrip %d failed", i)
		}
		if seen[w.Address()] {
			t.Fatalf("duplicate address at iteration %d", i)
		}
		seen[w.Address()] = true
	}
}
