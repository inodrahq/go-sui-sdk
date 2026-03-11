package txn

import "github.com/inodrahq/go-bcs"

// GasData contains gas payment info for a transaction.
// BCS field order: payment, owner, price, budget
type GasData struct {
	Payment []ObjectRef
	Owner   SuiAddress
	Price   bcs.U64
	Budget  bcs.U64
}

func (g GasData) MarshalBCS(e *bcs.Encoder) error {
	// Payment as Vec<ObjectRef>
	if err := e.WriteULEB128(uint32(len(g.Payment))); err != nil {
		return err
	}
	for i := range g.Payment {
		if err := g.Payment[i].MarshalBCS(e); err != nil {
			return err
		}
	}
	if err := g.Owner.MarshalBCS(e); err != nil {
		return err
	}
	if err := g.Price.MarshalBCS(e); err != nil {
		return err
	}
	return g.Budget.MarshalBCS(e)
}

func (g *GasData) UnmarshalBCS(d *bcs.Decoder) error {
	n, err := d.ReadULEB128()
	if err != nil {
		return err
	}
	g.Payment = make([]ObjectRef, n)
	for i := uint32(0); i < n; i++ {
		if err := g.Payment[i].UnmarshalBCS(d); err != nil {
			return err
		}
	}
	if err := g.Owner.UnmarshalBCS(d); err != nil {
		return err
	}
	if err := g.Price.UnmarshalBCS(d); err != nil {
		return err
	}
	return g.Budget.UnmarshalBCS(d)
}
