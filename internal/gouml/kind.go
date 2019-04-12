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
	if sig.Results().Len() == 0 {
		return true
	}
	if sig.Results().Len() == 1 {
		t := sig.Results().At(0).Type()
		errType := types.Universe.Lookup("error").Type()
		if types.Implements(t, errType.Underlying().(*types.Interface)) {
			return true
		}
	}
	return false
}
