package txn

import (
	"testing"

	"github.com/inodrahq/go-bcs"
)

// Cover no-op UnmarshalBCS methods that are structurally never called
// from the public decode path (GasCoin has no data, simpleTypeTag has no data).

func TestGasCoinVariant_UnmarshalBCS(t *testing.T) {
	v := &gasCoinVariant{}
	d := bcs.NewDecoder([]byte{})
	if err := v.UnmarshalBCS(d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSimpleTypeTag_UnmarshalBCS(t *testing.T) {
	v := &simpleTypeTag{tag: 0}
	d := bcs.NewDecoder([]byte{})
	if err := v.UnmarshalBCS(d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
