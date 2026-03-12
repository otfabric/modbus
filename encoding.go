package modbus

import (
	"encoding/binary"
	"math"
	"strings"
)

func uint16ToBytes(endianness Endianness, in uint16) (out []byte) {
	out = make([]byte, 2)
	switch endianness {
	case BigEndian:
		binary.BigEndian.PutUint16(out, in)
	case LittleEndian:
		binary.LittleEndian.PutUint16(out, in)
	}

	return
}

func uint16sToBytes(endianness Endianness, in []uint16) (out []byte) {
	for i := range in {
		out = append(out, uint16ToBytes(endianness, in[i])...)
	}

	return
}

func bytesToUint16(endianness Endianness, in []byte) (out uint16) {
	switch endianness {
	case BigEndian:
		out = binary.BigEndian.Uint16(in)
	case LittleEndian:
		out = binary.LittleEndian.Uint16(in)
	}

	return
}

func bytesToUint16s(endianness Endianness, in []byte) (out []uint16) {
	for i := 0; i < len(in); i += 2 {
		out = append(out, bytesToUint16(endianness, in[i:i+2]))
	}

	return
}

func bytesToUint32s(endianness Endianness, wordOrder WordOrder, in []byte) (out []uint32) {
	var u32 uint32

	for i := 0; i < len(in); i += 4 {
		switch endianness {
		case BigEndian:
			if wordOrder == HighWordFirst {
				u32 = binary.BigEndian.Uint32(in[i : i+4])
			} else {
				u32 = binary.BigEndian.Uint32(
					[]byte{in[i+2], in[i+3], in[i+0], in[i+1]})
			}
		case LittleEndian:
			if wordOrder == LowWordFirst {
				u32 = binary.LittleEndian.Uint32(in[i : i+4])
			} else {
				u32 = binary.LittleEndian.Uint32(
					[]byte{in[i+2], in[i+3], in[i+0], in[i+1]})
			}
		}

		out = append(out, u32)
	}

	return
}

func uint32ToBytes(endianness Endianness, wordOrder WordOrder, in uint32) (out []byte) {
	out = make([]byte, 4)

	switch endianness {
	case BigEndian:
		binary.BigEndian.PutUint32(out, in)

		// swap words if needed
		if wordOrder == LowWordFirst {
			out[0], out[1], out[2], out[3] = out[2], out[3], out[0], out[1]
		}
	case LittleEndian:
		binary.LittleEndian.PutUint32(out, in)

		// swap words if needed
		if wordOrder == HighWordFirst {
			out[0], out[1], out[2], out[3] = out[2], out[3], out[0], out[1]
		}
	}

	return
}

func bytesToFloat32s(endianness Endianness, wordOrder WordOrder, in []byte) (out []float32) {
	var u32s = bytesToUint32s(endianness, wordOrder, in)

	for _, u32 := range u32s {
		out = append(out, math.Float32frombits(u32))
	}

	return
}

func float32ToBytes(endianness Endianness, wordOrder WordOrder, in float32) (out []byte) {
	out = uint32ToBytes(endianness, wordOrder, math.Float32bits(in))

	return
}

func bytesToUint64s(endianness Endianness, wordOrder WordOrder, in []byte) (out []uint64) {
	var u64 uint64

	for i := 0; i < len(in); i += 8 {
		switch endianness {
		case BigEndian:
			if wordOrder == HighWordFirst {
				u64 = binary.BigEndian.Uint64(in[i : i+8])
			} else {
				u64 = binary.BigEndian.Uint64(
					[]byte{in[i+6], in[i+7], in[i+4], in[i+5],
						in[i+2], in[i+3], in[i+0], in[i+1]})
			}
		case LittleEndian:
			if wordOrder == LowWordFirst {
				u64 = binary.LittleEndian.Uint64(in[i : i+8])
			} else {
				u64 = binary.LittleEndian.Uint64(
					[]byte{in[i+6], in[i+7], in[i+4], in[i+5],
						in[i+2], in[i+3], in[i+0], in[i+1]})
			}
		}

		out = append(out, u64)
	}

	return
}

func uint64ToBytes(endianness Endianness, wordOrder WordOrder, in uint64) (out []byte) {
	out = make([]byte, 8)

	switch endianness {
	case BigEndian:
		binary.BigEndian.PutUint64(out, in)

		// swap words if needed
		if wordOrder == LowWordFirst {
			out[0], out[1], out[2], out[3], out[4], out[5], out[6], out[7] =
				out[6], out[7], out[4], out[5], out[2], out[3], out[0], out[1]
		}
	case LittleEndian:
		binary.LittleEndian.PutUint64(out, in)

		// swap words if needed
		if wordOrder == HighWordFirst {
			out[0], out[1], out[2], out[3], out[4], out[5], out[6], out[7] =
				out[6], out[7], out[4], out[5], out[2], out[3], out[0], out[1]
		}
	}

	return
}

func bytesToFloat64s(endianness Endianness, wordOrder WordOrder, in []byte) (out []float64) {
	var u64s = bytesToUint64s(endianness, wordOrder, in)

	for _, u64 := range u64s {
		out = append(out, math.Float64frombits(u64))
	}

	return
}

func float64ToBytes(endianness Endianness, wordOrder WordOrder, in float64) (out []byte) {
	out = uint64ToBytes(endianness, wordOrder, math.Float64bits(in))

	return
}

