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
	"path"
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
	fset    *token.FileSet
	astPkgs map[string]*ast.Package
	pkgs    []*types.Package
}

// NewGenerator ...
func NewGenerator(parser Parser) Generator {
	return &generator{
		parser:  parser,
		fset:    token.NewFileSet(),
		astPkgs: map[string]*ast.Package{},
		pkgs:    []*types.Package{},
	}
}

func (g generator) OutputFile(out string) error {
	g.check()

	g.parser.Build(g.pkgs)

	buf := &bytes.Buffer{}
	g.parser.WriteTo(buf)

	filename := path.Join(out + ".uml")
	uml, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer uml.Close()
	fmt.Fprintf(uml, buf.String())

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
		Importer: importer.ForCompiler(g.fset, "source", nil),
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
