package modbus

import (
	"context"
	"testing"
)

// typedReadHandler is a minimal Modbus server handler that serves a fixed array
// of 12 holding registers for the typed-read integration tests.
type typedReadHandler struct {
	holding [12]uint16
}

func (h *typedReadHandler) HandleCoils(_ context.Context, _ *CoilsRequest) ([]bool, error) {
	return nil, ErrIllegalFunction
}

func (h *typedReadHandler) HandleDiscreteInputs(_ context.Context, _ *DiscreteInputsRequest) ([]bool, error) {
	return nil, ErrIllegalFunction
}

func (h *typedReadHandler) HandleInputRegisters(_ context.Context, _ *InputRegistersRequest) ([]uint16, error) {
	return nil, ErrIllegalFunction
}

func (h *typedReadHandler) HandleHoldingRegisters(_ context.Context, req *HoldingRegistersRequest) (res []uint16, err error) {
	if req.UnitId != 1 {
		err = ErrIllegalFunction
		return
	}

	if req.Addr+req.Quantity > uint16(len(h.holding)) {
		err = ErrIllegalDataAddress
		return
	}

	res = make([]uint16, req.Quantity)

	for i := range res {
		res[i] = h.holding[int(req.Addr)+i]
	}

	return
}

func startTypedReadServer(t *testing.T, h *typedReadHandler, url string) (*ModbusServer, *ModbusClient) {
	t.Helper()

	var server *ModbusServer
	var client *ModbusClient
	var err error

	server, err = NewServer(&ServerConfiguration{
		URL:        url,
		MaxClients: 1,
	}, h)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	if err = server.Start(); err != nil {
		t.Fatalf("failed to start server: %v", err)
	}

	client, err = NewClient(&ClientConfiguration{URL: url})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	if err = client.Open(); err != nil {
		t.Fatalf("failed to open client: %v", err)
	}

	return server, client
}

// TestReadUint16AndInt16 tests ReadUint16, ReadUint16s, ReadInt16, ReadInt16s.
func TestReadUint16AndInt16(t *testing.T) {
	h := &typedReadHandler{}
	// holding[0] = 0x1234 (positive uint16 / int16)
	// holding[1] = 0xFFFF → uint16: 65535, int16: -1
	// holding[2] = 0x8000 → uint16: 32768, int16: -32768
	h.holding[0] = 0x1234
	h.holding[1] = 0xFFFF
	h.holding[2] = 0x8000

	server, client := startTypedReadServer(t, h, "tcp://localhost:5506")
	defer func() { _ = client.Close(); _ = server.Stop() }()

	ctx := context.Background()

	// --- ReadUint16 ---
	u16, err := client.ReadUint16(ctx, 1, 0x0000, HoldingRegister)
	if err != nil {
		t.Fatalf("ReadUint16: unexpected error: %v", err)
	}
	if u16 != 0x1234 {
		t.Errorf("ReadUint16: expected 0x1234, got 0x%04x", u16)
	}

	// --- ReadUint16s ---
	u16s, err := client.ReadUint16s(ctx, 1, 0x0000, 3, HoldingRegister)
	if err != nil {
		t.Fatalf("ReadUint16s: unexpected error: %v", err)
	}
	if len(u16s) != 3 {
		t.Fatalf("ReadUint16s: expected 3 values, got %d", len(u16s))
	}
	if u16s[0] != 0x1234 || u16s[1] != 0xFFFF || u16s[2] != 0x8000 {
		t.Errorf("ReadUint16s: unexpected values: %v", u16s)
	}

	// --- ReadInt16 ---
	i16, err := client.ReadInt16(ctx, 1, 0x0001, HoldingRegister)
	if err != nil {
		t.Fatalf("ReadInt16: unexpected error: %v", err)
	}
	if i16 != -1 {
		t.Errorf("ReadInt16: expected -1, got %v", i16)
	}

	// --- ReadInt16s ---
	i16s, err := client.ReadInt16s(ctx, 1, 0x0001, 2, HoldingRegister)
	if err != nil {
		t.Fatalf("ReadInt16s: unexpected error: %v", err)
	}
	if len(i16s) != 2 {
		t.Fatalf("ReadInt16s: expected 2 values, got %d", len(i16s))
	}
	if i16s[0] != -1 || i16s[1] != -32768 {
		t.Errorf("ReadInt16s: expected [-1, -32768], got %v", i16s)
	}
}

