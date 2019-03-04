package gouml

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// Token ...
type Token struct {
	pkg        string
	name       string
	hash       string
	objKind    objKind
	fieldNames TokenFieldNames
}

// TokenFieldNames ...
type TokenFieldNames map[string]string

// NewTokenFieldNames ...
func NewTokenFieldNames() TokenFieldNames {
	return TokenFieldNames{}
}

// Add ...
func (fs TokenFieldNames) Add(name string, typeName string) {
	fs[name] = typeName
}

func (fs TokenFieldNames) String() string {
	var dst string
	for k, v := range fs {
		dst += fmt.Sprintf("\t\t%s %s\n", k, v)
	}
	return dst
}

// NewToken ...
func NewToken(pkg string, name string, objKind objKind, fieldNames TokenFieldNames) *Token {
	return &Token{
		pkg:        pkg,
		name:       name,
		hash:       NewHash(pkg, name),
		objKind:    objKind,
		fieldNames: fieldNames,
	}
}

// NewHash ...
func NewHash(pkg string, name string) string {
	hasher := sha256.New()
	hasher.Write(bytes.Join([][]byte{[]byte(pkg), []byte(name)}, []byte{}))
	return hex.EncodeToString(hasher.Sum(nil))[:6]
}

func (t Token) String() string {
	return fmt.Sprintf(`
package %s {
	class "%s" as %s <<%s>> {
%s
	}
}
	`, t.pkg, t.name, t.hash, t.objKind, t.fieldNames.String())
}

type objKind string

const (
	objKindValueObject objKind = "V,orchid"
	objKindEntity      objKind = "E,orchid"
)

func (k objKind) String() string {
	return string(k)
}

// Edges ...
type Edges []Edge

func (es Edges) String() string {
	dst := make([]string, 0, len(es))
	for _, e := range es {
		dst = append(dst, e.String())
	}
	return strings.Join(dst, "\n")
}

// Edge ...
type Edge struct {
	From    Vertex
	To      Vertex
	IsArray bool
}

func (e Edge) String() string {
	var rel string
	if e.IsArray {
		rel = "*--"
	} else {
		rel = "-->"
	}
	return e.From.hash() + rel + e.To.hash()
}

// Vertex ...
type Vertex struct {
	Pkg  string
	Name string
}

func (v Vertex) hash() string {
	return NewHash(v.Pkg, v.Name)
}
