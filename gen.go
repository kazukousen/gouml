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
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// Generator ...
type Generator interface {
	UpdateIgnore(files []string) error
	Read(files []string) error
	WriteTo(buf *bytes.Buffer) error
}

type generator struct {
	logger      log.Logger
	parser      Parser
	targets     []string
	ignoreFiles map[string]struct{}
	fset        *token.FileSet
	astPkgs     map[string]*ast.Package
	pkgs        []*types.Package
	isDebug     bool
	mu          *sync.Mutex
}

// NewGenerator ...
func NewGenerator(logger log.Logger, parser Parser, isDebug bool) Generator {
	return &generator{
		logger:      log.With(logger, "component", "generator"),
		parser:      parser,
		targets:     []string{},
		ignoreFiles: map[string]struct{}{},
		fset:        token.NewFileSet(),
		astPkgs:     map[string]*ast.Package{},
		pkgs:        []*types.Package{},
		isDebug:     isDebug,
		mu:          &sync.Mutex{},
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

func (g *generator) Read(files []string) error {
	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		level.Debug(g.logger).Log("msg", "read .go files", "ms", elapsed.Milliseconds())
	}()

	for _, f := range files {
		if err := g.read(f); err != nil {
			return err
		}
	}
	return nil
}

func (g *generator) read(f string) error {
	fInfo, err := os.Stat(f)
	if err != nil {
		return err
	}

	if fInfo.IsDir() {
		if err := filepath.Walk(f, g.visit); err != nil {
			return err
		}
		return nil
	}

	if err := g.visit(f, nil, nil); err != nil {
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
	if _, ok := g.ignoreFiles[path]; ok {
		return nil
	}
	g.targets = append(g.targets, path)
	return nil
}

func (g *generator) UpdateIgnore(files []string) error {
	for _, f := range files {
		if err := g.updateIgnore(f); err != nil {
			return err
		}
	}
	return nil
}

func (g *generator) updateIgnore(f string) error {
	fInfo, err := os.Stat(f)
	if err != nil {
		return err
	}

	if fInfo.IsDir() {
		if err := filepath.Walk(f, g.doUpdateIgnore); err != nil {
			return err
		}
		return nil
	}

	if err := g.doUpdateIgnore(f, nil, nil); err != nil {
		return err
	}
	return nil
}

func (g *generator) doUpdateIgnore(path string, f os.FileInfo, err error) error {
	path, err = filepath.Abs(path)
	if err != nil {
		return err
	}
	g.ignoreFiles[path] = struct{}{}
	return nil
}

func (g generator) ast() error {
	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		level.Debug(g.logger).Log("msg", "parsed to AST", "ms", elapsed.Milliseconds())
	}()

	for _, path := range g.targets {
		if g.isDebug {
			fmt.Printf("parsing AST: %s\n", path)
		}
		astFile, err := parser.ParseFile(g.fset, path, nil, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("ParseFile panic: %w", err)
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
	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		level.Debug(g.logger).Log("msg", "checked type", "ms", elapsed.Milliseconds())
	}()

	conf := types.Config{
		Importer: importer.For("source", nil),
		Error: func(err error) {
			if g.isDebug {
				fmt.Printf("error: %+v\n", err)
			}
		},
	}
	wg := &sync.WaitGroup{}
	num := runtime.NumCPU()
	ch := make(chan *ast.Package, len(g.astPkgs))
	for i := 0; i < num; i++ {
		go g.workChecker(wg, conf, ch)
	}
	for _, astPkg := range g.astPkgs {
		wg.Add(1)
		ch <- astPkg
	}
	wg.Wait()
	close(ch)
	return nil
}

func (g *generator) appendPackage(pkg *types.Package) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.pkgs = append(g.pkgs, pkg)
}

func (g *generator) workChecker(wg *sync.WaitGroup, conf types.Config, ch <-chan *ast.Package) {
	for astPkg := range ch {
		files := []*ast.File{}
		for _, f := range astPkg.Files {
			files = append(files, f)
		}
		pkg, _ := conf.Check(astPkg.Name, g.fset, files, nil)
		g.appendPackage(pkg)

		wg.Done()
	}
}