func encodeBools(in []bool) (out []byte) {
	var byteCount uint
	var i uint

	byteCount = uint(len(in)) / 8
	if len(in)%8 != 0 {
		byteCount++
	}

	out = make([]byte, byteCount)
	for i = 0; i < uint(len(in)); i++ {
		if in[i] {
			out[i/8] |= (0x01 << (i % 8))
		}
	}

	return
}

func decodeBools(quantity uint16, in []byte) (out []bool) {
	var i uint
	for i = 0; i < uint(quantity); i++ {
		out = append(out, (((in[i/8] >> (i % 8)) & 0x01) == 0x01))
	}

	return
}

// bytesToInt16s reinterprets each pair of bytes as a signed 16-bit integer.
func bytesToInt16s(endianness Endianness, in []byte) (out []int16) {
	for _, u := range bytesToUint16s(endianness, in) {
		out = append(out, int16(u))
	}

	return
}

// bytesToInt32s reinterprets each group of 4 bytes as a signed 32-bit integer.
func bytesToInt32s(endianness Endianness, wordOrder WordOrder, in []byte) (out []int32) {
	for _, u := range bytesToUint32s(endianness, wordOrder, in) {
		out = append(out, int32(u))
	}

	return
}

// bytesToInt64s reinterprets each group of 8 bytes as a signed 64-bit integer.
func bytesToInt64s(endianness Endianness, wordOrder WordOrder, in []byte) (out []int64) {
	for _, u := range bytesToUint64s(endianness, wordOrder, in) {
		out = append(out, int64(u))
	}

	return
}

// bytesToUint48s decodes groups of 6 bytes (3 registers) into unsigned 48-bit
// values returned as uint64. Endianness controls byte order within each 16-bit
// word; wordOrder controls which word is most significant.
func bytesToUint48s(endianness Endianness, wordOrder WordOrder, in []byte) (out []uint64) {
	var u48 uint64

	for i := 0; i+5 < len(in); i += 6 {
		switch endianness {
		case BigEndian:
			if wordOrder == HighWordFirst {
				// W0 (MSW) … W2 (LSW), each word big-endian.
				u48 = uint64(in[i])<<40 | uint64(in[i+1])<<32 |
					uint64(in[i+2])<<24 | uint64(in[i+3])<<16 |
					uint64(in[i+4])<<8 | uint64(in[i+5])
			} else {
				// W2 (MSW) … W0 (LSW), each word big-endian.
				u48 = uint64(in[i+4])<<40 | uint64(in[i+5])<<32 |
					uint64(in[i+2])<<24 | uint64(in[i+3])<<16 |
					uint64(in[i])<<8 | uint64(in[i+1])
			}
		case LittleEndian:
			if wordOrder == LowWordFirst {
				// W0 (LSW) … W2 (MSW), each word little-endian.
				u48 = uint64(in[i+5])<<40 | uint64(in[i+4])<<32 |
					uint64(in[i+3])<<24 | uint64(in[i+2])<<16 |
					uint64(in[i+1])<<8 | uint64(in[i])
			} else {
				// W0 (MSW) … W2 (LSW), each word little-endian.
				u48 = uint64(in[i+1])<<40 | uint64(in[i])<<32 |
					uint64(in[i+3])<<24 | uint64(in[i+2])<<16 |
					uint64(in[i+5])<<8 | uint64(in[i+4])
			}
		}

		out = append(out, u48)
	}

	return
}

// bytesToInt48s decodes groups of 6 bytes (3 registers) into signed 48-bit
// values returned as int64. The 48-bit result is sign-extended to 64 bits.
func bytesToInt48s(endianness Endianness, wordOrder WordOrder, in []byte) (out []int64) {
	for _, u48 := range bytesToUint48s(endianness, wordOrder, in) {
		// sign-extend bit 47 into the upper 16 bits.
		if u48&(1<<47) != 0 {
			out = append(out, int64(u48|^uint64(0x0000ffffffffffff)))
		} else {
			out = append(out, int64(u48))
		}
	}

	return
}

// bytesToAscii decodes raw register bytes as ASCII text.
// The high byte of each register is the first character, the low byte the second.
// Trailing spaces are stripped from the returned string.
func bytesToAscii(in []byte) string {
	return strings.TrimRight(string(in), " ")
}

// bytesToAsciiReverse decodes raw register bytes as ASCII text with the byte
// order within each 16-bit word reversed: the low byte is the first character
// and the high byte is the second. Trailing spaces are stripped.
func bytesToAsciiReverse(in []byte) string {
	swapped := make([]byte, len(in))

	for i := 0; i+1 < len(in); i += 2 {
		swapped[i], swapped[i+1] = in[i+1], in[i]
	}

	return strings.TrimRight(string(swapped), " ")
}

// bytesToBCD decodes raw register bytes as Binary Coded Decimal (BCD).
// Each byte encodes exactly one decimal digit (value 0–9). Returns a string
// of decimal digits, most-significant digit first.
func bytesToBCD(in []byte) string {
	var sb strings.Builder

	for _, b := range in {
		sb.WriteByte('0' + b%10)
	}

	return sb.String()
}

// bytesToPackedBCD decodes raw register bytes as Packed BCD.
// Each nibble encodes one decimal digit (value 0–9): the high nibble is the
// more-significant digit. Returns a string of decimal digits, most-significant
// digit first.
func bytesToPackedBCD(in []byte) string {
	var sb strings.Builder

	for _, b := range in {
		sb.WriteByte('0' + (b>>4)%10)
		sb.WriteByte('0' + (b&0x0f)%10)
	}

	return sb.String()
}
