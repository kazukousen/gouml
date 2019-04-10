package gouml

import (
	"bytes"
	"go/types"
)

func newModel(obj *types.TypeName) *model {
	m := &model{obj: obj}
	m.init()
	return m
}

type model struct {
	obj *types.TypeName
	id
	field   field
	methods methods
}

func (m *model) init() {
	obj := m.obj
	// get type
	typ := obj.Type()
	m.full = typ.String()
	if obj.IsAlias() {
		// TODO: test
		typ = obj.Type().Underlying()
	}

	// named type (means user-defined class in OOP)
	if named, _ := typ.(*types.Named); named != nil {

		// implemented methods
		for i := 0; i < named.NumMethods(); i++ {
			m.methods = append(m.methods, method{f: named.Method(i)})
		}
	}

	// struct fields
	if st, _ := typ.Underlying().(*types.Struct); st != nil {
		m.field = field{st: st}
	}
}

func (m model) as() string {
	return m.id.getID()
}

func (m model) writeRelated(buf *bytes.Buffer, exists map[id]struct{}) {
	id := m.as()
	// relation
	newline(buf, 0)
	m.field.relateTo(buf, exists, id, 1)
	newline(buf, 0)
	m.methods.relateTo(buf, exists, id, 1)
}

func (m model) String() string {
	buf := &bytes.Buffer{}

	id := m.as()

	// package
	buf.WriteString(`package "`)
	buf.WriteString(m.pkg())
	buf.WriteString(`" {`)
	// class
	newline(buf, 1)
	buf.WriteString(`class "`)
	buf.WriteString(m.name())
	buf.WriteString(`" as `)
	buf.WriteString(id)
	if m.field.size() > 0 || len(m.methods) > 0 {
		buf.WriteString(` {`)
		// fields
		m.field.WriteTo(buf, 2)
		// methods
		m.methods.WriteTo(buf, 2)
		newline(buf, 1)
		buf.WriteString("}")
	}
	newline(buf, 0)
	buf.WriteString("}")

	return buf.String()
}

type field struct {
	st *types.Struct
}

func (f field) size() int {
	if f.st == nil {
		return 0
	}
	return f.st.NumFields()
}

func (f field) WriteTo(buf *bytes.Buffer, depth int) {
	if f.st == nil {
		return
	}
	for i := 0; i < f.st.NumFields(); i++ {
		newline(buf, depth)
		v := f.st.Field(i)
		buf.WriteString(v.Name())
		buf.WriteString(": ")
		buf.WriteString(v.Type().String())
	}
}

func (f field) relateTo(buf *bytes.Buffer, exists map[id]struct{}, from string, depth int) {
	if f.st == nil {
		return
	}
	for i := 0; i < f.st.NumFields(); i++ {
		typ := f.st.Field(i).Type()
		id := id{full: typ.String()}
		to := id.getID()
		if _, ok := exists[id]; !ok {
			continue
		}
		newline(buf, depth)
		buf.WriteString(from)
		buf.WriteString(" --> ")
		buf.WriteString(to)
	}
}

type methods []method

func (ms methods) WriteTo(buf *bytes.Buffer, depth int) {
	for _, m := range ms {
		newline(buf, depth)
		m.WriteTo(buf)
	}
}

func (ms methods) relateTo(buf *bytes.Buffer, exists map[id]struct{}, from string, depth int) {
	for _, m := range ms {
		m.relateTo(buf, exists, from, depth)
	}
}

type method struct {
	f *types.Func
}

func (m method) WriteTo(buf *bytes.Buffer) {
	if m.f == nil {
		return
	}

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
		name, typ := v.Name(), v.Type().String()
		buf.WriteString(name)
		buf.WriteString(": ")
		buf.WriteString(typ)
	}
	buf.WriteString(")")

	// results
	res := sig.Results()
	buf.WriteString(": ")
	if res.Len() > 1 {
		buf.WriteString("(")
	}
	for i := 0; i < res.Len(); i++ {
		if i > 0 {
			buf.WriteString(", ")
		}
		v := res.At(i)
		name, typ := v.Name(), v.Type().String()
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

func (m method) relateTo(buf *bytes.Buffer, exists map[id]struct{}, from string, depth int) {
	if m.f == nil {
		return
	}
	// Signature
	sig, _ := m.f.Type().(*types.Signature)

	// parameters
	param := sig.Params()
	for i := 0; i < param.Len(); i++ {
		typ := param.At(i).Type()
		id := id{full: typ.String()}
		to := id.getID()
		if _, ok := exists[id]; !ok {
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
		id := id{full: typ.String()}
		to := id.getID()
		if _, ok := exists[id]; !ok {
			continue
		}
		newline(buf, depth)
		buf.WriteString(from)
		buf.WriteString(" ..> ")
		buf.WriteString(to)
		buf.WriteString(" : <<return>> ")
	}
}
