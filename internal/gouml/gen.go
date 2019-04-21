package gouml

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"

	"golang.org/x/xerrors"
)

// Generator ...
type Generator interface {
	Read(path string) error
	ReadDir(baseDir string) error
	OutputFile(out string) error
}

type generator struct {
	parser  Parser
	targets []string
	fset    *token.FileSet
	astPkgs map[string]*ast.Package
	pkgs    []*types.Package
}

// NewGenerator ...
func NewGenerator(parser Parser) Generator {
	return &generator{
		parser:  parser,
		targets: []string{},
		fset:    token.NewFileSet(),
		astPkgs: map[string]*ast.Package{},
		pkgs:    []*types.Package{},
	}
}

func (g generator) OutputFile(filename string) error {
	if err := g.ast(); err != nil {
		return err
	}
	if err := g.check(); err != nil {
		return err
	}

	g.parser.Build(g.pkgs)

	buf := &bytes.Buffer{}
	g.parser.WriteTo(buf)

	uml, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer uml.Close()
	fmt.Fprintf(uml, buf.String())

	fmt.Printf("output to file: %s\n", filename)

	return nil
}

// Read ...
func (g *generator) Read(path string) error {
	if err := g.visit(path, nil, nil); err != nil {
		return err
	}
	return nil
}

// ReadDir ...
func (g *generator) ReadDir(baseDir string) error {
	if err := filepath.Walk(baseDir, g.visit); err != nil {
		return err
	}

	return nil
}

func (g *generator) check() error {
	conf := types.Config{
		Importer: importer.Default(),
		Error: func(err error) {
			// fmt.Printf("error: %+v\n", err)
		},
	}
	for _, astPkg := range g.astPkgs {
		files := []*ast.File{}
		for _, f := range astPkg.Files {
			files = append(files, f)
		}
		pkg, _ := conf.Check(astPkg.Name, g.fset, files, nil)
		g.pkgs = append(g.pkgs, pkg)
	}
	return nil
}

func (g *generator) visit(path string, f os.FileInfo, err error) error {
	if ext := filepath.Ext(path); ext == ".go" {
		path, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		g.targets = append(g.targets, path)
	}
	return nil
}

func (g generator) ast() error {
	for _, path := range g.targets {
		fmt.Printf("parsing AST: %s\n", path)
		src, err := parser.ParseFile(g.fset, path, nil, parser.ParseComments)
		if err != nil {
			return xerrors.Errorf("ParseFile panic: %w", err)
		}
		name := src.Name.Name
		pkg, ok := g.astPkgs[name]
		if !ok {
			pkg = &ast.Package{
				Name:  name,
				Files: make(map[string]*ast.File),
			}
		}
		pkg.Files[path] = src
		g.astPkgs[name] = pkg
	}
	return nil
}