// TestReadInt32 tests ReadInt32 and ReadInt32s (BigEndian, HighWordFirst).
func TestReadInt32(t *testing.T) {
	h := &typedReadHandler{}
	// holding[0..1] = 0xFFFFFFFF → -1
	h.holding[0] = 0xFFFF
	h.holding[1] = 0xFFFF
	// holding[2..3] = 0x00000001 → 1
	h.holding[2] = 0x0000
	h.holding[3] = 0x0001

	server, client := startTypedReadServer(t, h, "tcp://localhost:5508")
	defer func() { _ = client.Close(); _ = server.Stop() }()

	ctx := context.Background()

	i32, err := client.ReadInt32(ctx, 1, 0x0000, HoldingRegister)
	if err != nil {
		t.Fatalf("ReadInt32: unexpected error: %v", err)
	}
	if i32 != -1 {
		t.Errorf("ReadInt32: expected -1, got %v", i32)
	}

	i32s, err := client.ReadInt32s(ctx, 1, 0x0000, 2, HoldingRegister)
	if err != nil {
		t.Fatalf("ReadInt32s: unexpected error: %v", err)
	}
	if len(i32s) != 2 {
		t.Fatalf("ReadInt32s: expected 2 values, got %d", len(i32s))
	}
	if i32s[0] != -1 || i32s[1] != 1 {
		t.Errorf("ReadInt32s: expected [-1, 1], got %v", i32s)
	}
}

// TestReadInt64 tests ReadInt64 and ReadInt64s (BigEndian, HighWordFirst).
func TestReadInt64(t *testing.T) {
	h := &typedReadHandler{}
	// holding[0..3] = 0xFFFFFFFFFFFFFFFF → -1
	h.holding[0] = 0xFFFF
	h.holding[1] = 0xFFFF
	h.holding[2] = 0xFFFF
	h.holding[3] = 0xFFFF
	// holding[4..7] = 0x0000000000000001 → 1
	h.holding[4] = 0x0000
	h.holding[5] = 0x0000
	h.holding[6] = 0x0000
	h.holding[7] = 0x0001

	server, client := startTypedReadServer(t, h, "tcp://localhost:5510")
	defer func() { _ = client.Close(); _ = server.Stop() }()

	ctx := context.Background()

	i64, err := client.ReadInt64(ctx, 1, 0x0000, HoldingRegister)
	if err != nil {
		t.Fatalf("ReadInt64: unexpected error: %v", err)
	}
	if i64 != -1 {
		t.Errorf("ReadInt64: expected -1, got %v", i64)
	}

	i64s, err := client.ReadInt64s(ctx, 1, 0x0000, 2, HoldingRegister)
	if err != nil {
		t.Fatalf("ReadInt64s: unexpected error: %v", err)
	}
	if len(i64s) != 2 {
		t.Fatalf("ReadInt64s: expected 2 values, got %d", len(i64s))
	}
	if i64s[0] != -1 || i64s[1] != 1 {
		t.Errorf("ReadInt64s: expected [-1, 1], got %v", i64s)
	}
}

// TestReadUint48 tests ReadUint48 and ReadUint48s (BigEndian, HighWordFirst).
func TestReadUint48(t *testing.T) {
	h := &typedReadHandler{}
	// First value: holding[0..2] = W0=0x0001 (MSW), W1=0x0002, W2=0x0003 → 0x000100020003.
	h.holding[0] = 0x0001
	h.holding[1] = 0x0002
	h.holding[2] = 0x0003
	// Second value: holding[3..5] = 0x000400050006.
	h.holding[3] = 0x0004
	h.holding[4] = 0x0005
	h.holding[5] = 0x0006

	server, client := startTypedReadServer(t, h, "tcp://localhost:5512")
	defer func() { _ = client.Close(); _ = server.Stop() }()

	ctx := context.Background()

	u48, err := client.ReadUint48(ctx, 1, 0x0000, HoldingRegister)
	if err != nil {
		t.Fatalf("ReadUint48: unexpected error: %v", err)
	}
	if u48 != 0x000100020003 {
		t.Errorf("ReadUint48: expected 0x000100020003, got 0x%012x", u48)
	}

	u48s, err := client.ReadUint48s(ctx, 1, 0x0000, 2, HoldingRegister)
	if err != nil {
		t.Fatalf("ReadUint48s: unexpected error: %v", err)
	}
	if len(u48s) != 2 {
		t.Fatalf("ReadUint48s: expected 2 values, got %d", len(u48s))
	}
	if u48s[0] != 0x000100020003 || u48s[1] != 0x000400050006 {
		t.Errorf("ReadUint48s: expected [0x000100020003, 0x000400050006], got [0x%012x, 0x%012x]",
			u48s[0], u48s[1])
	}
}

