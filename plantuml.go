package gouml

import (
	"github.com/go-kit/kit/log"
	"github.com/kazukousen/gouml/internal/gouml/plantuml"
)

// PlantUMLParser ...
func PlantUMLParser(logger log.Logger) Parser {
	return plantuml.NewParser(logger)
}
