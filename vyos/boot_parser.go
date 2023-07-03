package vyos

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/ammesonb/ubiquiti-config-generator/logger"
	"github.com/charmbracelet/log"
)

/*
	- allow for modular configs - config.boot always checked, plus
		either interfaces.boot or maybe interfaces/<others>.boot?
	- With recursion support?
	- Lines for new paths end with {
	- Scopes always close with whitespace then }
	- Values are always on one line it seems
	- Comments start with /*
	- Tagged nodes will have the format "name <name>"
*/

// ParseBootDefinitions takes an opened reader of a config boot file and definitions
// and will update the definitions with the data from the file
func ParseBootDefinitions(reader io.Reader, definitions *Definitions, rootNode *Node) {
	logger := logger.DefaultLogger()
	logger.SetLevel(log.DebugLevel)
	definitionStack := make([]*Definition, 0)
	nodeStack := []*Node{rootNode}

	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		} else if lineIsComment(line) {
			definitionStack[len(definitionStack)-1].Comment = line
		} else if lineCreatesNode(line) {
			openScope(line, rootNode, definitions, &definitionStack, &nodeStack, logger)
		} else if lineEndsNode(line) {
			closeScope(&definitionStack, &nodeStack, logger)
		} else if lineHasValue(line) {
			setValue(line, definitions, definitionStack, nodeStack, logger)
		} else {
			handleUnknown(line, definitions, definitionStack, nodeStack, logger)
		}
	}
}

func last[T interface{}](slice []T) T {
	return slice[len(slice)-1]
}

func lineIsComment(line string) bool {
	commentExpr := regexp.MustCompile(`^[ \t]*/\*.*\*/$`)
	return commentExpr.MatchString(line)
}

func lineCreatesNode(line string) bool {
	return strings.HasSuffix(line, "{")
}

func lineEndsNode(line string) bool {
	closeExpr := regexp.MustCompile(`^[ \t]*}$`)
	return closeExpr.MatchString(line)
}

func lineHasValue(line string) bool {
	// Line must be in the form <name> <value> and not end with an open brace
	valuePattern := regexp.MustCompile(
		`^[[:space:]]*[a-zA-Z0-9-_]+ .*[^{]$`,
	)
	return valuePattern.MatchString(line)
}

// Splits either a key/value pair for a node or the description/name pair for a tagged node
// Since tagged nodes open scopes, also need to strip the open brace at the end of the line
// Presumably any value using it will be quoted, so should not be trimmed
func splitNameFromValue(line string) (string, string) {
	valueDefSplit := regexp.MustCompile(
		`^[[:space:]]*([a-zA-Z0-9-_]+) "?(.*)"?$`,
	)
	parts := valueDefSplit.FindAllStringSubmatch(line, -1)

	return strings.TrimSpace(parts[0][1]),
		strings.TrimSpace(
			strings.TrimRight(
				strings.TrimSpace(parts[0][2]), `{"`),
		)
}

func openScope(
	line string,
	rootNode *Node,
	definitions *Definitions,
	definitionStack *[]*Definition,
	nodeStack *[]*Node,
	logger *log.Logger,
) {
	parentNode := last(*nodeStack)
	// Parent definition may not be set
	var parentDefinition *Definition = nil
	if parentNode != rootNode {
		parentDefinition = last(*definitionStack)
	}

	logger.Debugf("Adding new scope: %s", getDefinitionName(line))
	definition := makeNewDefinition(
		parentNode.Children(),
		parentDefinition,
		getDefinitionName(line),
	)

	if definition.Node != nil && definition.Node.IsTag {
		// The name of the tag will need to be validated, so set it is the value here
		definition.Value = getTagDefinitionName(line)
		logger.Debugf("Scope is tag node, setting value to: %s", definition.Value)
	}

	definitions.add(&definition)
	*definitionStack = append(*definitionStack, &definition)
	if definition.Node == nil {
		logger.Errorf("Did not find templated node for %s", definition.FullPath())
	} else {
		*nodeStack = append(*nodeStack, definition.Node)
		if definition.Node != nil && definition.Node.IsTag {
			// Tag nodes are basically placeholders for children, so skip over it
			*nodeStack = append(*nodeStack, definition.Node.ChildNodes["node.tag"])
			fmt.Printf("Tag node: %#v\n", definition.Node.ChildNodes["node.tag"])
		}
	}

	log.Debug("")
}

