package gouml

import (
	"bytes"
	"strings"
)

func newline(dst *bytes.Buffer, depth int) {
	dst.WriteString("\n")
	for i := 0; i < depth; i++ {
		dst.WriteString("\t")
	}
}

type id struct {
	full string
}

func (id id) getID() string {
	pkg, name := id.split()
	return pkg + "." + name
}
func (id id) pkg() string {
	pkg, _ := id.split()
	return pkg
}

func (id id) name() string {
	_, name := id.split()
	return name
}

func (id id) split() (string, string) {
	full := id.full
	if strings.Contains(id.full, "/") {
		parts := strings.Split(id.full, "/")
		full = parts[len(parts)-1]
	}
	parts := strings.Split(full, ".")
	var pkg string
	var name string
	switch {
	case len(parts) > 2:
		pkg = parts[len(parts)-2]
		name = parts[len(parts)-1]
	case len(parts) == 2:
		pkg = parts[0]
		name = parts[1]
	}

	return pkg, name
}
