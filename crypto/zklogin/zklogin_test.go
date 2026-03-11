package zklogin_test

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/inodrahq/go-sui-sdk/crypto/zklogin"
)

func TestPoseidonHash1(t *testing.T) {
	result, err := zklogin.PoseidonHash([]string{"1"})
	if err != nil {
		t.Fatal(err)
	}
	expected := "18586133768512220936620570745912940619677854269274689475585506675881198879027"
	if result != expected {
		t.Errorf("got %s, want %s", result, expected)
	}
}

func TestPoseidonHash2(t *testing.T) {
	result, err := zklogin.PoseidonHash([]string{"1", "2"})
	if err != nil {
		t.Fatal(err)
	}
	expected := "7853200120776062878684798364095072458815029376092732009249414926327459813530"
	if result != expected {
		t.Errorf("got %s, want %s", result, expected)
	}
}

func TestPoseidonHash4(t *testing.T) {
	result, err := zklogin.PoseidonHash([]string{"1", "2", "3", "4"})
	if err != nil {
		t.Fatal(err)
	}
	expected := "18821383157269793795438455681495246036402687001665670618754263018637548127333"
	if result != expected {
		t.Errorf("got %s, want %s", result, expected)
	}
}

func TestPoseidonHashLargeInput(t *testing.T) {
	result, err := zklogin.PoseidonHash([]string{"12345678901234567890"})
	if err != nil {
		t.Fatal(err)
	}
	expected := "17610922722311195426938483481431943255028223790571250909270476711880232282197"
	if result != expected {
		t.Errorf("got %s, want %s", result, expected)
	}
}

func TestPoseidonRejectsEmpty(t *testing.T) {
	_, err := zklogin.PoseidonHash([]string{})
	if err == nil {
		t.Error("expected error for empty inputs")
	}
}

func TestPoseidonRejectsTooMany(t *testing.T) {
	inputs := make([]string, 17)
	for i := range inputs {
		inputs[i] = "1"
	}
	_, err := zklogin.PoseidonHash(inputs)
	if err == nil {
		t.Error("expected error for 17 inputs")
	}
}

func TestHashASCIIStrToField(t *testing.T) {
	result, err := zklogin.HashASCIIStrToField("sub")
	if err != nil {
		t.Fatal(err)
	}
	expected := "15390800048627689072117931694300922202652900684995719216451701324253462721593"
	if result != expected {
		t.Errorf("got %s, want %s", result, expected)
	}
}

func TestHashASCIIStrToFieldLonger(t *testing.T) {
	result, err := zklogin.HashASCIIStrToField("test-user-123")
	if err != nil {
		t.Fatal(err)
	}
	expected := "16367606601575717749645815002188278363021176339449607806144004569537090866233"
	if result != expected {
		t.Errorf("got %s, want %s", result, expected)
	}
}

func TestHashASCIIStrToFieldMultiChunk(t *testing.T) {
	result, err := zklogin.HashASCIIStrToField("test-client-id.apps.example.com")
	if err != nil {
		t.Fatal(err)
	}
	expected := "3404679439414438703701472354939133021268187379432033818480239280391781259276"
	if result != expected {
		t.Errorf("got %s, want %s", result, expected)
	}
}

func TestGenAddressSeed(t *testing.T) {
	seed, err := zklogin.GenAddressSeed(
		"12345678901234567890",
		"sub",
		"test-user-123",
		"test-client-id.apps.example.com",
	)
	if err != nil {
		t.Fatal(err)
	}
	expected := "18203684711091342283304588450803600367439511488086209929018923315000786681108"
	if seed != expected {
		t.Errorf("got %s, want %s", seed, expected)
	}
}

func TestNormalizeIssuer(t *testing.T) {
	if got := zklogin.NormalizeIssuer("accounts.google.com"); got != "https://accounts.google.com" {
		t.Errorf("got %s", got)
	}
	if got := zklogin.NormalizeIssuer("https://accounts.google.com"); got != "https://accounts.google.com" {
		t.Errorf("got %s", got)
	}
}

