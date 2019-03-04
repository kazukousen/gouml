package gouml

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Parser ...
type Parser struct {
	files           map[string]*ast.File
	typeDefinitions map[string]map[string]*ast.TypeSpec
}

// NewParser ...
func NewParser() *Parser {
	return &Parser{
		files:           map[string]*ast.File{},
		typeDefinitions: map[string]map[string]*ast.TypeSpec{},
	}
}

// Parse ...
func (p *Parser) Parse(baseDir string) error {
	p.readGoFiles(baseDir)

	p.storeTypeSpecs()

	p.print()

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
	for k, ast := range p.files {
		p.storeTypeSpec(k, ast)
	}
}

func (p *Parser) storeTypeSpec(fileName string, astFile *ast.File) {
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
					fmt.Printf("pkg=%s, name=%s\n", pkg, typeSpec.Name.Name)
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

func (p *Parser) parseType(pkgName, typeName string, expr ast.Expr) (*Token, Edges) {
	tokenFieldNames := NewTokenFieldNames()
	edges := Edges{}
	switch expr := expr.(type) {
	// type Foo struct{...}
	case *ast.StructType:
		from := Vertex{Pkg: pkgName, Name: typeName}
		for _, field := range expr.Fields.List {
			for _, name := range field.Names {
				fieldTypeName := p.parsefieldTypeName(field.Type)
				tokenFieldNames.Add(name.Name, fieldTypeName.String())
				pkg, typ, isArray := fieldTypeName.kv()
				if len(pkg) == 0 {
					pkg = pkgName
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
	return NewToken(pkgName, typeName, objKindValueObject, tokenFieldNames), edges
}

type fieldMap map[string]string

func (fm fieldMap) String() string {
	var dst string
	for k, v := range fm {
		dst += fmt.Sprintf("\t%s %s\n", k, v)
	}
	return dst
}

func (p Parser) print() {
	edgess := []Edges{}
	for pkg, types := range p.typeDefinitions {
		for typeName, typeSpec := range types {
			token, edges := p.parseType(pkg, typeName, typeSpec.Type)
			fmt.Println(token.String())
			edgess = append(edgess, edges)
		}
	}
	for _, edges := range edgess {
		fmt.Println(edges.String())
	}
}
