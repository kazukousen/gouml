package gouml

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Parser ...
type Parser struct {
	files           map[string]*ast.File
	typeDefinitions map[string]map[string]*ast.TypeSpec
	funcDefinitions map[string]map[string][]*ast.FuncDecl
}

// NewParser ...
func NewParser() *Parser {
	return &Parser{
		files:           map[string]*ast.File{},
		typeDefinitions: map[string]map[string]*ast.TypeSpec{},
		funcDefinitions: map[string]map[string][]*ast.FuncDecl{},
	}
}

// Parse ...
func (p *Parser) Parse(baseDir string) error {
	p.readGoFiles(baseDir)

	p.storeTypeSpecs()

	lines := p.print()
	uml, err := os.Create(path.Join(baseDir, "class.uml"))
	if err != nil {
		return err
	}
	defer uml.Close()
	fmt.Fprintf(uml, "%s\n", strings.Join(lines, "\n"))
	fmt.Println("ok")
	return nil
}

// readGoFiles stores AST convert from {.go} file.
func (p Parser) readGoFiles(baseDir string) {
	filepath.Walk(baseDir, p.visit)
}

func (p *Parser) visit(path string, f os.FileInfo, err error) error {
	if err := p.skip(f); err != nil {
		return err
	}
	if ext := filepath.Ext(path); ext == ".go" {
		fset := token.NewFileSet()
		astFile, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			log.Panicf("ParseFile panic: %+v", err)
		}
		p.files[path] = astFile
	}

	return nil
}

// skip returns filepath.SkipDir error if match vendor and hidden directory
func (p Parser) skip(f os.FileInfo) error {
	if f.IsDir() && f.Name() == "vendor" {
		return filepath.SkipDir
	}

	if f.IsDir() && len(f.Name()) > 1 && f.Name()[0] == '.' {
		return filepath.SkipDir
	}
	return nil
}

func (p Parser) storeTypeSpecs() {
	for _, ast := range p.files {
		p.storeTypeSpec(ast)
		p.storeFuncDecl(ast)
	}
}

func (p *Parser) storeFuncDecl(astFile *ast.File) {
	pkg := astFile.Name.Name
	for _, decl := range astFile.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok && funcDecl.Recv != nil {
			for _, field := range funcDecl.Recv.List {
				recv := fmt.Sprintf("%s", field.Type)
				if _, ok := p.typeDefinitions[pkg][recv]; ok {
					if _, ok := p.funcDefinitions[pkg]; !ok {
						p.funcDefinitions[pkg] = map[string][]*ast.FuncDecl{}
					}
					p.funcDefinitions[pkg][recv] = append(p.funcDefinitions[pkg][recv], funcDecl)
				}
			}
		}
	}
}

func (p Parser) parseFuncType(from Vertex, funcDecl *ast.FuncDecl) (string, string) {
	funcName := funcDecl.Name.Name
	ft := FuncToken{Name: funcName}
	edges := FuncEdges{}
	if funcDecl.Type.Params != nil {
		for _, param := range funcDecl.Type.Params.List {
			for _, name := range param.Names {
				ftn := p.parsefieldTypeName(param.Type)
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

	if funcDecl.Type.Results != nil {
		for _, result := range funcDecl.Type.Results.List {
			var name string
			for _, n := range result.Names {
				name = n.Name
			}
			ftn := p.parsefieldTypeName(result.Type)
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

func (p *Parser) storeTypeSpec(astFile *ast.File) {
	pkg := astFile.Name.Name
	for _, decl := range astFile.Decls {
		// decl is a declaration node on top-level
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
			// `type` identifier
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if _, ok := p.typeDefinitions[pkg]; !ok {
						p.typeDefinitions[pkg] = map[string]*ast.TypeSpec{}
					}
					p.typeDefinitions[pkg][typeSpec.Name.Name] = typeSpec
				}
			}
		}
	}
}

func (p Parser) parsefieldTypeName(expr ast.Expr) fieldTypeName {
	switch expr := expr.(type) {
	case *ast.Ident:
		return fieldTypeName(expr.Name)
	case *ast.ArrayType:
		return "[]" + p.parsefieldTypeName(expr.Elt)
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

func (p *Parser) parseType(from Vertex, expr ast.Expr) (string, string) {
	tokenFieldNames := NewTokenFieldNames()
	edges := Edges{}
	switch expr := expr.(type) {
	// type Foo struct{...}
	case *ast.StructType:
		for _, field := range expr.Fields.List {
			for _, name := range field.Names {
				fieldTypeName := p.parsefieldTypeName(field.Type)
				tokenFieldNames.Add(name.Name, fieldTypeName.String())
				pkg, typ, isArray := fieldTypeName.kv()
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
	}
	return NewToken(from.Pkg, from.Name, objKindValueObject, tokenFieldNames).String(), edges.String()
}

type fieldMap map[string]string

func (fm fieldMap) String() string {
	var dst string
	for k, v := range fm {
		dst += fmt.Sprintf("\t%s %s\n", k, v)
	}
	return dst
}

func (p Parser) print() []string {
	lines := []string{}
	es := map[string]struct{}{}
	for pkg, types := range p.typeDefinitions {
		for typeName, typeSpec := range types {
			token, e := p.parseType(Vertex{Pkg: pkg, Name: typeName}, typeSpec.Type)
			lines = append(lines, token)
			es[e] = struct{}{}
		}
	}

	for pkg, types := range p.funcDefinitions {
		for typeName, funcDecls := range types {
			for _, funcDecl := range funcDecls {
				fn, e := p.parseFuncType(Vertex{Pkg: pkg, Name: typeName}, funcDecl)
				lines = append(lines, fmt.Sprintf("%s : %s\n", NewHash(pkg, typeName), fn))
				es[e] = struct{}{}
			}
		}
	}
	for e := range es {
		lines = append(lines, e)
	}

	return lines
}
