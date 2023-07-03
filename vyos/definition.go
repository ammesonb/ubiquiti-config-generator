package vyos

import (
	"fmt"
	"strings"

	"github.com/ammesonb/ubiquiti-config-generator/logger"
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
	Children []*Definition
}

// ParentPath returns a slash-joined version of the path to the definition's parent
func (definition *Definition) ParentPath() string {
	if definition.Node.IsTag {
		// For tags, include the name since that will describe the parent path
		// and e.g. firewall groups could have overlapping port-group and address-group names
		return strings.Join(definition.Path, "/")
	}

	return strings.Join(definition.Path, "/")
}

// JoinedPath returns a slash-joined version of the definition's path
func (definition *Definition) JoinedPath() string {
	if definition.Node.IsTag {
		// For tags, include the name since that will describe the parent path
		// and e.g. firewall groups could have overlapping port-group and address-group names
		return strings.Join(append(definition.Path, definition.Name), "/")
	}

	return strings.Join(definition.Path, "/")
}

// FullPath returns the entire configuration path to this node, including itself
func (definition *Definition) FullPath() string {
	name := definition.Name
	// For a tag, the path to use is actually the value
	if definition.Node.IsTag {
		name = definition.Value.(string)
	}
	if len(definition.Path) > 0 {
		return fmt.Sprintf("%s/%s", definition.JoinedPath(), name)
	}

	return name
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
	logger := logger.DefaultLogger()
	if _, ok := definitions.NodeByPath[definition.FullPath()]; ok {
		logger.Warnf(
			"Already has node '%s' (%s) defined for path '%s'",
			definitions.NodeByPath[definition.FullPath()].Name,
			definitions.NodeByPath[definition.FullPath()].Help,
			definition.FullPath(),
		)
	}
	if _, ok := definitions.DefinitionByPath[definition.FullPath()]; ok {
		logger.Warnf(
			"Already has definition '%s' defined for path '%s'",
			definitions.DefinitionByPath[definition.FullPath()].FullPath(),
			definition.FullPath(),
		)
	}

	logger.Debugf("Adding indexing for path %s", definition.FullPath())
	definitions.NodeByPath[definition.FullPath()] = definition.Node
	definitions.DefinitionByPath[definition.FullPath()] = definition

	// Only add root definitions to the list since others will be nested as children
	if len(definition.Path) == 0 {
		logger.Debugf("Adding definition for '%s' to root list", definition.Name)
		definitions.Definitions = append(definitions.Definitions, *definition)
	} else {
		logger.Debugf(
			"Adding relationship for definition '%s' with parent '%s'",
			definition.FullPath(),
			definition.ParentPath(),
		)
		// Add this definition to its parent
		definitions.DefinitionByPath[definition.ParentPath()].Children = append(
			definitions.DefinitionByPath[definition.ParentPath()].Children, definition,
		)
	}
}

func initDefinitions() *Definitions {
	return &Definitions{
		Definitions:      make([]Definition, 0),
		NodeByPath:       make(map[string]*Node),
		DefinitionByPath: make(map[string]*Definition),
	}
}
