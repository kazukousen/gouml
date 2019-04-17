package plantuml

import (
	"bytes"
	"go/types"
	"strings"
)

// Notes ...
type Notes map[*types.Named]Note

func (ns Notes) append(k *types.Named, c *types.Const) {
	note, ok := ns[k]
	if !ok {
		note = []*types.Const{}
	}
	note = append(note, c)
	ns[k] = note
}

// Note ...
type Note []*types.Const

// WriteTo ...
func (ns Notes) WriteTo(buf *bytes.Buffer) {
	newline(buf, 0)
	for named, n := range ns {
		to := extractName(named.String())
		from := "N_" + strings.Replace(to, ".", "_", -1)

		newline(buf, 0)
		buf.WriteString(`package "`)
		buf.WriteString(named.Obj().Pkg().Name())
		buf.WriteString(`" {`)

		// write header
		newline(buf, 1)
		buf.WriteString("note as ")
		buf.WriteString(from)
		// write title
		newline(buf, 2)
		buf.WriteString("<b>")
		buf.WriteString(extractTypeName(to))
		buf.WriteString("</b>\n")

		// write elements
		n.WriteTo(buf, 2)
		// write footer
		newline(buf, 1)
		buf.WriteString("end note")

		newline(buf, 0)
		buf.WriteString(`}`)

		// write relation
		newline(buf, 0)
		buf.WriteString(from)
		buf.WriteString(" --> ")
		buf.WriteString(to)
	}
}

// WriteTo ...
func (n Note) WriteTo(buf *bytes.Buffer, depth int) {
	for _, row := range n {
		newline(buf, depth)
		buf.WriteString(row.Name())
	}
}
