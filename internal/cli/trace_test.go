package cli

import "testing"

func TestTraceGlobalsNoGlobals(t *testing.T) {
	traceGlobals(map[string]string{"EXISTING": "1"}, map[string]string{})
}