// TestReadInt48 tests ReadInt48 and ReadInt48s (BigEndian, HighWordFirst).
func TestReadInt48(t *testing.T) {
	h := &typedReadHandler{}
	// First value: all 0xFFFF words → 0xFFFFFFFFFFFF → -1.
	h.holding[0] = 0xFFFF
	h.holding[1] = 0xFFFF
	h.holding[2] = 0xFFFF
	// Second value: 0x800000000000 → minimum signed 48-bit.
	h.holding[3] = 0x8000
	h.holding[4] = 0x0000
	h.holding[5] = 0x0000

	server, client := startTypedReadServer(t, h, "tcp://localhost:5514")
	defer func() { _ = client.Close(); _ = server.Stop() }()

	ctx := context.Background()

	i48, err := client.ReadInt48(ctx, 1, 0x0000, HoldingRegister)
	if err != nil {
		t.Fatalf("ReadInt48: unexpected error: %v", err)
	}
	if i48 != -1 {
		t.Errorf("ReadInt48: expected -1, got %v", i48)
	}

	i48s, err := client.ReadInt48s(ctx, 1, 0x0000, 2, HoldingRegister)
	if err != nil {
		t.Fatalf("ReadInt48s: unexpected error: %v", err)
	}
	if len(i48s) != 2 {
		t.Fatalf("ReadInt48s: expected 2 values, got %d", len(i48s))
	}
	const minInt48 = -140737488355328
	if i48s[0] != -1 || i48s[1] != minInt48 {
		t.Errorf("ReadInt48s: expected [-1, %v], got %v", minInt48, i48s)
	}
}

// TestReadAscii tests ReadAscii: high byte of each register = first character.
// "Hello " stored as [0x4865, 0x6C6C, 0x6F20] → "Hello" (trailing space stripped).
func TestReadAscii(t *testing.T) {
	h := &typedReadHandler{}
	h.holding[0] = 0x4865 // 'H','e'
	h.holding[1] = 0x6C6C // 'l','l'
	h.holding[2] = 0x6F20 // 'o',' '

	server, client := startTypedReadServer(t, h, "tcp://localhost:5516")
	defer func() { _ = client.Close(); _ = server.Stop() }()

	s, err := client.ReadAscii(context.Background(), 1, 0x0000, 3, HoldingRegister)
	if err != nil {
		t.Fatalf("ReadAscii: unexpected error: %v", err)
	}
	if s != "Hello" {
		t.Errorf("ReadAscii: expected \"Hello\", got %q", s)
	}
}

// TestReadAsciiReverse tests ReadAsciiReverse: low byte of each register = first character.
// "Hello " stored reversed as [0x6548, 0x6C6C, 0x206F] → "Hello" (trailing space stripped).
func TestReadAsciiReverse(t *testing.T) {
	h := &typedReadHandler{}
	h.holding[0] = 0x6548 // 'e','H' → reversed → 'H','e'
	h.holding[1] = 0x6C6C // 'l','l' → reversed → 'l','l'
	h.holding[2] = 0x206F // ' ','o' → reversed → 'o',' '

	server, client := startTypedReadServer(t, h, "tcp://localhost:5518")
	defer func() { _ = client.Close(); _ = server.Stop() }()

	s, err := client.ReadAsciiReverse(context.Background(), 1, 0x0000, 3, HoldingRegister)
	if err != nil {
		t.Fatalf("ReadAsciiReverse: unexpected error: %v", err)
	}
	if s != "Hello" {
		t.Errorf("ReadAsciiReverse: expected \"Hello\", got %q", s)
	}
}

// TestReadBCD tests ReadBCD: each byte in a register is one decimal digit.
// Registers [0x0102, 0x0304] → bytes [0x01,0x02,0x03,0x04] → "1234".
func TestReadBCD(t *testing.T) {
	h := &typedReadHandler{}
	h.holding[0] = 0x0102 // bytes: 0x01=digit 1, 0x02=digit 2
	h.holding[1] = 0x0304 // bytes: 0x03=digit 3, 0x04=digit 4

	server, client := startTypedReadServer(t, h, "tcp://localhost:5520")
	defer func() { _ = client.Close(); _ = server.Stop() }()

	s, err := client.ReadBCD(context.Background(), 1, 0x0000, 2, HoldingRegister)
	if err != nil {
		t.Fatalf("ReadBCD: unexpected error: %v", err)
	}
	if s != "1234" {
		t.Errorf("ReadBCD: expected \"1234\", got %q", s)
	}
}

// TestReadPackedBCD tests ReadPackedBCD: each nibble is one decimal digit.
// Registers [0x1234, 0x5678] → bytes [0x12,0x34,0x56,0x78] → "12345678".
func TestReadPackedBCD(t *testing.T) {
	h := &typedReadHandler{}
	h.holding[0] = 0x1234 // nibbles: 1, 2, 3, 4
	h.holding[1] = 0x5678 // nibbles: 5, 6, 7, 8

	server, client := startTypedReadServer(t, h, "tcp://localhost:5522")
	defer func() { _ = client.Close(); _ = server.Stop() }()

	s, err := client.ReadPackedBCD(context.Background(), 1, 0x0000, 2, HoldingRegister)
	if err != nil {
		t.Fatalf("ReadPackedBCD: unexpected error: %v", err)
	}
	if s != "12345678" {
		t.Errorf("ReadPackedBCD: expected \"12345678\", got %q", s)
	}
}
