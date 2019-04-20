package plantuml

import (
	"bytes"
	"go/types"
)

type methods []method

func (ms methods) WriteTo(buf *bytes.Buffer, depth int) {
	for _, m := range ms {
		m.WriteTo(buf, depth)
	}
}

func (ms methods) writeDiagram(buf *bytes.Buffer, ex exists, from string, depth int) {
	for _, m := range ms {
		m.writeDiagram(buf, ex, from, depth)
	}
}

type method struct {
	f *types.Func
}

func (m method) WriteTo(buf *bytes.Buffer, depth int) {
	if m.f == nil {
		return
	}

	newline(buf, depth)
	buf.WriteString(exportedIcon(m.f.Exported()))
	// Name
	buf.WriteString(m.f.Name())

	// Signature
	sig, _ := m.f.Type().(*types.Signature)

	// parameters
	param := sig.Params()
	buf.WriteString("(")
	for i := 0; i < param.Len(); i++ {
		if i > 0 {
			buf.WriteString(", ")
		}
		v := param.At(i)
		name, typ := v.Name(), extractName(v.Type().String())
		buf.WriteString(name)
		buf.WriteString(": ")
		buf.WriteString(typ)
	}
	buf.WriteString(")")

	// results
	res := sig.Results()
	if res.Len() > 0 {
		buf.WriteString(": ")
	}
	if res.Len() > 1 {
		buf.WriteString("(")
	}
	for i := 0; i < res.Len(); i++ {
		if i > 0 {
			buf.WriteString(", ")
		}
		v := res.At(i)
		name, typ := v.Name(), extractName(v.Type().String())
		if name != "" {
			buf.WriteString(name)
			buf.WriteString(": ")
		}
		buf.WriteString(typ)
	}
	if res.Len() > 1 {
		buf.WriteString(")")
	}
}

func (m method) writeDiagram(buf *bytes.Buffer, ex exists, from string, depth int) {
	if m.f == nil {
		return
	}

	if !m.f.Exported() {
		// a non-exported method do not draw a diagram.
		return
	}

	// Signature
	sig, _ := m.f.Type().(*types.Signature)

	// parameters
	param := sig.Params()
	for i := 0; i < param.Len(); i++ {
		typ := param.At(i).Type()
		if ptr, ok := typ.(*types.Pointer); ok {
			typ = ptr.Elem()
		}
		if m, ok := typ.(*types.Map); ok {
			typ = m.Elem()
		}
		if sl, ok := typ.(*types.Slice); ok {
			typ = sl.Elem()
		}
		if _, ok := typ.(*types.Named); !ok {
			continue
		}
		to := extractName(typ.String())
		if _, ok := ex[to]; !ok {
			continue
		}
		newline(buf, depth)
		buf.WriteString(from)
		buf.WriteString(" ..> ")
		buf.WriteString(to)
		buf.WriteString(" : <<use>> ")
	}

	// results
	res := sig.Results()
	for i := 0; i < res.Len(); i++ {
		typ := res.At(i).Type()
		if ptr, ok := typ.(*types.Pointer); ok {
			typ = ptr.Elem()
		}
		if m, ok := typ.(*types.Map); ok {
			typ = m.Elem()
		}
		if sl, ok := typ.(*types.Slice); ok {
			typ = sl.Elem()
		}
		if _, ok := typ.(*types.Named); !ok {
			continue
		}
		to := extractName(typ.String())
		if _, ok := ex[to]; !ok {
			continue
		}
		newline(buf, depth)
		buf.WriteString(from)
		buf.WriteString(" ..> ")
		buf.WriteString(to)
		buf.WriteString(" : <<return>> ")
	}
}
