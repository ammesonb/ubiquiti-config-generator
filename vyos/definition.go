package vyos

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
)

// Definition contains the actual values for a given node
// Also requires a path to ensure we know the actual value for tag nodes
type Definition struct {
	Name     string
	Path     []string
	Node     *Node
	Comment  string
	Value    any
	Values   []any
	Children []Definition
}

// FullPath returns the entire configuration path to this node, including itself
func (definition *Definition) FullPath() string {
	return fmt.Sprintf("%s/%s", strings.Join(definition.Path, "/"), definition.Name)
}

// Definitions collects all user-defined node values, including methods of indexing them
// Used to merge in generic vyos configurations with custom YAML-based abstractions
type Definitions struct {
	Definitions []Definition
	// While some nodes will have multiple values, since they are tagged "multi" then
	// the values will be arrays so only one definition per path still
	NodeByPath       map[string]*Node
	DefinitionByPath map[string]*Definition
}

func (definitions *Definitions) add(definition *Definition) {
	if _, ok := definitions.NodeByPath[definition.FullPath()]; ok {
		log.Warnf(
			"Already has node '%s' (%s) defined for path '%s'",
			definitions.NodeByPath[definition.FullPath()].Name,
			definitions.NodeByPath[definition.FullPath()].Help,
			definition.FullPath(),
		)
	}
	if _, ok := definitions.DefinitionByPath[definition.FullPath()]; ok {
		log.Warnf(
			"Already has definition '%s' defined for path '%s'",
			definitions.DefinitionByPath[definition.FullPath()].FullPath(),
			definition.FullPath(),
		)
	}

	definitions.NodeByPath[definition.FullPath()] = definition.Node
	definitions.DefinitionByPath[definition.FullPath()] = definition

	// Only add root definitions to the list since others will be nested as children
	if len(definition.Path) == 0 {
		definitions.Definitions = append(definitions.Definitions, *definition)
	}
}
