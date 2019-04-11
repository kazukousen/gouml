package gouml

import (
	"bytes"
	"fmt"
	"go/types"
)

type models []model

func (ms *models) append(obj *types.TypeName) {
	m := model{obj: obj}
	m.build()
	*ms = append(*ms, m)
}

func (ms models) WriteTo(buf *bytes.Buffer, exists map[id]struct{}) {
	for _, m := range ms {
		m.writeClass(buf)
		m.writeDiagram(buf, exists)
	}
	ms.writeImplements(buf, 1)
}

func (ms models) writeImplements(buf *bytes.Buffer, depth int) {
	for _, t := range ms {
		T := t.obj.Type()
		for _, u := range ms {
			U := u.obj.Type()
			if T == U || !types.IsInterface(U) {
				continue
			}
			if types.AssignableTo(T, U) {
				newline(buf, depth)
				buf.WriteString(t.as())
				buf.WriteString(" --|> ")
				buf.WriteString(u.as())
			}
		}
	}
}

type model struct {
	obj *types.TypeName
	id
	kind    modelKind
	field   field
	methods methods
	wrap    *types.Named
}

func (m *model) build() {
	obj := m.obj
	// *types.TypeName represents ```type [typ] [underlying]```

	// get type
	typ := obj.Type()
	m.full = typ.String()
	// TODO: obj.IsAlias() is true

	// named type (means user-defined class in OOP)
	if named, _ := typ.(*types.Named); named != nil {

		// implemented methods
		for i := 0; i < named.NumMethods(); i++ {
			f := named.Method(i)

			m.methods = append(m.methods, method{f: f})
			if sig, ok := f.Type().(*types.Signature); ok {
				if _, ok := sig.Recv().Type().(*types.Pointer); ok {
					if sig.Results().Len() == 0 {
						m.kind = modelKindEntity
					}
				}
			}
		}
	}

	// underlying
	switch un := typ.Underlying().(type) {
	// struct
	case *types.Struct:
		m.field = field{st: un}

	// interface
	case *types.Interface:
		m.kind = modelKindInterface
		for i := 0; i < un.NumMethods(); i++ {
			m.methods = append(m.methods, method{f: un.Method(i)})
		}

	// wrap
	case *types.Slice:
		if named, _ := un.Elem().(*types.Named); named != nil {
			m.wrap = named
		}
	case *types.Map:
		if named, _ := un.Elem().(*types.Named); named != nil {
			m.wrap = named
		}
	}

	if m.kind == "" {
		m.kind = modelKindValueObject
	}
}

func (m model) as() string {
	return m.id.getID()
}

func (m model) writeClass(buf *bytes.Buffer) {
	id := m.as()

	newline(buf, 0)
	// package
	buf.WriteString(`package "`)
	buf.WriteString(m.pkg())
	buf.WriteString(`" {`)
	// class
	newline(buf, 1)
	buf.WriteString(m.kind.Printf(m.name(), id))
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
}

func (m model) writeDiagram(buf *bytes.Buffer, exists map[id]struct{}) {
	from := m.as()
	newline(buf, 0)
	m.field.writeDiagram(buf, exists, from, 1)

	newline(buf, 0)
	m.methods.writeDiagram(buf, exists, from, 1)

	newline(buf, 0)
	if wrap := m.wrap; wrap != nil {
		to := id{full: wrap.String()}.getID()
		buf.WriteString(from)
		buf.WriteString(" *-- ")
		buf.WriteString(to)
	}
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

func (f field) writeDiagram(buf *bytes.Buffer, exists map[id]struct{}, from string, depth int) {
	if f.st == nil {
		return
	}
	for i := 0; i < f.st.NumFields(); i++ {
		typ := f.st.Field(i).Type()
		if ptr, ok := typ.(*types.Pointer); ok {
			typ = ptr.Elem()
		}
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

func (ms methods) writeDiagram(buf *bytes.Buffer, exists map[id]struct{}, from string, depth int) {
	for _, m := range ms {
		m.writeDiagram(buf, exists, from, depth)
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
	if res.Len() > 1 {
		buf.WriteString(": ")
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

func (m method) writeDiagram(buf *bytes.Buffer, exists map[id]struct{}, from string, depth int) {
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
		if ptr, ok := typ.(*types.Pointer); ok {
			typ = ptr.Elem()
		}
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

type modelKind string

const (
	modelKindInterface   modelKind = `interface "%s" as %s`
	modelKindValueObject modelKind = `class "%s" as %s <<V,Orchid>>`
	modelKindEntity      modelKind = `class "%s" as %s <<E,#FFCC00>>`
)

func (k modelKind) Printf(name, alias string) string {
	return fmt.Sprintf(string(k), name, alias)
}
