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
