package vyos

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/ammesonb/ubiquiti-config-generator/logger"
)

// Definition contains the actual values for a given node
// Also requires a path to ensure we know the actual value for tag nodes
type Definition struct {
	Name     string        `yaml:"name"`
	Path     []string      `yaml:"path"`
	Node     *Node         `yaml:"node"`
	Comment  string        `yaml:"comment"`
	Value    any           `yaml:"value"`
	Values   []any         `yaml:"values"`
	Children []*Definition `yaml:"children"`
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

// Diff returns where another definition differs from this one
func (definition *Definition) Diff(other *Definition) []string {
	differences := []string{}
	if definition.Name != other.Name {
		differences = append(
			differences,
			fmt.Sprintf(
				"%s: 'Name' should be '%s' but got '%s'",
				definition.FullPath(),
				definition.Name,
				other.Name,
			),
		)
	}
	if !reflect.DeepEqual(definition.Path, other.Path) {
		differences = append(
			differences,
			fmt.Sprintf(
				"%s: 'Path' should be '%#v' but got '%#v'",
				definition.FullPath(),
				definition.Path,
				other.Path,
			),
		)
	}
	if definition.Comment != other.Comment {
		differences = append(
			differences,
			fmt.Sprintf(
				"%s: 'Comment' should be '%s' but got '%s'",
				definition.FullPath(),
				definition.Comment,
				other.Comment,
			),
		)
	}
	if definition.Value != other.Value {
		differences = append(
			differences,
			fmt.Sprintf(
				"%s: 'Value' should be '%#v' but got '%#v'",
				definition.FullPath(),
				definition.Value,
				other.Value,
			),
		)
	}
	if !reflect.DeepEqual(definition.Values, other.Values) {
		differences = append(
			differences,
			fmt.Sprintf(
				"%s: 'Values' should be '%#v' but got '%#v'",
				definition.FullPath(),
				definition.Values,
				other.Values,
			),
		)
	}
	if len(definition.Children) != len(other.Children) {
		differences = append(
			differences,
			fmt.Sprintf(
				"%s: Count of children be %d but got %d",
				definition.FullPath(),
				len(definition.Children),
				len(other.Children),
			),
		)
	}
	// Skip node child check since that could be expensive {}
	if definition.Node.Name != other.Node.Name {
		differences = append(
			differences,
			fmt.Sprintf(
				"%s: Node 'Name' should be '%s' but got '%s'",
				definition.FullPath(),
				definition.Node.Name,
				other.Node.Name,
			),
		)
	}
	if definition.Node.Type != other.Node.Type {
		differences = append(
			differences,
			fmt.Sprintf(
				"%s: Node 'Type' should be '%s' but got '%s'",
				definition.FullPath(),
				definition.Node.Type,
				other.Node.Type,
			),
		)
	}
	if definition.Node.IsTag != other.Node.IsTag {
		differences = append(
			differences,
			fmt.Sprintf(
				"%s: Node 'IsTag' should be '%t' but got '%t'",
				definition.FullPath(),
				definition.Node.IsTag,
				other.Node.IsTag,
			),
		)
	}
	if definition.Node.Multi != other.Node.Multi {
		differences = append(
			differences,
			fmt.Sprintf(
				"%s: Node 'Multi' should be '%t' but got '%t'",
				definition.FullPath(),
				definition.Node.Multi,
				other.Node.Multi,
			),
		)
	}
	if definition.Node.Path != other.Node.Path {
		differences = append(
			differences,
			fmt.Sprintf(
				"%s: Node 'Path' should be '%s' but got '%s'",
				definition.FullPath(),
				definition.Node.Path,
				other.Node.Path,
			),
		)
	}

	for idx, child := range definition.Children {
		otherChild := other.Children[idx]
		differences = append(differences, child.Diff(otherChild)...)
	}

	return differences
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
