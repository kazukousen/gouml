package plantuml

import (
	"bytes"
	"go/token"
	"go/types"
)

// Models ...
type Models []model

func (ms *Models) append(obj *types.TypeName) {
	m := model{obj: obj}
	m.build()
	*ms = append(*ms, m)
}

// WriteTo ...
func (ms Models) WriteTo(buf *bytes.Buffer, ex exists) {
	for _, m := range ms {
		m.writeClass(buf)
		m.writeDiagram(buf, ex)
	}
	ms.writeImplements(buf, 1)
}

func (ms Models) writeImplements(buf *bytes.Buffer, depth int) {
	for _, t := range ms {
		T := t.obj.Type()
		for _, u := range ms {
			U := u.obj.Type()
			if T == U || !types.IsInterface(U) {
				continue
			}
			if types.AssignableTo(T, U) || (!types.IsInterface(T) && types.AssignableTo(types.NewPointer(T), U)) {
				newline(buf, depth)
				buf.WriteString(t.as())
				buf.WriteString(" -up-|> ")
				buf.WriteString(u.as())
			}
		}
	}
}

type model struct {
	obj     *types.TypeName
	id      string
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
	m.id = extractName(typ.String())
	// TODO: obj.IsAlias() is true

	// named type (means user-defined class in OOP)
	if named, _ := typ.(*types.Named); named != nil {

		// implemented methods
		for i := 0; i < named.NumMethods(); i++ {
			f := named.Method(i)
			if isCommand(f) {
				m.kind = modelKindEntity
			}
			m.methods = append(m.methods, method{f: f})
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

	// first-class function
	case *types.Signature:
		f := types.NewFunc(token.NoPos, obj.Pkg(), obj.Name(), un)
		m.methods = append(m.methods, method{f: f})
	}

	if m.kind == "" {
		m.kind = modelKindValueObject
	}
}

func (m model) as() string {
	return m.id
}

func (m model) writeClass(buf *bytes.Buffer) {
	id := m.as()

	newline(buf, 0)
	// package
	buf.WriteString(`package "`)
	buf.WriteString(extractPkgName(id))
	buf.WriteString(`" {`)
	// class
	newline(buf, 1)
	buf.WriteString(m.kind.Printf(extractTypeName(id), id))
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

func (m model) writeDiagram(buf *bytes.Buffer, ex exists) {
	from := m.as()

	newline(buf, 0)
	m.field.writeDiagram(buf, ex, from, 1)

	newline(buf, 0)
	m.methods.writeDiagram(buf, ex, from, 1)

	newline(buf, 0)
	if wrap := m.wrap; wrap != nil {
		to := extractName(wrap.String())
		if _, ok := ex[to]; ok {
			buf.WriteString(from)
			buf.WriteString(" *-- ")
			buf.WriteString(to)
		}
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
		buf.WriteString(exportedIcon(v.Exported()))
		buf.WriteString(v.Name())
		buf.WriteString(": ")
		f.typeString(buf, v.Type())
	}
}

func (f field) typeString(buf *bytes.Buffer, typ types.Type) {
	switch typ := typ.(type) {
	case *types.Struct:
		buf.WriteString("struct{")
		for i := 0; i < typ.NumFields(); i++ {
			if i > 0 {
				buf.WriteString("; ")
			}
			v := typ.Field(i)
			buf.WriteString(v.Name())
			buf.WriteString(": ")
			f.typeString(buf, v.Type())
		}
		buf.WriteString("}")
		return
	}
	buf.WriteString(extractName(typ.String()))
}

func (f field) writeDiagram(buf *bytes.Buffer, ex exists, from string, depth int) {
	if f.st == nil {
		return
	}
	for i := 0; i < f.st.NumFields(); i++ {
		typ := f.st.Field(i).Type()
		if ptr, ok := typ.(*types.Pointer); ok {
			typ = ptr.Elem()
		}
		if m, ok := typ.(*types.Map); ok {
			typ = m.Elem()
		}
		if sl, ok := typ.(*types.Slice); ok {
			typ = sl.Elem()
		}
		to := extractName(typ.String())
		if _, ok := ex[to]; !ok {
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

func exportedIcon(exported bool) string {
	if exported {
		return "+"
	}
	return "-"
}
