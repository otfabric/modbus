package modbus

import (
	"errors"
	"testing"
)

// minimalUint16Codec is a test codec: one register, no permutation.
type minimalUint16Codec struct{}

func (minimalUint16Codec) ID() string   { return "uint16/test" }
func (minimalUint16Codec) Name() string { return "uint16" }
func (minimalUint16Codec) RegisterSpec() RegisterSpec {
	return RegisterSpec{Count: 1}
}
func (minimalUint16Codec) ByteSpec() ByteSpec {
	return ByteSpec{Count: 2}
}
func (c minimalUint16Codec) DecodeRegisters(regs []uint16) (uint16, error) {
	if err := ValidateRegisterSpec(c.RegisterSpec(), regs, c.ID()); err != nil {
		return 0, err
	}
	return regs[0], nil
}
func (c minimalUint16Codec) EncodeRegisters(v uint16) ([]uint16, error) {
	return []uint16{v}, nil
}

func TestDecodeRegisters_OK(t *testing.T) {
	codec := minimalUint16Codec{}
	v, err := DecodeRegisters([]uint16{0x1234}, codec)
	if err != nil {
		t.Fatal(err)
	}
	if v != 0x1234 {
		t.Errorf("got 0x%x, want 0x1234", v)
	}
}

func TestDecodeRegisters_WrongCount(t *testing.T) {
	codec := minimalUint16Codec{}
	_, err := DecodeRegisters([]uint16{0x12, 0x34}, codec)
	if err == nil {
		t.Fatal("expected error for wrong register count")
	}
	var e *CodecRegisterCountError
	if !errors.As(err, &e) {
		t.Errorf("expected CodecRegisterCountError, got %T", err)
	}
}

func TestEncodeRegisters_OK(t *testing.T) {
	codec := minimalUint16Codec{}
	regs, err := EncodeRegisters(uint16(0x5678), codec)
	if err != nil {
		t.Fatal(err)
	}
	if len(regs) != 1 || regs[0] != 0x5678 {
		t.Errorf("got %v, want [0x5678]", regs)
	}
}

func TestEncodeRegisters_Validation(t *testing.T) {
	// EncodeRegisters validates the codec's output; a buggy codec that returns wrong count would be caught
	// Here we just ensure EncodeRegisters returns what the codec returns when valid
	codec := minimalUint16Codec{}
	regs, err := EncodeRegisters(uint16(0), codec)
	if err != nil {
		t.Fatal(err)
	}
	if len(regs) != 1 {
		t.Errorf("len(regs) = %d, want 1", len(regs))
	}
}
