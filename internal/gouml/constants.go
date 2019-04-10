package gouml

import (
	"bytes"
	"go/types"
	"strings"
)

type constantsMap map[id]constants

type constants []*types.Const

func (m constantsMap) WriteTo(buf *bytes.Buffer) {
	newline(buf, 0)
	for id, cs := range m {
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
			cs.WriteTo(buf, 1)
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

func (cs constants) WriteTo(buf *bytes.Buffer, depth int) {
	for _, c := range cs {
		newline(buf, depth)
		buf.WriteString(c.Name())
	}
}
