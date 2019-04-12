package gouml

import (
	"bytes"
	"go/types"
	"strings"
)

// Notes ...
type Notes map[id]Note

// Note ...
type Note []*types.Const

// WriteTo ...
func (ns Notes) WriteTo(buf *bytes.Buffer) {
	newline(buf, 0)
	for id, n := range ns {
		// generate id
		to := id.getID()
		from := "N_" + strings.Replace(to, ".", "_", -1)

		// write header
		newline(buf, 0)
		buf.WriteString("note as ")
		buf.WriteString(from)
		{
			// write title
			newline(buf, 1)
			buf.WriteString("<b>")
			buf.WriteString(id.name())
			buf.WriteString("</b>\n")

			// write elements
			n.WriteTo(buf, 1)
		}
		// write footer
		newline(buf, 0)
		buf.WriteString("end note")

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
