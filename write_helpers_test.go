package modbus

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

// writeMockServer runs a TCP server that accepts FC06 and FC16 and responds with success (echo addr + value/qty).
func writeMockServer(t *testing.T, acceptFC06, acceptFC16 bool) (addr string, cleanup func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	go func() {
		for {
			sock, err := ln.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) {
				defer func() { _ = conn.Close() }()
				for {
					frame, err := readMBAPFrame(conn)
					if err != nil {
						return
					}
					txid, unitId, fc := frame[0:2], frame[6], frame[7]
					if len(frame) < 8 {
						continue
					}
					payload := frame[8:]
					if fc == byte(FCWriteSingleRegister) && acceptFC06 {
						if len(payload) >= 4 {
							_ = writeMBAPNormal(conn, txid, unitId, fc, payload[0:4])
						}
						continue
					}
					if fc == byte(FCWriteMultipleRegisters) && acceptFC16 {
						if len(payload) >= 4 {
							// response: addr (2) + quantity (2)
							_ = writeMBAPNormal(conn, txid, unitId, fc, payload[0:4])
						}
						continue
					}
					_ = writeMBAPException(conn, txid, unitId, fc, byte(exIllegalFunction))
				}
			}(sock)
		}
	}()
	return ln.Addr().String(), func() { _ = ln.Close() }
}

func TestWriteInt16(t *testing.T) {
	addr, cleanup := writeMockServer(t, true, true)
	defer cleanup()
	client, err := NewClient(&ClientConfiguration{URL: "tcp://" + addr, Timeout: 2 * time.Second})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := client.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = client.Close() }()
	if err := client.WriteInt16(context.Background(), 1, 0, -1); err != nil {
		t.Fatalf("WriteInt16: %v", err)
	}
}

func TestWriteInt16s(t *testing.T) {
	addr, cleanup := writeMockServer(t, false, true)
	defer cleanup()
	client, err := NewClient(&ClientConfiguration{URL: "tcp://" + addr, Timeout: 2 * time.Second})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := client.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = client.Close() }()
	if err := client.WriteInt16s(context.Background(), 1, 0, []int16{1, -2, 3}); err != nil {
		t.Fatalf("WriteInt16s: %v", err)
	}
}

func TestWriteInt32(t *testing.T) {
	addr, cleanup := writeMockServer(t, true, true)
	defer cleanup()
	client, err := NewClient(&ClientConfiguration{URL: "tcp://" + addr, Timeout: 2 * time.Second})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := client.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = client.Close() }()
	if err := client.WriteInt32(context.Background(), 1, 0, -123456789); err != nil {
		t.Fatalf("WriteInt32: %v", err)
	}
}

func TestWriteInt32s(t *testing.T) {
	addr, cleanup := writeMockServer(t, false, true)
	defer cleanup()
	client, err := NewClient(&ClientConfiguration{URL: "tcp://" + addr, Timeout: 2 * time.Second})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := client.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = client.Close() }()
	if err := client.WriteInt32s(context.Background(), 1, 0, []int32{1, -1}); err != nil {
		t.Fatalf("WriteInt32s: %v", err)
	}
}

func TestWriteInt48(t *testing.T) {
	addr, cleanup := writeMockServer(t, false, true)
	defer cleanup()
	client, err := NewClient(&ClientConfiguration{URL: "tcp://" + addr, Timeout: 2 * time.Second})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := client.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = client.Close() }()
	if err := client.WriteInt48(context.Background(), 1, 0, 0x123456789ABC); err != nil {
		t.Fatalf("WriteInt48: %v", err)
	}
}

func TestWriteInt48s(t *testing.T) {
	addr, cleanup := writeMockServer(t, false, true)
	defer cleanup()
	client, err := NewClient(&ClientConfiguration{URL: "tcp://" + addr, Timeout: 2 * time.Second})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := client.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = client.Close() }()
	if err := client.WriteInt48s(context.Background(), 1, 0, []int64{1, 2}); err != nil {
		t.Fatalf("WriteInt48s: %v", err)
	}
}

