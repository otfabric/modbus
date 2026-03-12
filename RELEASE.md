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