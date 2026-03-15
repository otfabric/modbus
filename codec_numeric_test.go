package modbus

import (
	"math"
	"testing"
)

func TestNewUint32Codec_ValidLayout(t *testing.T) {
	c, err := NewUint32Codec(Layout32_4321)
	if err != nil {
		t.Fatal(err)
	}
	if c.RegisterSpec().Count != 2 {
		t.Errorf("RegisterSpec().Count = %d, want 2", c.RegisterSpec().Count)
	}
	v, err := DecodeRegisters([]uint16{0x1234, 0x5678}, c)
	if err != nil {
		t.Fatal(err)
	}
	if v != 0x12345678 {
		t.Errorf("DecodeRegisters = 0x%x, want 0x12345678", v)
	}
	regs, err := EncodeRegisters(uint32(0x12345678), c)
	if err != nil {
		t.Fatal(err)
	}
	if len(regs) != 2 || regs[0] != 0x1234 || regs[1] != 0x5678 {
		t.Errorf("EncodeRegisters = %v", regs)
	}
}

func TestNewUint32Codec_InvalidLayout(t *testing.T) {
	_, err := NewUint32Codec(Layout64_87654321)
	if err == nil {
		t.Fatal("expected error for wrong layout register count")
	}
	if !isErrCodecLayout(err) {
		t.Errorf("expected layout error, got %v", err)
	}
}

func isErrCodecLayout(err error) bool {
	for err != nil {
		if err == ErrCodecLayout {
			return true
		}
		type unwrap interface{ Unwrap() error }
		if u, ok := err.(unwrap); ok {
			err = u.Unwrap()
		} else {
			break
		}
	}
	return false
}

func TestUint32Codec_RoundTrip(t *testing.T) {
	c := MustNewUint32Codec(Layout32_2143)
	val := uint32(0xDEADBEEF)
	regs, err := EncodeRegisters(val, c)
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecodeRegisters(regs, c)
	if err != nil {
		t.Fatal(err)
	}
	if got != val {
		t.Errorf("round-trip: got 0x%x, want 0x%x", got, val)
	}
}

func TestFloat64Codec_RoundTrip(t *testing.T) {
	c := MustNewFloat64Codec(Layout64_87654321)
	val := 3.14159265358979
	regs, err := EncodeRegisters(val, c)
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecodeRegisters(regs, c)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(got-val) > 1e-10 {
		t.Errorf("round-trip: got %v, want %v", got, val)
	}
}

func TestUint48Codec_RoundTrip(t *testing.T) {
	c := MustNewUint48Codec(Layout48_654321)
	val := uint64(0x0000FFFFFFFFFFFF)
	regs, err := EncodeRegisters(val, c)
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecodeRegisters(regs, c)
	if err != nil {
		t.Fatal(err)
	}
	if got != val {
		t.Errorf("round-trip: got 0x%x, want 0x%x", got, val)
	}
}

func TestInt16Codec_Signed(t *testing.T) {
	c := MustNewInt16Codec(Layout16_21)
	regs, err := EncodeRegisters(int16(-1), c)
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecodeRegisters(regs, c)
	if err != nil {
		t.Fatal(err)
	}
	if got != -1 {
		t.Errorf("DecodeRegisters = %d, want -1", got)
	}
}

func TestDescriptorConsistency_NumericCodecs(t *testing.T) {
	// For each registered descriptor that has a known layout, build the codec and assert RegisterSpec/ByteSpec match.
	all := AvailableCodecDescriptors()
	for _, d := range all {
		if d.Family != CodecFamilyInteger && d.Family != CodecFamilyFloat {
			continue
		}
		if len(d.Layouts) != 1 {
			continue
		}
		layout := d.Layouts[0].Layout
		// Build codec by type and layout
		var spec RegisterSpec
		var byteSpec ByteSpec
		switch d.ValueKind {
		case CodecValueUint16:
			c, err := NewUint16Codec(layout)
			if err != nil {
				t.Errorf("descriptor %s: %v", d.ID, err)
				continue
			}
			spec, byteSpec = c.RegisterSpec(), c.ByteSpec()
		case CodecValueInt16:
			c, err := NewInt16Codec(layout)
			if err != nil {
				t.Errorf("descriptor %s: %v", d.ID, err)
				continue
			}
			spec, byteSpec = c.RegisterSpec(), c.ByteSpec()
		case CodecValueUint32:
			c, err := NewUint32Codec(layout)
			if err != nil {
				t.Errorf("descriptor %s: %v", d.ID, err)
				continue
			}
			spec, byteSpec = c.RegisterSpec(), c.ByteSpec()
		case CodecValueInt32:
			c, _ := NewInt32Codec(layout)
			spec, byteSpec = c.RegisterSpec(), c.ByteSpec()
		case CodecValueFloat32:
			c, _ := NewFloat32Codec(layout)
			spec, byteSpec = c.RegisterSpec(), c.ByteSpec()
		case CodecValueUint48:
			c, _ := NewUint48Codec(layout)
			spec, byteSpec = c.RegisterSpec(), c.ByteSpec()
		case CodecValueInt48:
			c, _ := NewInt48Codec(layout)
			spec, byteSpec = c.RegisterSpec(), c.ByteSpec()
		case CodecValueUint64:
			c, _ := NewUint64Codec(layout)
			spec, byteSpec = c.RegisterSpec(), c.ByteSpec()
		case CodecValueInt64:
			c, _ := NewInt64Codec(layout)
			spec, byteSpec = c.RegisterSpec(), c.ByteSpec()
		case CodecValueFloat64:
			c, _ := NewFloat64Codec(layout)
			spec, byteSpec = c.RegisterSpec(), c.ByteSpec()
		default:
			continue
		}
		if spec != d.RegisterSpec {
			t.Errorf("descriptor %s: RegisterSpec mismatch: codec %+v, descriptor %+v", d.ID, spec, d.RegisterSpec)
		}
		if byteSpec != d.ByteSpec {
			t.Errorf("descriptor %s: ByteSpec mismatch: codec %+v, descriptor %+v", d.ID, byteSpec, d.ByteSpec)
		}
	}
}
