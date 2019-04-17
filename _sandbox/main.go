package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
)

func run() error {
	code := `
	package main

	type Human struct {
		Name string
		Age  Age
	}

	type Age int

	func (a Age) IsAdult() bool {
		return a >= 20
	}

	func main() {
	}
	`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "main.go", code, parser.ParseComments)
	if err != nil {
		return err
	}

	conf := types.Config{
		Importer: importer.Default(),
		Error: func(err error) {
			fmt.Printf("error: %+v\n", err)
		},
	}

	pkg, err := conf.Check(file.Name.Name, fset, []*ast.File{file}, nil)
	if err != nil {
		return err
	}

	for _, name := range pkg.Scope().Names() {
		obj := pkg.Scope().Lookup(name)
		switch obj := obj.(type) {

		// declared type
		case *types.TypeName:
			typ := obj.Type()
			fmt.Printf("name=%s typ=%s\n", obj.Name(), typ.String())
			named := typ.(*types.Named)
			for i := 0; i < named.NumMethods(); i++ {
				fn := named.Method(i)
				fmt.Println(fn.String())
			}

			switch un := typ.Underlying().(type) {
			// type Foo struct{}
			case *types.Struct:
				for i := 0; i < un.NumFields(); i++ {
					v := un.Field(i)
					fmt.Println(v.String())
				}
			}
		}
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
	}
}
