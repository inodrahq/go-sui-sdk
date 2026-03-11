package txn

import (
	"fmt"

	"github.com/inodrahq/go-bcs"
)

// Command represents a programmable transaction command (BCS enum).
// MoveCall=0, TransferObjects=1, SplitCoins=2, MergeCoins=3, Publish=4, MakeMoveVec=5, Upgrade=6
type Command struct {
	variant commandVariant
}

type commandVariant interface {
	bcs.Marshaler
	bcs.Unmarshaler
	commandTag() uint32
}

// MoveCallCommand creates a MoveCall command.
func MoveCallCommand(call MoveCall) Command {
	return Command{variant: &call}
}

// TransferObjectsCommand creates a TransferObjects command.
func TransferObjectsCommand(objects []Argument, destination Argument) Command {
	return Command{variant: &TransferObjects{Objects: objects, Destination: destination}}
}

// SplitCoinsCommand creates a SplitCoins command.
func SplitCoinsCommand(coin Argument, amounts []Argument) Command {
	return Command{variant: &SplitCoins{Coin: coin, Amounts: amounts}}
}

// MergeCoinsCommand creates a MergeCoins command.
func MergeCoinsCommand(destination Argument, sources []Argument) Command {
	return Command{variant: &MergeCoins{Destination: destination, Sources: sources}}
}

// PublishCommand creates a Publish command.
func PublishCommand(modules [][]byte, dependencies []SuiAddress) Command {
	return Command{variant: &Publish{Modules: modules, Dependencies: dependencies}}
}

func (c Command) MarshalBCS(e *bcs.Encoder) error {
	if err := e.WriteULEB128(c.variant.commandTag()); err != nil {
		return err
	}
	return c.variant.MarshalBCS(e)
}

func (c *Command) UnmarshalBCS(d *bcs.Decoder) error {
	tag, err := d.ReadULEB128()
	if err != nil {
		return err
	}
	switch tag {
	case 0:
		v := &MoveCall{}
		if err := v.UnmarshalBCS(d); err != nil {
			return err
		}
		c.variant = v
	case 1:
		v := &TransferObjects{}
		if err := v.UnmarshalBCS(d); err != nil {
			return err
		}
		c.variant = v
	case 2:
		v := &SplitCoins{}
		if err := v.UnmarshalBCS(d); err != nil {
			return err
		}
		c.variant = v
	case 3:
		v := &MergeCoins{}
		if err := v.UnmarshalBCS(d); err != nil {
			return err
		}
		c.variant = v
	case 4:
		v := &Publish{}
		if err := v.UnmarshalBCS(d); err != nil {
			return err
		}
		c.variant = v
	default:
		return fmt.Errorf("unknown Command tag: %d", tag)
	}
	return nil
}

// MoveCall represents a Move function call.
type MoveCall struct {
	Package  SuiAddress
	Module   bcs.String
	Function bcs.String
	TypeArgs []TypeTag
	Args     []Argument
}

func (*MoveCall) commandTag() uint32 { return 0 }

func (m *MoveCall) MarshalBCS(e *bcs.Encoder) error {
	if err := m.Package.MarshalBCS(e); err != nil {
		return err
	}
	if err := m.Module.MarshalBCS(e); err != nil {
		return err
	}
	if err := m.Function.MarshalBCS(e); err != nil {
		return err
	}
	// TypeArgs as Vec<TypeTag>
	if err := e.WriteULEB128(uint32(len(m.TypeArgs))); err != nil {
		return err
	}
	for i := range m.TypeArgs {
		if err := m.TypeArgs[i].MarshalBCS(e); err != nil {
			return err
		}
	}
	// Args as Vec<Argument>
	if err := e.WriteULEB128(uint32(len(m.Args))); err != nil {
		return err
	}
	for i := range m.Args {
		if err := m.Args[i].MarshalBCS(e); err != nil {
			return err
		}
	}
	return nil
}

func (m *MoveCall) UnmarshalBCS(d *bcs.Decoder) error {
	if err := m.Package.UnmarshalBCS(d); err != nil {
		return err
	}
	if err := m.Module.UnmarshalBCS(d); err != nil {
		return err
	}
	if err := m.Function.UnmarshalBCS(d); err != nil {
		return err
	}
	// TypeArgs
	n, err := d.ReadULEB128()
	if err != nil {
		return err
	}
	m.TypeArgs = make([]TypeTag, n)
	for i := uint32(0); i < n; i++ {
		if err := m.TypeArgs[i].UnmarshalBCS(d); err != nil {
			return err
		}
	}
	// Args
	n, err = d.ReadULEB128()
	if err != nil {
		return err
	}
	m.Args = make([]Argument, n)
	for i := uint32(0); i < n; i++ {
		if err := m.Args[i].UnmarshalBCS(d); err != nil {
			return err
		}
	}
	return nil
}