func TestWriteInt64(t *testing.T) {
	addr, cleanup := writeMockServer(t, false, true)
	defer cleanup()
	client, err := NewClient(&ClientConfiguration{URL: "tcp://" + addr, Timeout: 2 * time.Second})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := client.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = client.Close() }()
	if err := client.WriteInt64(context.Background(), 1, 0, -1); err != nil {
		t.Fatalf("WriteInt64: %v", err)
	}
}

func TestWriteInt64s(t *testing.T) {
	addr, cleanup := writeMockServer(t, false, true)
	defer cleanup()
	client, err := NewClient(&ClientConfiguration{URL: "tcp://" + addr, Timeout: 2 * time.Second})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := client.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = client.Close() }()
	if err := client.WriteInt64s(context.Background(), 1, 0, []int64{0, 1}); err != nil {
		t.Fatalf("WriteInt64s: %v", err)
	}
}

func TestWriteAscii(t *testing.T) {
	addr, cleanup := writeMockServer(t, false, true)
	defer cleanup()
	client, err := NewClient(&ClientConfiguration{URL: "tcp://" + addr, Timeout: 2 * time.Second})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := client.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = client.Close() }()
	if err := client.WriteAscii(context.Background(), 1, 0, "Hi"); err != nil {
		t.Fatalf("WriteAscii: %v", err)
	}
}

func TestWriteAsciiFixed(t *testing.T) {
	addr, cleanup := writeMockServer(t, false, true)
	defer cleanup()
	client, err := NewClient(&ClientConfiguration{URL: "tcp://" + addr, Timeout: 2 * time.Second})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := client.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = client.Close() }()
	if err := client.WriteAsciiFixed(context.Background(), 1, 0, "AB "); err != nil {
		t.Fatalf("WriteAsciiFixed: %v", err)
	}
}

func TestWriteAsciiReverse(t *testing.T) {
	addr, cleanup := writeMockServer(t, false, true)
	defer cleanup()
	client, err := NewClient(&ClientConfiguration{URL: "tcp://" + addr, Timeout: 2 * time.Second})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := client.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = client.Close() }()
	if err := client.WriteAsciiReverse(context.Background(), 1, 0, "Hi"); err != nil {
		t.Fatalf("WriteAsciiReverse: %v", err)
	}
}

func TestWriteBCD(t *testing.T) {
	addr, cleanup := writeMockServer(t, false, true)
	defer cleanup()
	client, err := NewClient(&ClientConfiguration{URL: "tcp://" + addr, Timeout: 2 * time.Second})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := client.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = client.Close() }()
	if err := client.WriteBCD(context.Background(), 1, 0, "1234"); err != nil {
		t.Fatalf("WriteBCD: %v", err)
	}
}

func TestWritePackedBCD(t *testing.T) {
	addr, cleanup := writeMockServer(t, false, true)
	defer cleanup()
	client, err := NewClient(&ClientConfiguration{URL: "tcp://" + addr, Timeout: 2 * time.Second})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := client.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = client.Close() }()
	if err := client.WritePackedBCD(context.Background(), 1, 0, "92"); err != nil {
		t.Fatalf("WritePackedBCD: %v", err)
	}
}

func TestWriteUint8s(t *testing.T) {
	addr, cleanup := writeMockServer(t, false, true)
	defer cleanup()
	client, err := NewClient(&ClientConfiguration{URL: "tcp://" + addr, Timeout: 2 * time.Second})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := client.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = client.Close() }()
	if err := client.WriteUint8s(context.Background(), 1, 0, []uint8{0xC0, 0xA8, 0x01, 0x0A}); err != nil {
		t.Fatalf("WriteUint8s: %v", err)
	}
}

