# modbus — Modbus Protocol Library

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE.txt)
[![Go Report Card](https://goreportcard.com/badge/github.com/otfabric/modbus)](https://goreportcard.com/report/github.com/otfabric/modbus)
[![CI](https://github.com/otfabric/modbus/actions/workflows/ci.yml/badge.svg)](https://github.com/otfabric/modbus/actions/workflows/ci.yml)
[![Release](https://img.shields.io/badge/release-v0.1.0-blue.svg)](https://github.com/otfabric/modbus/releases)

A production-ready Go implementation of the Modbus application protocol, providing both **client** and **server** capabilities. No C dependencies, no CGo — just Go.

The library exposes a high-level, idiomatic Go API for both client and server roles,
working with native Go types across all supported transports. Every request carries a
`context.Context` for cancellation and deadline propagation.
Advanced features — connection pooling, automatic retries, structured logging, and
metrics hooks — are built in.

> For the complete type signatures, configuration options, and runnable examples see
> **[API.md](API.md)**.

---

## Table of Contents

- [github.com/otfabric/modbus](#githubcomotfabricmodbus)
  - [Table of Contents](#table-of-contents)
  - [Install](#install)
  - [Transport modes](#transport-modes)
  - [Client](#client)
    - [Client supported function codes](#client-supported-function-codes)
    - [Supported Go types](#supported-go-types)
    - [Encoding / byte order](#encoding--byte-order)
  - [Server](#server)
    - [Server supported function codes](#server-supported-function-codes)
  - [Logging](#logging)
  - [Error handling](#error-handling)
  - [Advanced features](#advanced-features)
    - [Retry policy](#retry-policy)
    - [Connection pool](#connection-pool)
    - [Metrics hooks](#metrics-hooks)
  - [CLI client](#cli-client)
  - [Examples](#examples)
  - [Dependencies](#dependencies)
  - [License](#license)

---

## Install

```bash
go get github.com/otfabric/modbus
```

Requires **Go 1.21** or later.

---

## Transport modes

The transport is selected by the `scheme://address` URL in `ClientConfiguration.URL`
or `ServerConfiguration.URL`.

| Scheme | Transport | Client | Server |
|---|---|:---:|:---:|
| `tcp://<host:port>` | Modbus TCP (MBAP) | ✓ | ✓ |
| `tcp+tls://<host:port>` | Modbus TCP over TLS (MBAPS / Modbus Security) | ✓ | ✓ |
| `udp://<host:port>` | Modbus TCP framing over UDP | ✓ | — |
| `rtu://<device>` | Modbus RTU over serial (RS-232 / RS-485) | ✓ | — |
| `rtuovertcp://<host:port>` | Modbus RTU framing tunnelled over TCP | ✓ | — |
| `rtuoverudp://<host:port>` | Modbus RTU framing tunnelled over UDP | ✓ | — |

> **Note:** UDP transports are not part of the official Modbus specification. Both
> MBAP-over-UDP (`udp://`) and RTU-over-UDP (`rtuoverudp://`) are provided because
> different vendors use different framing conventions. When unsure, try both.

---

## Client

### Client supported function codes

All client methods accept a `context.Context` as their first argument and a
`unitId uint8` (slave / unit ID) as their second, enabling per-request deadline and
cancellation control independent of the connection lifecycle.

| FC | Hex | Name | Client method(s) |
|---|---|---|---|
| 01 | 0x01 | Read Coils | `ReadCoil`, `ReadCoils` |
| 02 | 0x02 | Read Discrete Inputs | `ReadDiscreteInput`, `ReadDiscreteInputs` |
| 03 | 0x03 | Read Holding Registers | `ReadRegister`, `ReadRegisters`, `ReadUint16(s)`, `ReadUint16Pair`, `ReadInt16(s)`, `ReadUint32(s)`, `ReadInt32(s)`, `ReadUint48(s)`, `ReadInt48(s)`, `ReadUint64(s)`, `ReadInt64(s)`, `ReadFloat32(s)`, `ReadFloat64(s)`, `ReadAscii`, `ReadAsciiFixed`, `ReadAsciiReverse`, `ReadBCD`, `ReadPackedBCD`, `ReadBytes`, `ReadRawBytes`, `ReadUint8s`, `ReadIPAddr`, `ReadIPv6Addr`, `ReadEUI48` |
| 04 | 0x04 | Read Input Registers | same methods as FC03, passing `InputRegister` |
| 05 | 0x05 | Write Single Coil | `WriteCoil`, `WriteCoilValue` |
| 06 | 0x06 | Write Single Register | `WriteRegister` |
| 15 | 0x0F | Write Multiple Coils | `WriteCoils` |
| 16 | 0x10 | Write Multiple Registers | `WriteRegisters`, `WriteUint32(s)`, `WriteUint64(s)`, `WriteFloat32(s)`, `WriteFloat64(s)`, `WriteBytes`, `WriteRawBytes` |
| 20 | 0x14 | Read File Record | `ReadFileRecords` |
| 21 | 0x15 | Write File Record | `WriteFileRecords` |
| 08 | 0x08 | Diagnostics | `Diagnostics` |
| 11 | 0x11 | Report Server ID | `ReportServerId` |
| 23 | 0x17 | Read/Write Multiple Registers | `ReadWriteMultipleRegisters` |
| 24 | 0x18 | Read FIFO Queue | `ReadFIFOQueue` |
| 43/14 | 0x2B/0x0E | Read Device Identification | `ReadDeviceIdentification`, `ReadAllDeviceIdentification` |

**Device detection:** `HasUnitReadFunction(ctx, unitId, fc)` checks a single read-style FC (FC08, FC43, FC03, FC04, FC01, FC02, FC11, FC18, FC20). `HasUnitIdentifyFunction(ctx, unitId)` checks FC43 (Read Device Identification). **SunSpec discovery:** `DetectSunSpec(ctx, opts)` probes candidate base addresses for the SunSpec "SunS" marker; `ReadSunSpecModelHeaders(ctx, opts, base)` enumerates the model chain (ID and length only); `DiscoverSunSpec(ctx, opts)` does both in one call for fingerprinting and inventory. The library does not decode SunSpec points or schemas — only transport-level detection and model headers. See [API.md § 2.8](API.md#28-modbus-device-detection) and [API.md § 2.9](API.md#29-sunspec-discovery).

### Supported Go types

| Modbus data model | Go types |
|---|---|
| Coils / discrete inputs | `bool`, `[]bool` |
| 16-bit registers | `uint16`, `[]uint16`, `int16`, `[]int16` |
| 32-bit registers (2 × 16-bit) | `uint32`, `[]uint32`, `int32`, `[]int32`, `float32`, `[]float32` |
| 48-bit registers (3 × 16-bit) | `uint64`, `[]uint64` (unsigned), `int64`, `[]int64` (signed) |
| 64-bit registers (4 × 16-bit) | `uint64`, `[]uint64`, `int64`, `[]int64`, `float64`, `[]float64` |
| ASCII string (N × 16-bit) | `string` (trailing spaces stripped) |
| BCD / Packed BCD (N × 16-bit) | `string` (decimal digits) |
| Raw wire bytes | `[]byte` (endianness-aware or unmodified) |
| File records | `[]FileRecordRequest` (read) / `[]FileRecord` (write) |

### Encoding / byte order

`SetEncoding(endianness, wordOrder)` controls how multi-byte and multi-register values
are decoded and encoded. Defaults to `BigEndian, HighWordFirst`. Changes apply to all
subsequent requests on that client instance.

| Setting | Constants | Meaning |
|---|---|---|
| Byte order | `BigEndian` (default), `LittleEndian` | Byte order within each 16-bit register |
| Word order | `HighWordFirst` (default), `LowWordFirst` | Which 16-bit word of a 32/64-bit value sits at the lower register address |

---

## Server

### Server supported function codes

The server dispatches decoded requests to a user-provided `RequestHandler`
implementation. All four handler methods cover the full set of supported function codes:

| FC(s) | Hex | Name | Handler method | `IsWrite` |
|---|---|---|---|---|
| 01 | 0x01 | Read Coils | `HandleCoils` | `false` |
| 02 | 0x02 | Read Discrete Inputs | `HandleDiscreteInputs` | — |
| 03 | 0x03 | Read Holding Registers | `HandleHoldingRegisters` | `false` |
| 04 | 0x04 | Read Input Registers | `HandleInputRegisters` | — |
| 05 | 0x05 | Write Single Coil | `HandleCoils` | `true` |
| 06 | 0x06 | Write Single Register | `HandleHoldingRegisters` | `true` |
| 15 | 0x0F | Write Multiple Coils | `HandleCoils` | `true` |
| 16 | 0x10 | Write Multiple Registers | `HandleHoldingRegisters` | `true` |

Returning a Modbus sentinel error (e.g. `ErrIllegalDataAddress`) causes the server to
send the corresponding exception code back to the client. Any other non-nil error maps
to `ServerDeviceFailure`.

---

## Logging

Both `ClientConfiguration` and `ServerConfiguration` expose a `Logger` field. When
`nil`, the library writes through `slog.Default()` — the Go standard structured logger.

| Constructor | Behaviour |
|---|---|
| `NewStdLogger(l *log.Logger)` | Wraps a stdlib `*log.Logger`; pass `nil` for a default stdout logger |
| `NewSlogLogger(h slog.Handler)` | Wraps any `slog.Handler` (e.g. `slog.NewJSONHandler`, `slog.NewTextHandler`) |
| `NopLogger()` | Discards all output — useful in tests |

The `Logger` interface is straightforward to implement for any custom logging library
(zap, zerolog, logrus, …). See [API.md § 5](API.md#5-logging) for details and an
adapter example.

---

## Error handling

All client methods return a typed `error`. The library defines sentinel errors that can
be tested with `errors.Is`:

| Sentinel | Cause |
|---|---|
| `ErrConfigurationError` | Invalid configuration passed to `NewClient` / `NewServer` |
| `ErrRequestTimedOut` | Request exceeded its deadline or configured timeout |
| `ErrIllegalFunction` | Modbus exception 0x01 |
| `ErrIllegalDataAddress` | Modbus exception 0x02 |
| `ErrIllegalDataValue` | Modbus exception 0x03 |
| `ErrServerDeviceFailure` | Modbus exception 0x04 |
| `ErrAcknowledge` | Modbus exception 0x05 |
| `ErrServerDeviceBusy` | Modbus exception 0x06 |
| `ErrMemoryParityError` | Modbus exception 0x08 |
| `ErrGWPathUnavailable` | Modbus exception 0x0A |
| `ErrGWTargetFailedToRespond` | Modbus exception 0x0B |
| `ErrBadCRC` | RTU CRC mismatch |
| `ErrShortFrame` | Frame too short to decode |
| `ErrProtocolError` | Malformed or unexpected response |
| `ErrBadUnitId` | Response unit ID does not match request |
| `ErrBadTransactionId` | TCP transaction ID mismatch (MBAP) |
| `ErrUnknownProtocolId` | Non-zero MBAP protocol identifier |
| `ErrUnexpectedParameters` | Invalid arguments passed to a client method |
| `ErrSunSpecModelChainInvalid` | Malformed or non-progressing SunSpec model chain |
| `ErrSunSpecModelChainLimitExceeded` | SunSpec model chain exceeded `MaxAddressSpan` |

When the remote device sends a Modbus exception response, the error is additionally
wrapped in `*ExceptionError`, which carries the raw `FunctionCode` and `ExceptionCode`
bytes. `errors.As` gives access to those fields while `errors.Is` still works against
the sentinel:

```go
var excErr *modbus.ExceptionError
if errors.As(err, &excErr) {
    fmt.Printf("fc=0x%02x exception=0x%02x\n", excErr.FunctionCode, excErr.ExceptionCode)
}
if errors.Is(err, modbus.ErrIllegalDataAddress) {
    // address does not exist on this device
}
```

See [API.md § 4](API.md#4-errors) for the full reference.

---

## Advanced features

### Retry policy

Configure automatic retry with exponential back-off on transient errors. The client
re-dials the transport between attempts (or replaces the failed pool connection when
pooling is enabled).

Set `ClientConfiguration.RetryPolicy` to one of the built-in implementations or
provide a custom `RetryPolicy` implementation. See [API.md § 7](API.md#7-retry-policy).

### Connection pool

Set `MaxConns > 1` to enable a bounded connection pool. Multiple goroutines sharing a
single `*ModbusClient` can then execute requests concurrently, each on its own
connection, without serialising on a single TCP socket.

`MinConns` controls how many connections are pre-warmed during `Open()`. Applies to
TCP-based transports only; RTU (serial) always uses a single connection. See
[API.md § 8](API.md#8-connection-pool).

### Metrics hooks

Implement `ClientMetrics` and/or `ServerMetrics` and assign them to the `Metrics`
field of the respective configuration struct. Callbacks (`OnRequest`, `OnResponse`,
`OnError`, `OnTimeout`) fire synchronously on every request outcome and must be
non-blocking. See [API.md § 6](API.md#6-metrics).

---

## CLI client

A command-line Modbus client is included in `cmd/modbus-cli.go`:

```bash
go build -o modbus-cli cmd/modbus-cli.go
./modbus-cli --help
```

---

## Examples

| File | Description |
|---|---|
| [examples/tcp_server.go](examples/tcp_server.go) | Modbus TCP server with an in-memory `RequestHandler` |
| [examples/tls_server.go](examples/tls_server.go) | MBAPS (Modbus over TLS) server with client certificate authentication |
| [examples/tls_client.go](examples/tls_client.go) | MBAPS client with mutual TLS |

For the full public API reference — all types, method signatures, configuration
details, and annotated examples — see **[API.md](API.md)**.

---

## Dependencies

- [github.com/goburrow/serial](https://github.com/goburrow/serial) — serial port access for RTU mode

---

## License

MIT.
