package client

import (
	"testing"

	pb "github.com/inodrahq/go-sui-sdk/pb/sui/rpc/v2"
)

func TestObjectToRef(t *testing.T) {
	// A valid object with base58-encoded digest (32 bytes of zeros = "11111111111111111111111111111111" in base58).
	obj := &pb.Object{}
	obj.ObjectId = strPtr("0x0000000000000000000000000000000000000000000000000000000000000005")
	v := uint64(42)
	obj.Version = &v
	obj.Digest = strPtr("11111111111111111111111111111111")

	ref, err := objectToRef(obj)
	if err != nil {
		t.Fatal(err)
	}
	if ref.ObjectID[31] != 5 {
		t.Errorf("expected object ID to end with 5, got %x", ref.ObjectID)
	}
	if uint64(ref.Version) != 42 {
		t.Errorf("expected version 42, got %d", ref.Version)
	}
}

func TestObjectToRefInvalidObjectId(t *testing.T) {
	obj := &pb.Object{}
	obj.ObjectId = strPtr("not-a-valid-address")
	obj.Digest = strPtr("11111111111111111111111111111111")

	_, err := objectToRef(obj)
	if err == nil {
		t.Fatal("expected error for invalid object ID")
	}
}

func TestObjectToRefInvalidDigest(t *testing.T) {
	obj := &pb.Object{}
	obj.ObjectId = strPtr("0x0000000000000000000000000000000000000000000000000000000000000005")
	v := uint64(1)
	obj.Version = &v
	obj.Digest = strPtr("!!!invalid-base58!!!")

	_, err := objectToRef(obj)
	if err == nil {
		t.Fatal("expected error for invalid digest")
	}
}

func TestObjectToRefShortDigest(t *testing.T) {
	obj := &pb.Object{}
	obj.ObjectId = strPtr("0x0000000000000000000000000000000000000000000000000000000000000005")
	v := uint64(1)
	obj.Version = &v
	// "1" in base58 is a single zero byte — too short for a 32-byte digest.
	obj.Digest = strPtr("1")

	_, err := objectToRef(obj)
	if err == nil {
		t.Fatal("expected error for too-short digest")
	}
}

func TestStrPtr(t *testing.T) {
	p := strPtr("hello")
	if *p != "hello" {
		t.Errorf("expected 'hello', got '%s'", *p)
	}
}

func TestUint32Ptr(t *testing.T) {
	p := uint32Ptr(42)
	if *p != 42 {
		t.Errorf("expected 42, got %d", *p)
	}
}
