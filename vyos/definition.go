package vyos

import (
	"fmt"
	"github.com/ammesonb/ubiquiti-config-generator/utils"
	"reflect"
	"strings"

	"github.com/ammesonb/ubiquiti-config-generator/config"
	"github.com/ammesonb/ubiquiti-config-generator/console_logger"
)

// Definition contains the actual values for a given node
// Also requires a path to ensure we know the actual value for tag nodes
type Definition struct {
	Name       string        `yaml:"name"`
	Path       []string      `yaml:"path"`
	Node       *Node         `yaml:"node"`
	Comment    string        `yaml:"comment"`
	Value      any           `yaml:"value"`
	Values     []any         `yaml:"values"`
	Children   []*Definition `yaml:"children"`
	ParentNode *Node
}

var errMergeConflict = "got %d differences between nodes at path %s: %s"

// ParentPath returns a slash-joined version of the path to the definition's parent
// Dynamic nodes are an essential part of the configuration path, so do not need to skip those
func (definition *Definition) ParentPath() string {
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
		differences = append(differences, definition.Node.diffNode(other.Node)...)
	}

	if definition.ParentNode != nil && other.ParentNode != nil {
		differences = append(differences, definition.ParentNode.diffNode(other.ParentNode)...)
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

	// Only check this if length of children matches
	for _, child := range definition.Children {
		// Slightly inefficient n^2 comparison of children here, but quick fail cases and
		// generally configs are small enough, so not worth time to re-index the entire children array
		for _, otherChild := range other.Children {
			// If name matches, and also node tag value matches if applicable
			if child.Name == otherChild.Name && (!child.Node.IsTag || child.Value == otherChild.Value) {
				differences = append(differences, child.Diff(otherChild)...)
			}
		}
	}

	return differences
}

func (definition *Definition) diffDefinition(other *Definition) []string {
	var differences []string
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
	if len(definition.Path) > 0 && definition.ParentNode == nil {
		differences = append(
			differences,
			fmt.Sprintf(
				"%s: Definition parent node should not be nil",
				definition.FullPath(),
			),
		)
	}
	if len(other.Path) > 0 && other.ParentNode == nil {
		differences = append(
			differences,
			fmt.Sprintf(
				"%s: Other parent node should not be nil",
				definition.FullPath(),
			),
		)
	}

	return differences
}

