package txn

import "github.com/inodrahq/go-bcs"

// ProgrammableTransaction contains inputs and commands.
// BCS field order: inputs, commands
type ProgrammableTransaction struct {
	Inputs   []CallArg
	Commands []Command
}

func (pt ProgrammableTransaction) MarshalBCS(e *bcs.Encoder) error {
	// Inputs as Vec<CallArg>
	if err := e.WriteULEB128(uint32(len(pt.Inputs))); err != nil {
		return err
	}
	for i := range pt.Inputs {
		if err := pt.Inputs[i].MarshalBCS(e); err != nil {
			return err
		}
	}
	// Commands as Vec<Command>
	if err := e.WriteULEB128(uint32(len(pt.Commands))); err != nil {
		return err
	}
	for i := range pt.Commands {
		if err := pt.Commands[i].MarshalBCS(e); err != nil {
			return err
		}
	}
	return nil
}

func (pt *ProgrammableTransaction) UnmarshalBCS(d *bcs.Decoder) error {
	// Inputs
	n, err := d.ReadULEB128()
	if err != nil {
		return err
	}
	pt.Inputs = make([]CallArg, n)
	for i := uint32(0); i < n; i++ {
		if err := pt.Inputs[i].UnmarshalBCS(d); err != nil {
			return err
		}
	}
	// Commands
	n, err = d.ReadULEB128()
	if err != nil {
		return err
	}
	pt.Commands = make([]Command, n)
	for i := uint32(0); i < n; i++ {
		if err := pt.Commands[i].UnmarshalBCS(d); err != nil {
			return err
		}
	}
	return nil
}
