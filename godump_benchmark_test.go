package godump

import (
	"io"
	"testing"
)

func BenchmarkDumpStruct(b *testing.B) {
	type Sample struct {
		ID    int
		Name  string
		Email string
		Meta  map[string]any
	}

	s := Sample{
		ID:    1,
		Name:  "Test User",
		Email: "",
		Meta:  map[string]any{"active": true, "roles": []string{"admin", "user"}},
	}

	d := NewDumper(WithWriter(io.Discard)) // no console output

	b.ReportAllocs() // capture memory allocations

	for i := 0; i < b.N; i++ {
		d.Dump(s)
	}
}
