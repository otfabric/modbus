# Release v0.3.0

**Date:** 2026-03-15
**Previous release:** v0.2.5

## Summary

Introduce a **codec-first API** for typed register read/write with explicit layout and discovery. Codecs own interpretation; transport remains register-native. The library continues to offer legacy helpers (e.g. `ReadUint32`, `WriteFloat64`) with `SetEncoding`; the codec path uses explicit `RegisterLayout` and does not use `SetEncoding`.

## Changes

### Added

- **Codec interfaces** — `Decoder[T]`, `Encoder[T]`, and `Codec[T]` with `ID()`, `Name()`, `RegisterSpec()`, `ByteSpec()`, `DecodeRegisters`, `EncodeRegisters`. All codec instances are fixed-width (parameterized at construction).
- **RegisterLayout** — Immutable layout describing byte order across registers (1-based positions: 1 = LSB). `NewRegisterLayout`, `MustNewRegisterLayout`, getters `RegisterCount()`, `BytePositions()`, `String()`. Common vars: `Layout16_21`, `Layout16_12`, `Layout32_4321`, `Layout32_2143`, `Layout48_654321`, `Layout48_214365`, `Layout64_87654321`, `Layout64_21436587`.
- **Transport** — Package-level generic `ReadWithCodec[T](mc, ctx, unitID, addr, regType, codec)` and `WriteWithCodec[T](mc, ctx, unitID, addr, value, codec)`. Wire order (big-endian per register); codec owns layout. Convenience: `ReadUint32WithLayout`, `WriteUint32WithLayout`.
- **Numeric codecs** — `New*Codec(layout)` and `MustNew*Codec(layout)` for uint16, int16, uint32, int32, uint48, int48, uint64, int64, float32, float64. Constructors validate layout and return `(Codec[T], error)` or panic for `Must`.
- **Text codecs** — `NewAsciiCodec`, `NewAsciiFixedCodec`, `NewAsciiReverseCodec`, `NewBCDCodec`, `NewPackedBCDCodec` (register count = width). Full ASCII validation on encode; overlong input truncated; BCD truncation keeps rightmost digits.
- **Bytes and network codecs** — `NewBytesCodec(byteCount)`, `NewUint8SliceCodec(byteCount)` (even byte count required); `NewIPAddrCodec()`, `NewIPv6AddrCodec()` (IPv6 rejects IPv4), `NewEUI48Codec()`.
- **Offline helpers** — `DecodeRegisters`, `EncodeRegisters`, `ValidateRegisterSpec(spec, regs, codecID)`, `ValidateByteSpec(spec, b, codecID)` for tests and tooling.
- **Discovery (registry)** — `CodecDescriptor`, `CodecCandidate`, `CodecQuery`. `AvailableCodecDescriptors()`, `CodecDescriptorsForRegisterCount`, `CodecDescriptorsForByteCount`, `CodecDescriptorByID`, `CodecCandidatesForRegisterCount`, `CodecCandidatesForByteCount`, `FindCodecDescriptors`. Returned descriptors are deep-copied. Discovery exposes a curated subset of common widths; constructors accept any valid width.
- **Codec errors** — Sentinels: `ErrCodecRegisterCount`, `ErrCodecLayout`, `ErrCodecValue`, `ErrEncodingError`. Typed: `*CodecRegisterCountError`, `*CodecLayoutError`, `*CodecByteCountError`, `*CodecValueError` (all unwrap to the appropriate sentinel). `ReadWithCodec` returns `*CodecRegisterCountError` when `spec.Count == 0`.

### Unchanged

- Legacy read/write helpers and `SetEncoding` unchanged. Codec API is additive. SunSpec discovery, bitfield ops, and all existing client/server behaviour unchanged.

---

# Release v0.2.5

**Date:** 2026-03-14
**Previous release:** v0.2.4

## Summary

Add bitfield and masked-register operations for devices that expose booleans and enums inside holding or input registers (status bits, alarm words, control words, mode enums). Read single or multiple bits from a register; write one bit or update a masked field without clobbering adjacent bits.

## Changes

### Added

- **ReadRegisterBit(ctx, unitId, addr, bitIndex, regType)** — Reads one register (FC03/FC04) and returns the bit at `bitIndex` (0 = LSB, 15 = MSB). Supports both holding and input registers.
- **ReadRegisterBits(ctx, unitId, addr, bitIndex, count, regType)** — Reads one register and returns `count` bits (1–16) starting at `bitIndex`. Use for multi-bit mode enums.
- **WriteRegisterBit(ctx, unitId, addr, bitIndex, value)** — Read-modify-write: reads holding register, sets or clears one bit, writes back (FC03 + FC16). Other bits unchanged.
- **UpdateRegisterMask(ctx, unitId, addr, mask, value)** — Read-modify-write: `newVal = (old & ^mask) | (value & mask)`. Only bits set in `mask` are updated; use for control words without affecting adjacent bits.

