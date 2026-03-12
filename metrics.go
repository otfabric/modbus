package modbus

import "time"

// ClientMetrics is an optional callback interface for observing client-side request
// outcomes. All methods are called synchronously in the goroutine executing the
// request; implementations must be non-blocking (e.g. increment an atomic counter,
// send on a buffered channel). A nil ClientMetrics is valid and disables collection.
type ClientMetrics interface {
	// OnRequest is called immediately before the first attempt to send a request.
	// unitId and functionCode identify the target device and operation.
	OnRequest(unitId uint8, functionCode FunctionCode)

	// OnResponse is called after a successful round-trip.
	// duration covers the total elapsed time including any retry delays.
	OnResponse(unitId uint8, functionCode FunctionCode, duration time.Duration)

	// OnError is called when a request ultimately fails with a non-timeout error
	// (after all retry attempts are exhausted).
	// duration covers the total elapsed time including any retry delays.
	OnError(unitId uint8, functionCode FunctionCode, duration time.Duration, err error)

	// OnTimeout is called when a request ultimately fails because it exceeded
	// its deadline (errors.Is(err, ErrRequestTimedOut) or context deadline exceeded).
	// duration covers the total elapsed time including any retry delays.
	OnTimeout(unitId uint8, functionCode FunctionCode, duration time.Duration)
}

// ServerMetrics is an optional callback interface for observing server-side request
// outcomes. All methods are called synchronously; implementations must be non-blocking.
// A nil ServerMetrics is valid and disables collection.
type ServerMetrics interface {
	// OnRequest is called immediately before the handler is invoked.
	// unitId and functionCode identify the incoming request.
	OnRequest(unitId uint8, functionCode FunctionCode)

	// OnResponse is called after the handler returns without error.
	// duration is the handler execution time.
	OnResponse(unitId uint8, functionCode FunctionCode, duration time.Duration)

	// OnError is called when the handler returns an error.
	// duration is the handler execution time.
	OnError(unitId uint8, functionCode FunctionCode, duration time.Duration, err error)
}
