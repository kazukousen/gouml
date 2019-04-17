package plantuml

import (
	"bytes"
	"go/types"

	"github.com/kazukousen/gouml/internal/gouml"
)

// NewParser ...
func NewParser() gouml.Parser {
	return &parser{
		models: Models{},
		notes:  Notes{},
		ex:     exists{},
	}
}

type parser struct {
	models Models
	notes  Notes
	ex     exists
}

func (p *parser) Build(pkgs []*types.Package) {
	objects := []types.Object{}
	for _, pkg := range pkgs {
		scope := pkg.Scope()
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)
			objects = append(objects, obj)

			if obj.Pkg().Name() == pkg.Name() {
				if named, _ := obj.Type().(*types.Named); named != nil {
					p.ex[extractName(named.String())] = struct{}{}
				}
			}
		}
	}

	for _, obj := range objects {
		switch obj := obj.(type) {

		// declared type
		case *types.TypeName:
			p.models.append(obj)

		// declared constant
		case *types.Const:
			if named, _ := obj.Type().(*types.Named); named != nil {
				p.notes.append(named, obj)
			}
		}
	}
}

func (p parser) WriteTo(buf *bytes.Buffer) {
	p.models.WriteTo(buf, p.ex)
	p.notes.WriteTo(buf)
	newline(buf, 0)
}