Invalid `bitIndex` (> 15) or invalid `ReadRegisterBits` range returns `ErrUnexpectedParameters`.

### Unchanged

- Coils and discrete inputs unchanged. New methods are additive.

---

# Release v0.2.4

**Date:** 2026-03-14
**Previous release:** v0.2.3

## Summary

Add typed write helpers that mirror the existing read helpers: signed integers (Int16/32/48/64), ASCII (normal, fixed-width, reverse), BCD and packed BCD, and raw/address types (Uint8s, IPAddr, IPv6Addr, EUI48). All use FC16 (Write Multiple Registers) with the same encoding conventions as the corresponding read methods.

## Changes

### Added

- **Signed integer writes** — `WriteInt16`, `WriteInt16s`, `WriteInt32`, `WriteInt32s`, `WriteInt48`, `WriteInt48s`, `WriteInt64`, `WriteInt64s`. Encoding follows `SetEncoding`; empty slice returns `ErrUnexpectedParameters`.
- **ASCII writes** — `WriteAscii` (trim trailing spaces, same layout as ReadAscii), `WriteAsciiFixed` (no trim), `WriteAsciiReverse` (same layout as ReadAsciiReverse).
- **BCD writes** — `WriteBCD` (one byte per digit), `WritePackedBCD` (two digits per byte; odd byte count padded for register alignment). Non-digit characters return an error.
- **Raw and address writes** — `WriteUint8s` (raw bytes, no reordering), `WriteIPAddr` (4 bytes from `net.IP.To4()`), `WriteIPv6Addr` (16 bytes), `WriteEUI48` (6 bytes from `net.HardwareAddr`). Invalid input returns `ErrUnexpectedParameters`.
- **Encoding helpers** (internal) — `uint48ToBytes`, `asciiToBytes`, `asciiToBytesReverse`, `bcdToBytes`, `packedBCDToBytes` for use by the write methods.

### Unchanged

- Existing write and read behaviour unchanged. New methods are additive.

---

# Release v0.2.3

**Date:** 2026-03-12
**Previous release:** v0.2.2

## Summary

Align the library with common Modbus/TCP and Wireshark dissector behaviour: spec-compliant MBAP length validation, standard port constants, additional function-code coverage, optional transaction-ID diagnostics, and clearer protocol error reporting.

## Changes

### Added

- **Standard port constants** — `PortModbusTCP` (502) and `PortModbusTLS` (802) for use in URLs or documentation. Modbus RTU over TCP has no standard port.
- **MBAP length validation** — TCP transport rejects MBAP length &lt; 2 or &gt; 254 and returns an error wrapping `ErrInvalidMBAPLength` (received length included in the message). Validation applied on both receive and send.
- **Function codes** — `FCReadExceptionStatus` (0x07), `FCGetCommEventCounters` (0x0B), `FCGetCommEventLog` (0x0C) added to known FCs and `KnownFunctionCodes()`. FC07 supported in RTU response length handling.
- **LastTransactionID()** — Client method returns the MBAP transaction ID of the last successful TCP response (0 for RTU/non-TCP). Useful for diagnostics and correlating with packet captures.
- **RTU PDU length rules** — Comment block in `expectedResponseLenth` documents response length rules per FC for spec/dissector alignment.

### Changed

- **TCP receive** — Frames with invalid MBAP length now return `ErrInvalidMBAPLength` (with value) instead of generic `ErrProtocolError`; log message includes expected range 2–254.
- **TCP send** — Requests that would produce MBAP length &gt; 254 are rejected before send with `ErrInvalidMBAPLength`.

### Unchanged

- All existing client/server behaviour and API contracts unchanged. New constants and `LastTransactionID()` are additive.

---

# Release v0.2.2

**Date:** 2026-03-12
**Previous release:** v0.2.1

## Summary

Export SunSpec protocol constants so downstream consumers (e.g. strategies parsing raw `ScanResult.Data`) can reference the canonical marker, end-model sentinel, and default base address values directly instead of maintaining mirrored copies.

## Changes

### Changed

