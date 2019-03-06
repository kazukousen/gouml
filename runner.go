package gouml

import (
	"fmt"
	"os"
	"path"
	"strings"
)

// Runner ...
type Runner struct {
}

// NewRunner ...
func NewRunner() *Runner {
	return &Runner{}
}

// Run ...
func (r Runner) Run(baseDir string, out string) error {

	p := NewParser()
	if err := p.Parse(baseDir); err != nil {
		return err
	}

	g := NewGenerator()

	if err := r.print(p, g, path.Join(baseDir, out+".uml")); err != nil {
		return err
	}

	return nil
}

func (r Runner) print(p *Parser, g *Generator, fileName string) error {
	stmts := []string{}
	es := map[string]struct{}{}

	// print classes
	for pkg, types := range p.typeDefinitions {
		for typeName, typeSpec := range types {
			stmt, e := g.printClass(p, Vertex{Pkg: pkg, Name: typeName}, typeSpec.Type)
			stmts = append(stmts, stmt)
			es[e] = struct{}{}
		}
	}

	// print methods
	for pkg, types := range p.funcDefinitions {
		for typeName, funcDecls := range types {
			for _, funcDecl := range funcDecls {
				stmt, e := g.printMethod(p, Vertex{Pkg: pkg, Name: typeName}, funcDecl)
				stmts = append(stmts, fmt.Sprintf("%s : %s\n", NewHash(pkg, typeName), stmt))
				es[e] = struct{}{}
			}
		}
	}
	for e := range es {
		stmts = append(stmts, e)
	}

	uml, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer uml.Close()
	fmt.Fprintf(uml, "%s\n", strings.Join(stmts, "\n"))
	fmt.Println("ok")
	return err
}