func TestWriteIPAddr(t *testing.T) {
	addr, cleanup := writeMockServer(t, false, true)
	defer cleanup()
	client, err := NewClient(&ClientConfiguration{URL: "tcp://" + addr, Timeout: 2 * time.Second})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := client.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = client.Close() }()
	ip := net.IP{192, 168, 1, 10}
	if err := client.WriteIPAddr(context.Background(), 1, 0, ip); err != nil {
		t.Fatalf("WriteIPAddr: %v", err)
	}
}

func TestWriteIPv6Addr(t *testing.T) {
	addr, cleanup := writeMockServer(t, false, true)
	defer cleanup()
	client, err := NewClient(&ClientConfiguration{URL: "tcp://" + addr, Timeout: 2 * time.Second})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := client.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = client.Close() }()
	ip := net.IP{0x20, 0x01, 0x0d, 0xb8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	if err := client.WriteIPv6Addr(context.Background(), 1, 0, ip); err != nil {
		t.Fatalf("WriteIPv6Addr: %v", err)
	}
}

func TestWriteEUI48(t *testing.T) {
	addr, cleanup := writeMockServer(t, false, true)
	defer cleanup()
	client, err := NewClient(&ClientConfiguration{URL: "tcp://" + addr, Timeout: 2 * time.Second})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := client.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = client.Close() }()
	mac := net.HardwareAddr{0x00, 0x1A, 0x2B, 0x3C, 0x4D, 0x5E}
	if err := client.WriteEUI48(context.Background(), 1, 0, mac); err != nil {
		t.Fatalf("WriteEUI48: %v", err)
	}
}

func TestWriteHelpers_InvalidInputs(t *testing.T) {
	addr, cleanup := writeMockServer(t, false, true)
	defer cleanup()
	client, err := NewClient(&ClientConfiguration{URL: "tcp://" + addr, Timeout: 2 * time.Second})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := client.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = client.Close() }()
	ctx := context.Background()

	if err := client.WriteInt16s(ctx, 1, 0, nil); err == nil {
		t.Error("WriteInt16s(nil) should error")
	}
	if err := client.WriteInt16s(ctx, 1, 0, []int16{}); !errors.Is(err, ErrUnexpectedParameters) {
		t.Errorf("WriteInt16s(empty) want ErrUnexpectedParameters, got %v", err)
	}
	if err := client.WriteAscii(ctx, 1, 0, "   "); !errors.Is(err, ErrUnexpectedParameters) {
		t.Errorf("WriteAscii(only spaces) want ErrUnexpectedParameters, got %v", err)
	}
	if err := client.WriteAsciiFixed(ctx, 1, 0, ""); !errors.Is(err, ErrUnexpectedParameters) {
		t.Errorf("WriteAsciiFixed(empty) want ErrUnexpectedParameters, got %v", err)
	}
	if err := client.WriteBCD(ctx, 1, 0, "12a4"); err == nil {
		t.Error("WriteBCD(non-digit) should error")
	}
	if err := client.WritePackedBCD(ctx, 1, 0, "9x"); err == nil {
		t.Error("WritePackedBCD(non-digit) should error")
	}
	if err := client.WriteUint8s(ctx, 1, 0, nil); !errors.Is(err, ErrUnexpectedParameters) {
		t.Errorf("WriteUint8s(nil) want ErrUnexpectedParameters, got %v", err)
	}
	if err := client.WriteIPAddr(ctx, 1, 0, nil); !errors.Is(err, ErrUnexpectedParameters) {
		t.Errorf("WriteIPAddr(nil) want ErrUnexpectedParameters, got %v", err)
	}
	if err := client.WriteEUI48(ctx, 1, 0, nil); !errors.Is(err, ErrUnexpectedParameters) {
		t.Errorf("WriteEUI48(nil) want ErrUnexpectedParameters, got %v", err)
	}
	if err := client.WriteEUI48(ctx, 1, 0, net.HardwareAddr{1, 2, 3}); !errors.Is(err, ErrUnexpectedParameters) {
		t.Errorf("WriteEUI48(short) want ErrUnexpectedParameters, got %v", err)
	}
}
