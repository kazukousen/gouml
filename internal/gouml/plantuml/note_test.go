package plantuml_test

import (
	"bytes"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/kazukousen/gouml/internal/gouml/plantuml"
)

func TestNote(t *testing.T) {
	notes := plantuml.Notes{}
	fset := token.NewFileSet()
	src := `
	package time
	type Weekday int

	const (
		Sunday Weekday = iota
		Monday
		Tuesday
		Wednesday
		Thursday
		Friday
		Saturday
	)
	`
	want := `

	package "time" {
		note as N_time_Weekday
		<b>Weekday</b>

		Friday
		Monday
		Saturday
		Sunday
		Thursday
		Tuesday
		Wednesday
	end note
	}
	N_time_Weekday --> time.Weekday`
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		t.Errorf(": %+v", err)
		return
	}
	conf := types.Config{
		Importer: importer.Default(),
	}
	pkg, err := conf.Check(file.Name.Name, fset, []*ast.File{file}, nil)
	if err != nil {
		t.Errorf(": %+v", err)
		return
	}
	for _, name := range pkg.Scope().Names() {
		obj := pkg.Scope().Lookup(name)
		if c, _ := obj.(*types.Const); c != nil {
			if named, _ := obj.Type().(*types.Named); named != nil {
				plantuml.ExportTestNotesAppend(notes, named, c)
			}
		}
	}
	buf := &bytes.Buffer{}
	notes.WriteTo(buf)
	if g, w := trim(buf.String()), trim(want); g != w {
		t.Errorf("not equal\ngot: %s\nwant: %s", g, w)
	}
}