func (definition *Definition) hasChild(other *Definition) bool {
	for _, child := range definition.Children {
		if (!child.Node.IsTag && child.Name == other.Name) ||
			(child.Node.IsTag && other.Node.IsTag && child.Value == other.Value) {
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
		console_logger.DefaultLogger().Debugf(
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

		for _, child := range utils.Last(children).Children {
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
			console_logger.DefaultLogger().Debugf(
				"Could not find node for step '%s' of path '%v'",
				step,
				path,
			)
			return nil
		}
	}

	return utils.Last(children)
}

func (definitions *Definitions) add(definition *Definition) {
	logger := console_logger.DefaultLogger()

	// For dynamic nodes, there are actually two paths - one for the root tagged element, and one for the instance of the tagged element
	// Need to add both, since if we want to check if a certain path exists, we may need the root of the tag as well
	nodeParent := definition.Node.ParentPath()
	if _, ok := definitions.NodeByPath[nodeParent]; !ok {
		definitions.NodeByPath[nodeParent] = definition.ParentNode
	}

	fullPath := definition.FullPath()

	// Check if we have defined the definition already
	// This will occasionally happen if there are overlapping paths between dynamic and static configs, but usually
	// is okay as long as you are careful
	if _, ok := definitions.DefinitionByPath[fullPath]; ok {
		logger.Warnf(
			"Already has definition '%s' defined for path '%s'",
			definitions.DefinitionByPath[fullPath].FullPath(),
			fullPath,
		)
	} else if len(definition.Path) == 0 {
		// Only add root definitions to the list since others will be nested as children,
		// and make sure we only do it once
		logger.Debugf("Adding definition for '%s' to root list", definition.Name)
		definitions.Definitions = append(definitions.Definitions, definition)
		definitions.NodeByPath[fullPath] = definition.Node
		definitions.DefinitionByPath[fullPath] = definition
	} else {
		// For any other new node, add indexing to this definition
		// If it already exists, then this new index would clobber it
		logger.Debugf("Adding indexing for path %s", fullPath)
		definitions.NodeByPath[fullPath] = definition.Node
		definitions.DefinitionByPath[fullPath] = definition
	}

	// If this definition has a parent, check to see if the parent already has this node as a child
	if len(definition.Path) > 0 {
		parentPath := definition.ParentPath()
		logger.Debugf(
			"Adding relationship for definition '%s' with parent '%s'",
			fullPath,
			parentPath,
		)
		// Add this definition to its parent
		parent := definitions.DefinitionByPath[parentPath]
		if len(definition.Path) > 0 && !parent.hasChild(definition) {
			parent.Children = append(
				parent.Children, definition,
			)
		}
	}

	// Recurse into any children this definition may have
	for _, child := range definition.Children {
		definitions.add(child)
	}
}

var errDiffLength = "cannot ensure paths for differing definition and node lengths: got %d definitions and %d nodes"
var errUnmatchedDynamicNode = "cannot end tree path on dynamic entry without defined tag: %s"

func (definitions *Definitions) ensureTree(nodes *Node, path *utils.VyosPath) error {
	if len(path.Path) != len(path.NodePath) {
		return utils.ErrWithVarCtx(errDiffLength, len(path.Path), len(path.NodePath))
	}

	lastDynamic := false
	defPath := make([]string, 0)
	nodePath := make([]string, 0)
	var parentNode *Node = nil
	// Loop through each step of the path and add missing definitions
	// Working in reverse (bottom-up) would allow detection of when a tree starts existing,
	// but then need to either build definitions backwards or reverse direction again so top-down is simpler
	for idx := range path.Path {
		// Dynamic nodes will consume the next as its value for a tag, so skip this index but do append the
		// tagged values, for proper definition traversals
		// Also clear the flag, to avoid infinitely skipping later paths
		if lastDynamic {
			defPath = append(defPath, path.Path[idx])
			nodePath = append(nodePath, path.NodePath[idx])
			lastDynamic = false
			continue
		}

		// Track this definition's path
		childPath := utils.CopySliceWith(defPath, path.Path[idx])
		fullPath := strings.Join(childPath, "/")
		childNodePath := utils.CopySliceWith(nodePath, path.NodePath[idx])
		fullNodePath := strings.Join(childNodePath, "/")
		// Identify this node, and update the last dynamic state based on its configuration
		node := nodes.FindChild(childNodePath)
		if node == nil {
			return utils.ErrWithCtx(errNonexistentNode, fullNodePath)
		} else if node.IsTag {
			lastDynamic = true
			if len(path.Path) == idx {
				return utils.ErrWithCtx(errUnmatchedDynamicNode, fullPath)
			}
			// Add the next path definition name to the string path, so when we check if the given definition is already
			// included, it is fully qualified instead of the placeholder tag instead
			// e.g. firewall -> name is meaningless without the trailing -> <some firewall name>
			fullPath = fullPath + "/" + path.Path[idx+1]
		}
		// Check if this path is defined or not
		// If this path is already defined, then we don't need to add anything here
		// Child paths will be added to parent definitions by the `add` function, so can safely append new children
		if _, ok := definitions.DefinitionByPath[fullPath]; !ok {
			var value any
			// For tags, we require the next definition to be present, so we know the value to use
			if node.IsTag {
				value = path.Path[idx+1]
			}
			// Since definition not present, we can safely add it
			definitions.add(&Definition{
				Name:       path.Path[idx],
				Path:       defPath,
				Node:       node,
				Value:      value,
				Children:   make([]*Definition, 0),
				ParentNode: parentNode,
			})
		}

		// Assign paths to the children versions, to recurse further
		defPath = childPath
		nodePath = childNodePath
		// Only reassign parent node on meaningful nodes, since no definitions will be added
		// on the placeholder tag nodes
		parentNode = node
	}

	return nil
}

func (definitions *Definitions) addValue(nodes *Node, path *utils.VyosPath, keyName string, value any) {
	parent := nodes.FindChild(path.NodePath)
	if parent.Name == utils.DYNAMIC_NODE {
		parent = nodes.FindChild(utils.AllExcept(path.NodePath, 1))
	}
	definitions.add(&Definition{
		Name:       keyName,
		Node:       nodes.FindChild(append(path.NodePath, keyName)),
		Path:       path.Path,
		Value:      value,
		ParentNode: parent,
	})
}

func (definitions *Definitions) addListValue(nodes *Node, path *utils.VyosPath, keyName string, value []any) {
	parent := nodes.FindChild(path.NodePath)
	if parent.Name == utils.DYNAMIC_NODE {
		parent = nodes.FindChild(utils.AllExcept(path.NodePath, 1))
	}
	definitions.add(&Definition{
		Name:       keyName,
		Node:       nodes.FindChild(append(path.NodePath, keyName)),
		Path:       path.Path,
		Values:     value,
		ParentNode: parent,
	})
}

func (definitions *Definitions) appendToListValue(nodes *Node, path *utils.VyosPath, keyName string, value any) {
	node := definitions.FindChild(config.SliceStrToAny(utils.CopySliceWith(path.Path, keyName)))
	if node == nil {
		definitions.addListValue(nodes, path, keyName, []any{value})
	} else {
		node.Values = append(node.Values, value)
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
		return utils.ErrWithVarCtx(errMergeConflict, len(diffs), definition.FullPath(), strings.Join(diffs, "\n"))
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

func generateSparseDefinitionTree(nodes *Node, path *utils.VyosPath) *Definition {
	steps := make([]string, 0)
	definition := &Definition{
		Name:     path.Path[0],
		Path:     steps,
		Node:     nodes.FindChild([]string{path.NodePath[0]}),
		Children: []*Definition{},
	}

	nodeSteps := utils.CopySliceWith(steps, path.Path[0])
	steps = append(steps, path.Path[0])

	def := definition
	for idx, step := range path.Path[1:] {
		// First check parent node
		parent := nodes.FindChild(nodeSteps)
		// If parent is a dynamic node, then the tag's name/value was pulled from the next path entry
		// However, it still needs to be added to both paths
		// Since we already created a definition referencing it, do nothing else and continue iterating
		if parent.IsTag {
			steps = append(steps, step)
			nodeSteps = append(nodeSteps, utils.DYNAMIC_NODE)
			continue
		} else if parent.Name == utils.DYNAMIC_NODE {
			// If parent is a placeholder node, then traverse up one further to find the real parent,
			// since this one contains no real information
			parent = nodes.FindChild(utils.AllExcept(nodeSteps, 1))
		}

		node := nodes.FindChild(append(nodeSteps, step))

		name := step
		var value any = nil
		// Tag nodes will need to get an extra step appended to the node path,
		// but the definition path will skip over the placeholder
		if node.IsTag {
			name = step
			// Offset by 1 since starting at index 1, then one more to look ahead
			value = path.Path[idx+2]
		}
		def.Children = []*Definition{
			{
				Name:       name,
				Path:       steps,
				Node:       node,
				Value:      value,
				Children:   []*Definition{},
				ParentNode: parent,
			},
		}

		steps = append(steps, step)
		nodeSteps = append(nodeSteps, step)

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

func generatePopulatedDefinitionTree(nodes *Node, definition BasicDefinition, path *utils.VyosPath, parentNode *Node) *Definition {
	node := nodes.FindChild(append(path.NodePath, definition.Name))

	def := &Definition{
		Name:       definition.Name,
		Path:       path.Path,
		Node:       node,
		Comment:    definition.Comment,
		Value:      definition.Value,
		Values:     definition.Values,
		Children:   make([]*Definition, 0),
		ParentNode: parentNode,
	}

	for _, child := range definition.Children {
		var childPath *utils.VyosPath
		// If node is flagged as a tag, or the name of the node is the tag placeholder,
		// then we need to use the definition's value instead
		if node.IsTag || node.Name == utils.DYNAMIC_NODE {
			childPath = path.Extend(
				utils.MakeVyosPC(definition.Name),
				utils.MakeVyosDynamicPC(definition.Value.(string)),
			)
		} else {
			childPath = path.Extend(utils.MakeVyosPC(definition.Name))
		}

		// Recurse into each child and add the generated definition
		def.Children = append(
			def.Children,
			generatePopulatedDefinitionTree(
				nodes,
				child,
				childPath,
				node,
			),
		)
	}

	return def
}
