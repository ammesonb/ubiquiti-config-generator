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
	if definition.Node != nil && definition.Node.IsTag {
		// For tags, include the name since that will describe the parent path
		// and e.g. firewall groups could have overlapping port-group and address-group names
		return strings.Join(append(definition.Path, definition.Name), "/")
	}

	return strings.Join(definition.Path, "/")
}

// FullPath returns the entire configuration path to this node, including itself
func (definition *Definition) FullPath() string {
	if len(definition.Path) == 0 {
		return definition.Name
	}

	var name any = definition.Name
	// For a tag, the path to use is actually the value
	if definition.Node != nil && definition.Node.IsTag {
		name = definition.Value
	}

	return fmt.Sprintf("%s/%v", definition.JoinedPath(), name)
}

// Diff returns where another definition differs from this one
func (definition *Definition) Diff(other *Definition) []string {
	differences := definition.diffDefinition(other)

	if definition.Node != nil && other.Node != nil {
		differences = append(differences, definition.diffNode(other)...)
	}

	if len(definition.Children) != len(other.Children) {
		return append(
			differences,
			fmt.Sprintf(
				"%s: Should have %d children but got %d",
				definition.FullPath(),
				len(definition.Children),
				len(other.Children),
			),
		)
	}

	for idx, child := range definition.Children {
		otherChild := other.Children[idx]
		differences = append(differences, child.Diff(otherChild)...)
	}

	return differences
}

func (definition *Definition) diffDefinition(other *Definition) []string {
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
	if definition.Node == nil {
		differences = append(
			differences,
			fmt.Sprintf(
				"%s: Definition node should not be nil",
				definition.FullPath(),
			),
		)
	}
	if other.Node == nil {
		differences = append(
			differences,
			fmt.Sprintf(
				"%s: Other node should not be nil",
				definition.FullPath(),
			),
		)
	}

	return differences
}

func (definition *Definition) diffNode(other *Definition) []string {
	differences := []string{}

	// Skip node child check since that could be expensive
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

	return differences
}

func (definition *Definition) hasChild(other *Definition) bool {
	for _, child := range definition.Children {
		if len(child.diffDefinition(other)) == 0 {
			return true
		}
	}

	return false
}

// Definitions collects all user-defined node values, including methods of indexing them
// Used to merge in generic vyos configurations with custom YAML-based abstractions
type Definitions struct {
	Definitions []*Definition
	// While some nodes will have multiple values, since they are tagged "multi" then
	// the values will be arrays so only one definition per path still
	NodeByPath       map[string]*Node
	DefinitionByPath map[string]*Definition
}

// FindChild recursively delves into this node's children along the given path
func (definitions *Definitions) FindChild(path []any) *Definition {
	var children []*Definition

	top, ok := definitions.DefinitionByPath[path[0].(string)]
	children = append(children, top)
	if !ok {
		logger.DefaultLogger().Debugf(
			"Could not find node for step '%s' of path '%v'",
			path[0],
			path,
		)
		return nil
	}

	foundTag := false
	for idx, step := range path[1:] {
		found := false
		// Tags will skip a step in the path, so reset flag and continue
		if foundTag {
			foundTag = false
			continue

		}

		for _, child := range children[len(children)-1].Children {
			// Child matches if the name matches the step and either
			// - the node is not a tag
			// - the node IS a tag, and its value matches the next step of the path (+2 since starting from index 1 on path)
			if child.Name == step && (!child.Node.IsTag || (child.Node.IsTag && child.Value == path[idx+2])) {
				foundTag = child.Node.IsTag

				children = append(children, child)
				found = true
				break
			}
		}

		if !found {
			logger.DefaultLogger().Debugf(
				"Could not find node for step '%s' of path '%v'",
				step,
				path,
			)
			return nil
		}
	}

	return children[len(children)-1]
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
		definitions.Definitions = append(definitions.Definitions, definition)
	} else {
		logger.Debugf(
			"Adding relationship for definition '%s' with parent '%s'",
			definition.FullPath(),
			definition.ParentPath(),
		)
		// Add this definition to its parent
		parent := definitions.DefinitionByPath[definition.ParentPath()]
		if !parent.hasChild(definition) {
			parent.Children = append(
				parent.Children, definition,
			)
		}
	}

	for _, child := range definition.Children {
		definitions.add(child)
	}
}

func (definitions *Definitions) merge(other *Definitions) error {
	// Check top-level definitions
	for _, definition := range other.Definitions {
		// If not present already, then add it
		if _, ok := definitions.DefinitionByPath[definition.FullPath()]; !ok {
			definitions.add(definition)
		} else {
			// Otherwise, recurse into the definition to merge them
			if err := definitions.DefinitionByPath[definition.FullPath()].merge(definitions, definition); err != nil {
				return err
			}
		}
	}

	return nil
}

func (definition *Definition) merge(definitions *Definitions, other *Definition) error {
	diffs := definition.diffDefinition(other)
	// Make sure the attributes on the definition are the same, otherwise they are not compatible
	if len(diffs) > 0 {
		return fmt.Errorf(
			"got %d differences between nodes at path %s: %s",
			len(diffs),
			definition.FullPath(),
			strings.Join(diffs, "\n"),
		)
	}

	for _, child := range other.Children {
		// For each child, if not present then we can simply add it to this definition
		// Otherwise, try merging it in recursively
		if _, ok := definitions.DefinitionByPath[child.FullPath()]; !ok {
			definitions.add(child)
		} else {
			if err := definitions.DefinitionByPath[child.FullPath()].merge(definitions, child); err != nil {
				return err
			}
		}
	}
	return nil
}

func initDefinitions() *Definitions {
	return &Definitions{
		Definitions:      make([]*Definition, 0),
		NodeByPath:       make(map[string]*Node),
		DefinitionByPath: make(map[string]*Definition),
	}
}

func generateSparseDefinitionTree(nodes *Node, path []string) *Definition {
	steps := make([]string, 0)
	definition := &Definition{
		Name:     path[0],
		Path:     steps,
		Node:     nodes.FindChild([]string{path[0]}),
		Children: []*Definition{},
	}

	steps = append(steps, path[0])

	def := definition
	for _, step := range path[1:] {
		def.Children = []*Definition{
			{
				Name:     step,
				Path:     steps,
				Node:     nodes.FindChild(append(steps, step)),
				Children: []*Definition{},
			},
		}

		steps = append(steps, step)

		// Reassign the pointer to the newly-created child so we recurse
		def = def.Children[0]
	}

	return definition
}

type BasicDefinition struct {
	Name     string
	Comment  string
	Value    any
	Values   []any
	Children []BasicDefinition
}

func generatePopulatedDefinitionTree(nodes *Node, definition BasicDefinition, path []string, nodePath []string) *Definition {
	def := &Definition{
		Name:     definition.Name,
		Path:     path,
		Node:     nodes.FindChild(append(nodePath, definition.Name)),
		Comment:  definition.Comment,
		Value:    definition.Values,
		Values:   definition.Values,
		Children: make([]*Definition, 0),
	}

	for _, child := range definition.Children {
		pathSuffix := definition.Name
		if pathSuffix == "node.tag" {
			pathSuffix = definition.Value.(string)
		}
		// Recurse into each child and add the generated definition
		def.Children = append(
			def.Children,
			generatePopulatedDefinitionTree(
				nodes,
				child,
				append(path, pathSuffix),
				append(nodePath, definition.Name),
			),
		)
	}

	return def
}
