package gouml

import (
	"fmt"
	"go/types"
)

type modelKind string

const (
	modelKindInterface   modelKind = `interface "%s" as %s`
	modelKindValueObject modelKind = `class "%s" as %s <<V,Orchid>>`
	modelKindEntity      modelKind = `class "%s" as %s <<E,#FFCC00>>`
)

func (k modelKind) Printf(name, alias string) string {
	return fmt.Sprintf(string(k), name, alias)
}

func isEntity(f *types.Func) bool {
	sig, ok := f.Type().(*types.Signature)
	if !ok {
		return false
	}
	if _, ok := sig.Recv().Type().(*types.Pointer); !ok {
		return false
	}
	return sig.Results().Len() == 0
}
