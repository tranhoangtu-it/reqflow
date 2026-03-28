package driven

// TraceSetter is an optional interface that HTTP clients can implement
// to enable or disable detailed timing instrumentation at runtime.
type TraceSetter interface {
	SetTrace(enabled bool)
}
