package modbus

import (
	"errors"
	"testing"
)

func TestNewAsciiCodec_RejectZero(t *testing.T) {
	_, err := NewAsciiCodec(0)
	if err == nil {
		t.Fatal("expected error for registerCount 0")
	}
	if !errors.Is(err, ErrCodecValue) {
		t.Errorf("expected ErrCodecValue, got %v", err)
	}
}

func TestNewAsciiCodec_RoundTrip(t *testing.T) {
	c, err := NewAsciiCodec(4)
	if err != nil {
		t.Fatal(err)
	}
	s := "AB"
	regs, err := EncodeRegisters(s, c)
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecodeRegisters(regs, c)
	if err != nil {
		t.Fatal(err)
	}
	if got != "AB" {
		t.Errorf("round-trip: got %q, want AB", got)
	}
}

func TestAsciiCodec_TrimTrailingSpaces(t *testing.T) {
	c := mustTextCodec(t, "ascii", 2)
	regs := []uint16{0x4142, 0x4320}
	got, err := DecodeRegisters(regs, c)
	if err != nil {
		t.Fatal(err)
	}
	if got != "ABC" {
		t.Errorf("DecodeRegisters (trim spaces) = %q, want ABC", got)
	}
}

func TestAsciiFixedCodec_PreserveSpaces(t *testing.T) {
	c, err := NewAsciiFixedCodec(2)
	if err != nil {
		t.Fatal(err)
	}
	regs := []uint16{0x4142, 0x4320}
	got, err := DecodeRegisters(regs, c)
	if err != nil {
		t.Fatal(err)
	}
	if got != "ABC " {
		t.Errorf("DecodeRegisters (preserve) = %q, want ABC ", got)
	}
}

func TestAsciiReverseCodec_RoundTrip(t *testing.T) {
	c, err := NewAsciiReverseCodec(2)
	if err != nil {
		t.Fatal(err)
	}
	s := "AB"
	regs, err := EncodeRegisters(s, c)
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecodeRegisters(regs, c)
	if err != nil {
		t.Fatal(err)
	}
	if got != "AB" {
		t.Errorf("round-trip: got %q", got)
	}
}

func TestBCDCodec_RoundTrip(t *testing.T) {
	c, err := NewBCDCodec(2)
	if err != nil {
		t.Fatal(err)
	}
	s := "1234"
	regs, err := EncodeRegisters(s, c)
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecodeRegisters(regs, c)
	if err != nil {
		t.Fatal(err)
	}
	if got != s {
		t.Errorf("round-trip: got %q, want %q", got, s)
	}
}

func TestBCDCodec_RejectNonDigit(t *testing.T) {
	c, err := NewBCDCodec(2)
	if err != nil {
		t.Fatal(err)
	}
	_, err = EncodeRegisters("12a4", c)
	if err == nil {
		t.Fatal("expected error for non-digit")
	}
}

func TestPackedBCDCodec_RoundTrip(t *testing.T) {
	c, err := NewPackedBCDCodec(2)
	if err != nil {
		t.Fatal(err)
	}
	// 2 registers = 4 bytes = 8 digits; full width round-trip
	s := "12345678"
	regs, err := EncodeRegisters(s, c)
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecodeRegisters(regs, c)
	if err != nil {
		t.Fatal(err)
	}
	if got != s {
		t.Errorf("round-trip: got %q, want %q", got, s)
	}
}

func TestPackedBCDCodec_RejectNonDigit(t *testing.T) {
	c, err := NewPackedBCDCodec(2)
	if err != nil {
		t.Fatal(err)
	}
	_, err = EncodeRegisters("12x4", c)
	if err == nil {
		t.Fatal("expected error for non-digit")
	}
}

func TestAsciiCodec_RejectNonASCII(t *testing.T) {
	c, err := NewAsciiCodec(2)
	if err != nil {
		t.Fatal(err)
	}
	_, err = EncodeRegisters("café", c)
	if err == nil {
		t.Fatal("expected error for non-ASCII (UTF-8 multi-byte)")
	}
}

func TestAsciiCodec_RejectNonASCII_BeyondWidth(t *testing.T) {
	// Full input is validated; non-ASCII beyond the codec width must still be rejected.
	c, err := NewAsciiCodec(2)
	if err != nil {
		t.Fatal(err)
	}
	_, err = EncodeRegisters("ABé", c)
	if err == nil {
		t.Fatal("expected error when non-ASCII appears after valid prefix")
	}
}

func TestAsciiCodec_OverlongASCII_Truncated(t *testing.T) {
	// Overlong but all-ASCII input is truncated to width, not rejected.
	c, err := NewAsciiCodec(1)
	if err != nil {
		t.Fatal(err)
	}
	regs, err := EncodeRegisters("ABCD", c)
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecodeRegisters(regs, c)
	if err != nil {
		t.Fatal(err)
	}
	if got != "AB" {
		t.Errorf("overlong ASCII truncated to width: got %q, want AB", got)
	}
}

func TestTextCodec_ZeroRegistersRejected(t *testing.T) {
	for name, fn := range map[string]func(uint16) (Codec[string], error){
		"ascii":         NewAsciiCodec,
		"ascii_fixed":   NewAsciiFixedCodec,
		"ascii_reverse": NewAsciiReverseCodec,
		"bcd":           NewBCDCodec,
		"packed_bcd":    NewPackedBCDCodec,
	} {
		_, err := fn(0)
		if err == nil {
			t.Errorf("%s(0): expected error", name)
		}
	}
}

func mustTextCodec(t *testing.T, kind string, n uint16) Codec[string] {
	t.Helper()
	var c Codec[string]
	var err error
	switch kind {
	case "ascii":
		c, err = NewAsciiCodec(n)
	case "ascii_fixed":
		c, err = NewAsciiFixedCodec(n)
	case "ascii_reverse":
		c, err = NewAsciiReverseCodec(n)
	case "bcd":
		c, err = NewBCDCodec(n)
	case "packed_bcd":
		c, err = NewPackedBCDCodec(n)
	default:
		t.Fatalf("unknown kind %s", kind)
	}
	if err != nil {
		t.Fatal(err)
	}
	return c
}
