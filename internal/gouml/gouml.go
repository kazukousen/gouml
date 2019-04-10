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

// Gen ...
func Gen(baseDir, out string) error {
	pkgs, err := read(baseDir)
	if err != nil {
		return err
	}
	objects := []types.Object{}
	exists := map[id]struct{}{}
	for _, pkg := range pkgs {
		scope := pkg.Scope()
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)
			objects = append(objects, obj)

			if obj.Pkg().Name() == pkg.Name() {
				exists[id{full: obj.Type().String()}] = struct{}{}
			}
		}
	}
	models := models{}
	cons := constantsMap{}
	for _, obj := range objects {
		switch obj := obj.(type) {

		// declared type
		case *types.TypeName:
			models.append(obj)

		// declared constant
		case *types.Const:
			if named, _ := obj.Type().(*types.Named); named != nil {
				id := id{full: named.String()}
				cons[id] = append(cons[id], obj)
			}
		}
	}

	buf := &bytes.Buffer{}
	models.WriteTo(buf, exists)

	cons.WriteTo(buf)

	newline(buf, 0)

	// fmt.Printf(buf.String())
	filename := path.Join(baseDir, out+".uml")
	uml, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer uml.Close()
	fmt.Fprintf(uml, buf.String())

	return nil
}

func read(baseDir string) ([]*types.Package, error) {
	fset := token.NewFileSet()
	astPkgs := map[string]*ast.Package{}
	files := []*ast.File{}
	filepath.Walk(baseDir, func(path string, f os.FileInfo, err error) error {
		if ext := filepath.Ext(path); ext == ".go" {
			src, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
			if err != nil {
				return xerrors.Errorf("ParseFile panic: %w", err)
			}
			name := src.Name.Name
			pkg, ok := astPkgs[name]
			if !ok {
				pkg = &ast.Package{
					Name:  name,
					Files: make(map[string]*ast.File),
				}
			}
			pkg.Files[path] = src
			astPkgs[name] = pkg
			files = append(files, src)
		}
		return nil
	})

	pkgs := []*types.Package{}

	conf := types.Config{
		Importer: importer.ForCompiler(fset, "source", nil),
		Error: func(err error) {
			// fmt.Printf("!!! %#v\n", err)
		},
	}
	for _, astPkg := range astPkgs {
		files := []*ast.File{}
		for _, f := range astPkg.Files {
			files = append(files, f)
		}
		pkg, _ := conf.Check(astPkg.Name, fset, files, nil)
		pkgs = append(pkgs, pkg)
	}
	return pkgs, nil
}