- **SunSpec constants** — The following previously-unexported values are now exported:
  - `SunSpecMarkerReg0` (`0x5375`) / `SunSpecMarkerReg1` (`0x6E53`) — "SunS" marker registers.
  - `SunSpecEndModelID` (`0xFFFF`) / `SunSpecEndModelLength` (`0`) — end-of-chain sentinel.
  - `SunSpecDefaultBaseAddresses` (`[]uint16{0, 40000, 50000, 1, 39999, 40001, 49999, 50001}`) — default probe addresses.

### Unchanged

- All SunSpec discovery methods, types, and behaviour unchanged. This is a purely additive API change.

---

# Release v0.2.1

**Date:** 2026-03-12
**Previous release:** v0.2.0

## Summary

Relax SunSpec discovery **UnitID** handling so the full range **0–255** is accepted. SunSpec-enabled devices behind a Modbus gateway may use unit IDs outside the classic 1–247 range; validation no longer rejects them.

## Changes

### Changed

- **SunSpec options** — Removed UnitID range check (was 1–247). `SunSpecOptions.UnitID` now accepts 0–255. Zero still defaults to 1 for scanner ergonomics. Docs and API comments updated; invalid-options text no longer mentions UnitID.

### Unchanged

- All other SunSpec and FC03/FC04 helper behaviour unchanged.

---

# Release v0.2.0

**Date:** 2026-03-12
**Previous release:** v0.1.0

## Summary

Add minimal, transport-level **SunSpec discovery** APIs so callers can detect SunSpec devices, discover the SunSpec map base address, and enumerate **SunSpec model headers** (not full model metadata: no point decoding, names, or schema) without external SunSpec JSON or schema. These APIs are **read-only discovery helpers** and do not modify device state. Intended for device fingerprinting, protocol detection, and as a foundation for higher-level SunSpec libraries.

Default probe addresses are the official protocol candidates **0, 40000, 50000**, plus adjacent compatibility offsets (**1, 39999, 40001, 49999, 50001**) to tolerate 0-based vs 1-based addressing confusion found in vendor documentation and tooling. Reaching **MaxModels** stops enumeration and returns the models collected so far **without error** (normal truncation).

## Changes

### Added

- **SunSpec discovery (client)**  
  - `DetectSunSpec(ctx, opts)` — Probes candidate base addresses for the "SunS" marker; returns a structured result. "Not SunSpec" is not an error (`Detected: false`, `error == nil`). Uses the same request path as other client methods (lock per read, retries, metrics).
  - `ReadSunSpecModelHeaders(ctx, opts, baseAddress)` — Walks the model chain from `baseAddress+2`, returning model ID, length, and address ranges. Stops at end model (0xFFFF/0) or when guards trigger. Reaching MaxModels returns collected models without error. Malformed or non-progressing chains return partial results plus `ErrSunSpecModelChainInvalid`; exceeding `MaxAddressSpan` returns `ErrSunSpecModelChainLimitExceeded`. Invalid options (unsupported RegType, empty BaseAddresses) return `ErrUnexpectedParameters`.
  - `DiscoverSunSpec(ctx, opts)` — Convenience: runs detection then model-header enumeration; returns combined result. Includes partial model results when the chain read fails partway.
- **Types:** `SunSpecOptions`, `SunSpecProbeAttempt`, `SunSpecDetectionResult`, `SunSpecModelHeader`, `SunSpecDiscoveryResult`.
- **Sentinels:** `ErrSunSpecModelChainInvalid`, `ErrSunSpecModelChainLimitExceeded`.
- UnitID zero defaults to 1 for scanner ergonomics (documented tradeoff).
- **FC03/FC04 convenience read helpers** — Generic read helpers usable for SunSpec and other fixed-field protocols (no SunSpec-specific logic):
  - `ReadUint16Pair` — Exactly two registers as `[2]uint16`.
  - `ReadAsciiFixed` — Same ASCII layout as `ReadAscii` but trailing spaces preserved.
  - `ReadUint8s` — Raw bytes in wire order (no `SetEncoding`).
  - `ReadIPAddr` — 4 bytes as IPv4 `net.IP`.
  - `ReadIPv6Addr` — 16 bytes as IPv6 `net.IP`.
  - `ReadEUI48` — 6 bytes as MAC/EUI-48 `net.HardwareAddr`.
  Address and byte helpers use raw wire order and are unaffected by `SetEncoding`.

### Unchanged

- No point decoding, scale factors, or schema-driven parsing; no JSON model definitions. SunSpec semantics remain the responsibility of a separate SunSpec library.

---

# Release v0.1.0

**Date:** 2026-03-12
**Previous release:** v0.0.0

## Summary

Initial release.

---