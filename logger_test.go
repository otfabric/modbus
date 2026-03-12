package modbus

import (
	"bytes"
	"log"
	"testing"
)

func TestClientCustomLogger(t *testing.T) {
	var buf bytes.Buffer
	var stdl = log.New(&buf, "external-prefix: ", 0)

	_, _ = NewClient(&ClientConfiguration{
		Logger: NewStdLogger(stdl),
		URL:    "sometype://sometarget",
	})

	if buf.String() != "external-prefix: modbus-client(sometarget) [error]: unsupported client type 'sometype'\n" {
		t.Errorf("unexpected logger output '%s'", buf.String())
	}
}

func TestServerCustomLogger(t *testing.T) {
	var buf bytes.Buffer
	var stdl = log.New(&buf, "external-prefix: ", 0)

	_, _ = NewServer(&ServerConfiguration{
		Logger: NewStdLogger(stdl),
		URL:    "tcp://",
	}, nil)

	if buf.String() != "external-prefix: modbus-server() [error]: missing host part in URL 'tcp://'\n" {
		t.Errorf("unexpected logger output '%s'", buf.String())
	}
}