// TransferObjects transfers objects to a destination.
type TransferObjects struct {
	Objects     []Argument
	Destination Argument
}

func (*TransferObjects) commandTag() uint32 { return 1 }

func (t *TransferObjects) MarshalBCS(e *bcs.Encoder) error {
	if err := e.WriteULEB128(uint32(len(t.Objects))); err != nil {
		return err
	}
	for i := range t.Objects {
		if err := t.Objects[i].MarshalBCS(e); err != nil {
			return err
		}
	}
	return t.Destination.MarshalBCS(e)
}

func (t *TransferObjects) UnmarshalBCS(d *bcs.Decoder) error {
	n, err := d.ReadULEB128()
	if err != nil {
		return err
	}
	t.Objects = make([]Argument, n)
	for i := uint32(0); i < n; i++ {
		if err := t.Objects[i].UnmarshalBCS(d); err != nil {
			return err
		}
	}
	return t.Destination.UnmarshalBCS(d)
}

// SplitCoins splits a coin into multiple amounts.
type SplitCoins struct {
	Coin    Argument
	Amounts []Argument
}

func (*SplitCoins) commandTag() uint32 { return 2 }

func (s *SplitCoins) MarshalBCS(e *bcs.Encoder) error {
	if err := s.Coin.MarshalBCS(e); err != nil {
		return err
	}
	if err := e.WriteULEB128(uint32(len(s.Amounts))); err != nil {
		return err
	}
	for i := range s.Amounts {
		if err := s.Amounts[i].MarshalBCS(e); err != nil {
			return err
		}
	}
	return nil
}

func (s *SplitCoins) UnmarshalBCS(d *bcs.Decoder) error {
	if err := s.Coin.UnmarshalBCS(d); err != nil {
		return err
	}
	n, err := d.ReadULEB128()
	if err != nil {
		return err
	}
	s.Amounts = make([]Argument, n)
	for i := uint32(0); i < n; i++ {
		if err := s.Amounts[i].UnmarshalBCS(d); err != nil {
			return err
		}
	}
	return nil
}

// MergeCoins merges source coins into a destination coin.
type MergeCoins struct {
	Destination Argument
	Sources     []Argument
}

func (*MergeCoins) commandTag() uint32 { return 3 }

func (m *MergeCoins) MarshalBCS(e *bcs.Encoder) error {
	if err := m.Destination.MarshalBCS(e); err != nil {
		return err
	}
	if err := e.WriteULEB128(uint32(len(m.Sources))); err != nil {
		return err
	}
	for i := range m.Sources {
		if err := m.Sources[i].MarshalBCS(e); err != nil {
			return err
		}
	}
	return nil
}

func (m *MergeCoins) UnmarshalBCS(d *bcs.Decoder) error {
	if err := m.Destination.UnmarshalBCS(d); err != nil {
		return err
	}
	n, err := d.ReadULEB128()
	if err != nil {
		return err
	}
	m.Sources = make([]Argument, n)
	for i := uint32(0); i < n; i++ {
		if err := m.Sources[i].UnmarshalBCS(d); err != nil {
			return err
		}
	}
	return nil
}

// Publish publishes a new Move package.
type Publish struct {
	Modules      [][]byte     // compiled module bytecode
	Dependencies []SuiAddress // transitive dependency package IDs
}

func (*Publish) commandTag() uint32 { return 4 }

func (p *Publish) MarshalBCS(e *bcs.Encoder) error {
	// Vec<Vec<u8>> — modules
	if err := e.WriteULEB128(uint32(len(p.Modules))); err != nil {
		return err
	}
	for _, mod := range p.Modules {
		// Each module is a ByteVector (ULEB128 length + bytes)
		if err := e.WriteULEB128(uint32(len(mod))); err != nil {
			return err
		}
		if err := e.WriteBytes(mod); err != nil {
			return err
		}
	}
	// Vec<SuiAddress> — dependencies
	if err := e.WriteULEB128(uint32(len(p.Dependencies))); err != nil {
		return err
	}
	for i := range p.Dependencies {
		if err := p.Dependencies[i].MarshalBCS(e); err != nil {
			return err
		}
	}
	return nil
}

func (p *Publish) UnmarshalBCS(d *bcs.Decoder) error {
	nMods, err := d.ReadULEB128()
	if err != nil {
		return err
	}
	p.Modules = make([][]byte, nMods)
	for i := uint32(0); i < nMods; i++ {
		length, err := d.ReadULEB128()
		if err != nil {
			return err
		}
		mod, err := d.ReadBytes(int(length))
		if err != nil {
			return err
		}
		p.Modules[i] = mod
	}
	nDeps, err := d.ReadULEB128()
	if err != nil {
		return err
	}
	p.Dependencies = make([]SuiAddress, nDeps)
	for i := uint32(0); i < nDeps; i++ {
		if err := p.Dependencies[i].UnmarshalBCS(d); err != nil {
			return err
		}
	}
	return nil
}
