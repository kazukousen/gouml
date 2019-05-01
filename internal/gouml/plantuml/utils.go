package plantuml

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

func exportedIcon(exported bool) string {
	if exported {
		return "+"
	}
	return "-"
}

func extractName(full string) string {
	if strings.Contains(full, "/") {
		parts := strings.Split(full, "/")
		full = parts[len(parts)-1]
	}
	return full
}

func extractPkgName(name string) string {
	if strings.Contains(name, ".") {
		parts := strings.Split(name, ".")
		name = parts[len(parts)-2]
	}
	return name
}

func extractTypeName(name string) string {
	if strings.Contains(name, ".") {
		parts := strings.Split(name, ".")
		name = parts[len(parts)-1]
	}
	return name
}

type exists map[string]struct{}
