package gouml

import "github.com/kazukousen/gouml/internal/gouml/plantuml"

// PlantUMLParser ...
func PlantUMLParser() Parser {
	return plantuml.NewParser()
}
