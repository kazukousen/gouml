package gouml

import (
	"fmt"
	"go/ast"
	"strings"
)

// Generator ...
type Generator struct {
}

// NewGenerator ...
func NewGenerator() *Generator {
	return &Generator{}
}

func (g Generator) printClass(p *Parser, from Vertex, expr ast.Expr) (string, string) {
	fields := NewTokenFieldNames()
	edges := Edges{}
	switch expr := expr.(type) {
	// type Foo struct{...}
	case *ast.StructType:
		for _, field := range expr.Fields.List {
			for _, name := range field.Names {
				ftn := parsefieldTypeName(field.Type)
				fields.Add(name.Name, ftn.String())
				pkg, typ, isArray := ftn.kv()
				if len(pkg) == 0 {
					pkg = from.Pkg
				}
				if _, ok := p.typeDefinitions[pkg][typ]; ok {
					to := Vertex{Pkg: pkg, Name: typ}
					edges = append(edges, Edge{From: from, To: to, IsArray: isArray})
				}
			}
		}
	// type Foo Baz
	case *ast.Ident:
	// type Foo []Baz
	case *ast.ArrayType:
		ftn := parsefieldTypeName(expr.Elt)
		pkg, typ, _ := ftn.kv()
		if len(pkg) == 0 {
			pkg = from.Pkg
		}
		if _, ok := p.typeDefinitions[pkg][typ]; ok {
			to := Vertex{Pkg: pkg, Name: typ}
			edges = append(edges, Edge{From: from, To: to, IsArray: true})
		}
	}
	return NewToken(from.Pkg, from.Name, objKindValueObject, fields).String(), edges.String()
}

func (g Generator) printMethod(p *Parser, from Vertex, funcDecl *ast.FuncDecl) (string, string) {
	funcName := funcDecl.Name.Name
	ft := FuncToken{Name: funcName}
	edges := FuncEdges{}

	// parse Parameters
	if funcDecl.Type.Params != nil {
		for _, param := range funcDecl.Type.Params.List {
			for _, name := range param.Names {
				ftn := parsefieldTypeName(param.Type)
				ft.Params = append(ft.Params, NameTypeKV{Name: name.Name, Type: ftn.String()})
				pkg, typ, isArray := ftn.kv()
				if len(pkg) == 0 {
					pkg = from.Pkg
				}
				if _, ok := p.typeDefinitions[pkg][typ]; ok {
					to := Vertex{Pkg: pkg, Name: typ}
					edges = append(edges, FuncEdge{From: from, To: to, IsArray: isArray})
				}
			}
		}
	}

	// parse Results
	if funcDecl.Type.Results != nil {
		for _, result := range funcDecl.Type.Results.List {
			var name string
			for _, n := range result.Names {
				name = n.Name
			}
			ftn := parsefieldTypeName(result.Type)
			ft.Results = append(ft.Results, NameTypeKV{Name: name, Type: ftn.String()})
			pkg, typ, isArray := ftn.kv()
			if len(pkg) == 0 {
				pkg = from.Pkg
			}
			if _, ok := p.typeDefinitions[pkg][typ]; ok {
				to := Vertex{Pkg: pkg, Name: typ}
				edges = append(edges, FuncEdge{From: from, To: to, IsArray: isArray})
			}
		}
	}

	return ft.String(), edges.String()
}

func (g Generator) printNote(p *Parser, pkgName, text string, from Vertex, specs []*ast.ValueSpec) (string, string) {
	names := []string{}
	edges := map[Edge]struct{}{}
	for _, spec := range specs {
		pkg, typ, _ := parsefieldTypeName(spec.Type).kv()
		if len(pkg) == 0 {
			pkg = pkgName
		}
		edges[Edge{From: from, To: Vertex{Pkg: pkg, Name: typ}}] = struct{}{}
		for _, name := range spec.Names {
			names = append(names, name.Name)
		}
	}

	dst := make(Edges, 0, len(edges))
	for e := range edges {
		dst = append(dst, e)
	}
	return fmt.Sprintf(`
note as %s
	<b>%s</b>
	%s
end note
	`, from.hash(), strings.Trim(text, "\n"), strings.Join(names, "\n\t")), dst.String()
}

func parsefieldTypeName(expr ast.Expr) fieldTypeName {
	switch expr := expr.(type) {
	case *ast.Ident:
		return fieldTypeName(expr.Name)
	case *ast.ArrayType:
		return "[]" + parsefieldTypeName(expr.Elt)
	}
	typeName := fmt.Sprintf("%v", expr)
	typeName = strings.Trim(typeName, "&{}")
	typeName = strings.Replace(typeName, " ", ".", 1)
	return fieldTypeName(typeName)
}

type fieldTypeName string

func (f fieldTypeName) String() string {
	return string(f)
}

func (f fieldTypeName) kv() (string, string, bool) {
	var isArray bool
	s := f.String()
	if strings.Index(s, "[]") != -1 {
		isArray = true
		s = strings.Trim(s, "[]")
	}

	parts := strings.Split(s, ".")
	switch {
	case len(parts) == 1:
		return "", parts[0], isArray
	case len(parts) == 2:
		return parts[0], parts[1], isArray
	}
	return "", "", isArray
}