func TestDeriveAddress(t *testing.T) {
	seed, err := zklogin.GenAddressSeed(
		"12345678901234567890",
		"sub",
		"test-user-123",
		"test-client-id.apps.example.com",
	)
	if err != nil {
		t.Fatal(err)
	}

	addr, err := zklogin.DeriveAddress("https://accounts.google.com", seed)
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

func TestComputeAddress(t *testing.T) {
	addr, err := zklogin.ComputeAddress(
		"https://accounts.google.com",
		"test-client-id.apps.example.com",
		"test-user-123",
		"12345678901234567890",
	)
	if err != nil {
		t.Fatal(err)
	}

	// Should match DeriveAddress with same seed
	seed, _ := zklogin.GenAddressSeed("12345678901234567890", "sub", "test-user-123", "test-client-id.apps.example.com")
	expected, _ := zklogin.DeriveAddress("https://accounts.google.com", seed)
	if addr != expected {
		t.Errorf("mismatch: got %s, want %s", addr, expected)
	}
}

func TestComputeAddressDeterministic(t *testing.T) {
	addr1, _ := zklogin.ComputeAddress("https://accounts.google.com", "client-id", "user-sub", "999")
	addr2, _ := zklogin.ComputeAddress("https://accounts.google.com", "client-id", "user-sub", "999")
	if addr1 != addr2 {
		t.Error("same inputs should produce same address")
	}
}

func TestDifferentSaltDifferentAddress(t *testing.T) {
	addr1, _ := zklogin.ComputeAddress("https://accounts.google.com", "client", "sub", "111")
	addr2, _ := zklogin.ComputeAddress("https://accounts.google.com", "client", "sub", "222")
	if addr1 == addr2 {
		t.Error("different salts should produce different addresses")
	}
}

func TestPoseidonHashInvalidDecimal(t *testing.T) {
	_, err := zklogin.PoseidonHash([]string{"not-a-number"})
	if err == nil {
		t.Error("expected error for invalid decimal string")
	}
	if !strings.Contains(err.Error(), "invalid decimal input") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestPoseidonHashMixedValidInvalid(t *testing.T) {
	_, err := zklogin.PoseidonHash([]string{"1", "abc", "3"})
	if err == nil {
		t.Error("expected error for invalid decimal in mixed inputs")
	}
}

func TestGenAddressSeedInvalidSalt(t *testing.T) {
	_, err := zklogin.GenAddressSeed("not-a-number", "sub", "user", "aud")
	if err == nil {
		t.Error("expected error for invalid salt")
	}
}

func TestAddressSeedToBytesInvalid(t *testing.T) {
	_, err := zklogin.AddressSeedToBytes("not-a-number")
	if err == nil {
		t.Error("expected error for invalid seed string")
	}
	if !strings.Contains(err.Error(), "invalid address seed") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAddressSeedToBytesTooLarge(t *testing.T) {
	// 33 bytes worth of value (2^264) is too large for 32 bytes
	tooLarge := "29642774844752946028434172162224104410437116074403984394101141506025761187823616"
	_, err := zklogin.AddressSeedToBytes(tooLarge)
	if err == nil {
		t.Error("expected error for too-large seed")
	}
	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDeriveAddressInvalidSeed(t *testing.T) {
	_, err := zklogin.DeriveAddress("https://accounts.google.com", "not-a-number")
	if err == nil {
		t.Error("expected error for invalid address seed")
	}
}

func TestComputeAddressInvalidSalt(t *testing.T) {
	_, err := zklogin.ComputeAddress("https://accounts.google.com", "aud", "sub", "bad-salt")
	if err == nil {
		t.Error("expected error for invalid salt in ComputeAddress")
	}
}

func TestHashASCIIStrToFieldEmpty(t *testing.T) {
	result, err := zklogin.HashASCIIStrToField("")
	if err != nil {
		t.Fatal(err)
	}
	// Empty string should hash "0" via Poseidon
	if result == "" {
		t.Error("expected non-empty result for empty string")
	}
}

func TestAssembleWithInvalidG1Proof(t *testing.T) {
	// A point with only 1 element instead of 2
	proof := zklogin.ZkLoginProof{
		A: []string{"1"},            // too few elements
		B: [][]string{{"3", "4"}, {"5", "6"}},
		C: []string{"7", "8"},
	}
	inputs := zklogin.ZkLoginInputs{
		Proof:            proof,
		IssBase64Details: base64.StdEncoding.EncodeToString([]byte("test")),
		HeaderBase64:     "header",
		AddressSeed:      "123",
	}
	dummySig := base64.StdEncoding.EncodeToString(make([]byte, 97))
	_, err := zklogin.Assemble(inputs, 100, dummySig)
	if err == nil {
		t.Error("expected error for invalid G1 point A")
	}
}

func TestAssembleWithInvalidG2Proof(t *testing.T) {
	proof := zklogin.ZkLoginProof{
		A: []string{"1", "2"},
		B: [][]string{{"3"}}, // too few elements in G2
		C: []string{"7", "8"},
	}
	inputs := zklogin.ZkLoginInputs{
		Proof:            proof,
		IssBase64Details: base64.StdEncoding.EncodeToString([]byte("test")),
		HeaderBase64:     "header",
		AddressSeed:      "123",
	}
	dummySig := base64.StdEncoding.EncodeToString(make([]byte, 97))
	_, err := zklogin.Assemble(inputs, 100, dummySig)
	if err == nil {
		t.Error("expected error for invalid G2 point B")
	}
}

func TestAssembleWithInvalidG1C(t *testing.T) {
	proof := zklogin.ZkLoginProof{
		A: []string{"1", "2"},
		B: [][]string{{"3", "4"}, {"5", "6"}},
		C: []string{"7"}, // too few elements
	}
	inputs := zklogin.ZkLoginInputs{
		Proof:            proof,
		IssBase64Details: base64.StdEncoding.EncodeToString([]byte("test")),
		HeaderBase64:     "header",
		AddressSeed:      "123",
	}
	dummySig := base64.StdEncoding.EncodeToString(make([]byte, 97))
	_, err := zklogin.Assemble(inputs, 100, dummySig)
	if err == nil {
		t.Error("expected error for invalid G1 point C")
	}
}

func TestAssembleWithInvalidFieldElement(t *testing.T) {
	proof := zklogin.ZkLoginProof{
		A: []string{"not-a-number", "2"},
		B: [][]string{{"3", "4"}, {"5", "6"}},
		C: []string{"7", "8"},
	}
	inputs := zklogin.ZkLoginInputs{
		Proof:            proof,
		IssBase64Details: base64.StdEncoding.EncodeToString([]byte("test")),
		HeaderBase64:     "header",
		AddressSeed:      "123",
	}
	dummySig := base64.StdEncoding.EncodeToString(make([]byte, 97))
	_, err := zklogin.Assemble(inputs, 100, dummySig)
	if err == nil {
		t.Error("expected error for invalid field element")
	}
}

func TestAssembleWithTooLargeFieldElement(t *testing.T) {
	// Value > 32 bytes
	tooLarge := "29642774844752946028434172162224104410437116074403984394101141506025761187823616"
	proof := zklogin.ZkLoginProof{
		A: []string{tooLarge, "2"},
		B: [][]string{{"3", "4"}, {"5", "6"}},
		C: []string{"7", "8"},
	}
	inputs := zklogin.ZkLoginInputs{
		Proof:            proof,
		IssBase64Details: base64.StdEncoding.EncodeToString([]byte("test")),
		HeaderBase64:     "header",
		AddressSeed:      "123",
	}
	dummySig := base64.StdEncoding.EncodeToString(make([]byte, 97))
	_, err := zklogin.Assemble(inputs, 100, dummySig)
	if err == nil {
		t.Error("expected error for too-large field element")
	}
}

func TestAssembleWithInvalidUserSignature(t *testing.T) {
	proof := zklogin.ZkLoginProof{
		A: []string{"1", "2"},
		B: [][]string{{"3", "4"}, {"5", "6"}},
		C: []string{"7", "8"},
	}
	inputs := zklogin.ZkLoginInputs{
		Proof:            proof,
		IssBase64Details: base64.StdEncoding.EncodeToString([]byte("test")),
		HeaderBase64:     "header",
		AddressSeed:      "123",
	}
	_, err := zklogin.Assemble(inputs, 100, "not-valid-base64!!!")
	if err == nil {
		t.Error("expected error for invalid base64 user signature")
	}
}

func TestAssembleWithRawIssDetails(t *testing.T) {
	// IssBase64Details that is NOT valid base64 — should be treated as raw bytes
	proof := zklogin.ZkLoginProof{
		A: []string{"1", "2"},
		B: [][]string{{"3", "4"}, {"5", "6"}},
		C: []string{"7", "8"},
	}
	inputs := zklogin.ZkLoginInputs{
		Proof:            proof,
		IssBase64Details: "not-valid-base64!!!",
		HeaderBase64:     "header",
		AddressSeed:      "123",
	}
	dummySig := base64.StdEncoding.EncodeToString(make([]byte, 97))
	result, err := zklogin.Assemble(inputs, 100, dummySig)
	if err != nil {
		t.Fatalf("expected raw iss details to be accepted, got: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestAssembleWithInvalidG2InnerElement(t *testing.T) {
	// G2 with correct outer length but invalid inner decimal
	proof := zklogin.ZkLoginProof{
		A: []string{"1", "2"},
		B: [][]string{{"3", "bad"}, {"5", "6"}},
		C: []string{"7", "8"},
	}
	inputs := zklogin.ZkLoginInputs{
		Proof:            proof,
		IssBase64Details: base64.StdEncoding.EncodeToString([]byte("test")),
		HeaderBase64:     "header",
		AddressSeed:      "123",
	}
	dummySig := base64.StdEncoding.EncodeToString(make([]byte, 97))
	_, err := zklogin.Assemble(inputs, 100, dummySig)
	if err == nil {
		t.Error("expected error for invalid G2 inner element")
	}
}

func TestNormalizeIssuerHTTP(t *testing.T) {
	got := zklogin.NormalizeIssuer("http://example.com")
	if got != "http://example.com" {
		t.Errorf("http prefix should be preserved, got %s", got)
	}
}

func TestAssembleAuthenticator(t *testing.T) {
	proof := zklogin.ZkLoginProof{
		A: []string{"1", "2"},
		B: [][]string{{"3", "4"}, {"5", "6"}},
		C: []string{"7", "8"},
	}

	inputs := zklogin.ZkLoginInputs{
		Proof:            proof,
		IssBase64Details: base64.StdEncoding.EncodeToString([]byte("test-iss-details")),
		HeaderBase64:     "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9",
		AddressSeed:      "12345",
	}

	dummySig := base64.StdEncoding.EncodeToString(append([]byte{0x00}, make([]byte, 96)...))

	result, err := zklogin.Assemble(inputs, 100, dummySig)
	if err != nil {
		t.Fatal(err)
	}

	decoded, err := base64.StdEncoding.DecodeString(result)
	if err != nil {
		t.Fatal(err)
	}
	if decoded[0] != 0x05 {
		t.Errorf("expected zkLogin flag 0x05, got 0x%02x", decoded[0])
	}
}
