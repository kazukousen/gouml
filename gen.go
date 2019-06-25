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
	"strings"

	"golang.org/x/xerrors"
)

// Generator ...
type Generator interface {
	Read(path string) error
	ReadDir(baseDir string) error
	WriteTo(buf *bytes.Buffer) error
}

type generator struct {
	parser  Parser
	targets []string
	fset    *token.FileSet
	astPkgs map[string]*ast.Package
	pkgs    []*types.Package
	isDebug bool
}

// NewGenerator ...
func NewGenerator(parser Parser, isDebug bool) Generator {
	return &generator{
		parser:  parser,
		targets: []string{},
		fset:    token.NewFileSet(),
		astPkgs: map[string]*ast.Package{},
		pkgs:    []*types.Package{},
		isDebug: isDebug,
	}
}

func (g generator) WriteTo(buf *bytes.Buffer) error {
	if err := g.ast(); err != nil {
		return err
	}
	if err := g.check(); err != nil {
		return err
	}
	g.parser.Build(g.pkgs)
	g.parser.WriteTo(buf)
	return nil
}

func (g *generator) Read(path string) error {
	if err := g.visit(path, nil, nil); err != nil {
		return err
	}
	return nil
}

func (g *generator) ReadDir(baseDir string) error {
	if err := filepath.Walk(baseDir, g.visit); err != nil {
		return err
	}

	return nil
}

func (g *generator) visit(path string, f os.FileInfo, err error) error {
	if ext := filepath.Ext(path); ext != ".go" {
		return nil
	}
	if strings.HasSuffix(path, "_test.go") {
		return nil
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return err
	}
	g.targets = append(g.targets, path)
	return nil
}

func (g generator) ast() error {
	for _, path := range g.targets {
		fmt.Printf("parsing AST: %s\n", path)
		astFile, err := parser.ParseFile(g.fset, path, nil, parser.ParseComments)
		if err != nil {
			return xerrors.Errorf("ParseFile panic: %w", err)
		}
		name := astFile.Name.Name
		pkg, ok := g.astPkgs[name]
		if !ok {
			pkg = &ast.Package{
				Name:  name,
				Files: make(map[string]*ast.File),
			}
		}
		pkg.Files[path] = astFile
		g.astPkgs[name] = pkg
	}
	return nil
}

func (g *generator) check() error {
	conf := types.Config{
		Importer: importer.For("source", nil),
		Error: func(err error) {
			if g.isDebug {
				fmt.Printf("error: %+v\n", err)
			}
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
