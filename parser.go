package gouml

import (
	"bytes"
	"go/types"
)

// Parser ...
type Parser interface {
	Build(pkgs []*types.Package)
	WriteTo(buf *bytes.Buffer)
}
