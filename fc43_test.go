package modbus

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"
	"time"
)

func TestReadDeviceIdentification(t *testing.T) {
	var err error
	var ln net.Listener
	var client *ModbusClient
	var di *DeviceIdentification

	ln, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test listener: %v", err)
	}
	defer func() { _ = ln.Close() }()

	go func() {
		var err error
		var sock net.Conn
		var req []byte
		var payload []byte
		var txid []byte
		var unitId byte

		sock, err = ln.Accept()
		if err != nil {
			return
		}
		defer func() { _ = sock.Close() }()

		req = make([]byte, 11)
		_, err = io.ReadFull(sock, req)
		if err != nil {
			return
		}

		if req[2] != 0x00 || req[3] != 0x00 ||
			req[4] != 0x00 || req[5] != 0x05 ||
			req[7] != byte(FCEncapsulatedInterface) ||
			req[8] != byte(MEIReadDeviceIdentification) ||
			req[9] != ReadDeviceIdBasic || req[10] != 0x00 {
			return
		}

		txid = req[0:2]
		unitId = req[6]

		payload = []byte{
			byte(MEIReadDeviceIdentification),
			ReadDeviceIdBasic,
			0x01,
			0x00,
			0x00,
			0x02,
			0x00, 0x03, 'A', 'C', 'M',
			0x01, 0x05, 'P', '1', '2', '3', '4',
		}

		_, _ = sock.Write(append([]byte{
			txid[0], txid[1],
			0x00, 0x00,
			0x00, byte(2 + len(payload)),
			unitId,
			byte(FCEncapsulatedInterface),
		}, payload...))
	}()

	client, err = NewClient(&ClientConfiguration{
		URL:     "tcp://" + ln.Addr().String(),
		Timeout: 1 * time.Second,
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	err = client.Open()
	if err != nil {
		t.Fatalf("failed to open client: %v", err)
	}
	defer func() { _ = client.Close() }()

	di, err = client.ReadDeviceIdentification(context.Background(), 1, ReadDeviceIdBasic, 0x00)
	if err != nil {
		t.Fatalf("ReadDeviceIdentification() should have succeeded, got: %v", err)
	}

	if di.ReadDeviceIdCode != 0x01 || di.ConformityLevel != 0x01 ||
		di.MoreFollows != 0x00 || di.NextObjectId != 0x00 {
		t.Fatalf("unexpected FC43 header fields: %#v", di)
	}

	if len(di.Objects) != 2 {
		t.Fatalf("expected 2 objects, got: %v", len(di.Objects))
	}

	if di.Objects[0].Id != 0x00 || di.Objects[0].Name != "VendorName" || di.Objects[0].Value != "ACM" {
		t.Fatalf("unexpected first object: %#v", di.Objects[0])
	}

	if di.Objects[1].Id != 0x01 || di.Objects[1].Name != "ProductCode" || di.Objects[1].Value != "P1234" {
		t.Fatalf("unexpected second object: %#v", di.Objects[1])
	}
}

func TestReadDeviceIdentificationException(t *testing.T) {
	var err error
	var ln net.Listener
	var client *ModbusClient

	ln, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test listener: %v", err)
	}
	defer func() { _ = ln.Close() }()

	go func() {
		var err error
		var sock net.Conn
		var req []byte
		var txid []byte
		var unitId byte

		sock, err = ln.Accept()
		if err != nil {
			return
		}
		defer func() { _ = sock.Close() }()

		req = make([]byte, 11)
		_, err = io.ReadFull(sock, req)
		if err != nil {
			return
		}

		txid = req[0:2]
		unitId = req[6]

		_, _ = sock.Write([]byte{
			txid[0], txid[1],
			0x00, 0x00,
			0x00, 0x03,
			unitId,
			byte(FCEncapsulatedInterface) | 0x80,
			byte(exIllegalFunction),
		})
	}()

	client, err = NewClient(&ClientConfiguration{
		URL:     "tcp://" + ln.Addr().String(),
		Timeout: 1 * time.Second,
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	err = client.Open()
	if err != nil {
		t.Fatalf("failed to open client: %v", err)
	}
	defer func() { _ = client.Close() }()

	_, err = client.ReadDeviceIdentification(context.Background(), 1, ReadDeviceIdBasic, 0x00)
	if !errors.Is(err, ErrIllegalFunction) {
		t.Fatalf("expected %v, got: %v", ErrIllegalFunction, err)
	}
}

func TestReadDeviceIdentificationRejectsUnexpectedCode(t *testing.T) {
	var err error
	var client *ModbusClient

	client, err = NewClient(&ClientConfiguration{URL: "tcp://127.0.0.1:1"})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	_, err = client.ReadDeviceIdentification(context.Background(), 1, 0x00, 0x00)
	if err != ErrUnexpectedParameters {
		t.Fatalf("expected %v, got: %v", ErrUnexpectedParameters, err)
	}
}

// TestReadAllDeviceIdentification verifies that ReadAllDeviceIdentification requests
// Extended (0x03) and returns all objects the device reports (basic + regular + extended).
func TestReadAllDeviceIdentification(t *testing.T) {
	var err error
	var ln net.Listener
	var client *ModbusClient
	var di *DeviceIdentification

	ln, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test listener: %v", err)
	}
	defer func() { _ = ln.Close() }()

	go func() {
		var err error
		var sock net.Conn
		var req []byte
		var payload []byte
		var txid []byte
		var unitId byte

		sock, err = ln.Accept()
		if err != nil {
			return
		}
		defer func() { _ = sock.Close() }()

		req = make([]byte, 11)
		_, err = io.ReadFull(sock, req)
		if err != nil {
			return
		}

		// ReadAllDeviceIdentification sends readDeviceIdCode 0x03 (Extended), objectId 0x00
		if req[2] != 0x00 || req[3] != 0x00 ||
			req[4] != 0x00 || req[5] != 0x05 ||
			req[7] != byte(FCEncapsulatedInterface) ||
			req[8] != byte(MEIReadDeviceIdentification) ||
			req[9] != ReadDeviceIdExtended || req[10] != 0x00 {
			return
		}

		txid = req[0:2]
		unitId = req[6]

		// Simulate device that supports regular: basic (0x00–0x02) + VendorUrl (0x03), ProductName (0x04)
		payload = []byte{
			byte(MEIReadDeviceIdentification),
			ReadDeviceIdExtended,
			0x02, // conformity level: regular
			0x00, 0x00,
			0x05, // number of objects
			0x00, 0x03, 'A', 'C', 'M',
			0x01, 0x05, 'P', '1', '2', '3', '4',
			0x02, 0x03, '1', '.', '0',
			0x03, 0x09, 'h', 't', 't', 'p', 's', ':', '/', '/', 'x',
			0x04, 0x06, 'M', 'y', 'P', 'r', 'o', 'd',
		}

		_, _ = sock.Write(append([]byte{
			txid[0], txid[1],
			0x00, 0x00,
			0x00, byte(2 + len(payload)),
			unitId,
			byte(FCEncapsulatedInterface),
		}, payload...))
	}()

	client, err = NewClient(&ClientConfiguration{
		URL:     "tcp://" + ln.Addr().String(),
		Timeout: 1 * time.Second,
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	err = client.Open()
	if err != nil {
		t.Fatalf("failed to open client: %v", err)
	}
	defer func() { _ = client.Close() }()

	di, err = client.ReadAllDeviceIdentification(context.Background(), 1)
	if err != nil {
		t.Fatalf("ReadAllDeviceIdentification() should have succeeded, got: %v", err)
	}

	if di.ReadDeviceIdCode != ReadDeviceIdExtended || di.ConformityLevel != 0x02 ||
		di.MoreFollows != 0x00 || di.NextObjectId != 0x00 {
		t.Fatalf("unexpected FC43 header: ReadDeviceIdCode=%v ConformityLevel=%v MoreFollows=%v NextObjectId=%v",
			di.ReadDeviceIdCode, di.ConformityLevel, di.MoreFollows, di.NextObjectId)
	}

	if len(di.Objects) != 5 {
		t.Fatalf("expected 5 objects (basic + regular), got: %v", len(di.Objects))
	}

	if di.Objects[0].Id != 0x00 || di.Objects[0].Name != "VendorName" || di.Objects[0].Value != "ACM" {
		t.Fatalf("object 0: got %#v", di.Objects[0])
	}
	if di.Objects[1].Id != 0x01 || di.Objects[1].Name != "ProductCode" || di.Objects[1].Value != "P1234" {
		t.Fatalf("object 1: got %#v", di.Objects[1])
	}
	if di.Objects[2].Id != 0x02 || di.Objects[2].Name != "MajorMinorRevision" || di.Objects[2].Value != "1.0" {
		t.Fatalf("object 2: got %#v", di.Objects[2])
	}
	if di.Objects[3].Id != 0x03 || di.Objects[3].Name != "VendorUrl" || di.Objects[3].Value != "https://x" {
		t.Fatalf("object 3: got %#v", di.Objects[3])
	}
	if di.Objects[4].Id != 0x04 || di.Objects[4].Name != "ProductName" || di.Objects[4].Value != "MyProd" {
		t.Fatalf("object 4: got %#v", di.Objects[4])
	}
}