func closeScope(
	definitionStack *[]*Definition,
	nodeStack *[]*Node,
	logger *log.Logger,
) {
	logger.Debugf("Closing out scope: %s", last(*definitionStack).FullPath())
	*definitionStack = (*definitionStack)[:len(*definitionStack)-1]
	*nodeStack = (*nodeStack)[:len(*nodeStack)-1]
	// If top of stack is a tag, then we'll need to skip that one too
	// since when opening the tag we appended two nodes
	if len(*nodeStack) > 0 && last(*nodeStack).IsTag {
		logger.Debugf("Scope was tag node, also closing: %s", last(*nodeStack).Name)
		*nodeStack = (*nodeStack)[:len(*nodeStack)-1]
	}
}

func setValue(
	line string,
	definitions *Definitions,
	definitionStack []*Definition,
	nodeStack []*Node,
	logger *log.Logger,
) {
	attribute, value := splitNameFromValue(line)
	logger.Debugf(
		"Detected value '%v' for attribute '%s' on path %s",
		value,
		attribute,
		last(definitionStack).FullPath(),
	)

	definition := makeNewDefinition(last(nodeStack).Children(), last(definitionStack), attribute)
	if _, ok := definitions.DefinitionByPath[definition.FullPath()]; definition.Node.Multi && ok {
		definitions.DefinitionByPath[definition.FullPath()].Values = append(
			definitions.DefinitionByPath[definition.FullPath()].Values,
			value,
		)
	} else if definition.Node.Multi {
		logger.Debug("Attribute is multi-valued, appending to array")
		definitions.add(&definition)
		definition.Values = []any{value}
	} else {
		logger.Debug("Setting single-valued attribute")
		definitions.add(&definition)
		definition.Value = value
	}
}

func handleUnknown(
	line string,
	definitions *Definitions,
	definitionStack []*Definition,
	nodeStack []*Node,
	logger *log.Logger,
) {
	// Try to get a definition anyways
	definition := makeNewDefinition(last(nodeStack).Children(), last(definitionStack), strings.TrimSpace(line))

	// If a node is not found, or has a type, then there should be a value set
	// so should not be here
	if definition.Node == nil || len(definition.Node.Type) > 0 {
		logger.Warnf("Could not determine purpose of line: '%s'", line)
	} else if len(definition.Node.Type) == 0 {
		logger.Debugf("Found implicit boolean attribute on path %s", last(definitionStack).FullPath())
		// Otherwise, this is an implicit boolean by being present and we should
		// keep the definition
		definitions.add(&definition)
	}

}

func makeNewDefinition(nodes []*Node, parentDefinition *Definition, name string) Definition {
	path := []string{}
	if parentDefinition != nil {
		path = append(parentDefinition.Path, parentDefinition.Name)
		// For a tag node, path must also include the actual name of the node
		if parentDefinition.Node.IsTag {
			path = append(path, parentDefinition.Value.(string))
		}
	}
	definition := Definition{
		Name:     name,
		Path:     path,
		Children: make([]*Definition, 0),
	}

	for _, node := range nodes {
		if node.Name == name {
			definition.Node = node
			break
		}
	}

	return definition
}

func getDefinitionName(line string) string {
	// Could either be `firewall {` or `name some-firewall {`
	// Regardless, value is first
	return strings.Split(
		strings.TrimSpace(line), " ",
	)[0]
}

func getTagDefinitionName(line string) string {
	// e.g. `name some-firewall, so value is second to last
	parts := strings.Split(
		strings.TrimSpace(line), " ",
	)

	return parts[len(parts)-2]
}
