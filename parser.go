package gouml

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
)

// Parser ...
type Parser struct {
	files           map[string]*ast.File
	typeDefinitions map[string]map[string]*ast.TypeSpec
	variableNodes   map[string]map[string][]*ast.ValueSpec
	funcDefinitions map[string]map[string][]*ast.FuncDecl
}

// NewParser ...
func NewParser() *Parser {
	return &Parser{
		files:           map[string]*ast.File{},
		typeDefinitions: map[string]map[string]*ast.TypeSpec{},
		variableNodes:   map[string]map[string][]*ast.ValueSpec{},
		funcDefinitions: map[string]map[string][]*ast.FuncDecl{},
	}
}

// Parse ...
func (p *Parser) Parse(baseDir string) error {
	if err := p.readGoFiles(baseDir); err != nil {
		return err
	}

	p.storeNodes()

	return nil
}

// readGoFiles stores AST convert from {.go} file.
func (p Parser) readGoFiles(baseDir string) error {
	return filepath.Walk(baseDir, p.visit)
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

// storeNodes stores nodes of type declaretion and func declaretion
func (p Parser) storeNodes() {
	for _, ast := range p.files {
		p.storeGenDecl(ast)
		p.storeFuncDecl(ast)
	}
}

func (p *Parser) storeGenDecl(astFile *ast.File) {
	pkg := astFile.Name.Name
	for _, decl := range astFile.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range genDecl.Specs {
				switch spec := spec.(type) {
				case *ast.TypeSpec:
					if _, ok := p.typeDefinitions[pkg]; !ok {
						p.typeDefinitions[pkg] = map[string]*ast.TypeSpec{}
					}
					p.typeDefinitions[pkg][spec.Name.Name] = spec
				case *ast.ValueSpec:
					if _, ok := p.variableNodes[pkg]; !ok {
						p.variableNodes[pkg] = map[string][]*ast.ValueSpec{}
					}
					if doc := genDecl.Doc; doc != nil {
						text := doc.Text()
						p.variableNodes[pkg][text] = append(p.variableNodes[pkg][text], spec)
					}
				}
			}
		}
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
