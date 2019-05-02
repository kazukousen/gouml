package gouml_test

import (
	"testing"

	"github.com/kazukousen/gouml"
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

	got := gouml.Compress(src)
	if got != want {
		t.Errorf("\ngot %s\nwant %s\n", got, want)
	}
}
