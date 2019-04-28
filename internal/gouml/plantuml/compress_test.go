package plantuml_test

import (
	"testing"

	"github.com/kazukousen/gouml/internal/gouml/plantuml"
)

func TestCompress(t *testing.T) {
	src := `
	class Foo {
		Bar: Bar
	}

	class Bar

	Foo --> Bar
	`

	want := `UDhYIiv9B2vMSClFLwZcSaeiib9mIYpYgkM2YeCuN219NLqxC0SG003__rvv3QC0`

	got := plantuml.Compress(src)
	if got != want {
		t.Errorf("\ngot %s\nwant %s\n", got, want)
	}
}
