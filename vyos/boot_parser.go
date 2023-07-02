package vyos

import (
	"bufio"
	"io"
	"regexp"
	"strings"

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
func ParseBootDefinitions(reader io.Reader, definitions *Definitions, rootNode *Node) error {
	definitionStack := make([]*Definition, 0)
	nodeStack := []*Node{rootNode}

	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		} else if lineEndsNode(line) {
			definitionStack = definitionStack[:len(definitionStack)-1]
			nodeStack = nodeStack[:len(nodeStack)-1]
		} else if lineIsComment(line) {
			definitionStack[len(definitionStack)-1].Comment = line
		} else if lineCreatesNode(line) {
			definition := makeNewDefinition(
				last(nodeStack).Children(),
				last(definitionStack),
				getDefinitionName(line),
			)
			definitions.add(&definition)
			definitionStack = append(definitionStack, &definition)
			nodeStack = append(nodeStack, definition.Node)
		} else if lineHasValue(line) {
			attribute, value := splitNameFromValue(line)
			definition := makeNewDefinition(last(nodeStack).Children(), last(definitionStack), attribute)
			if _, ok := definitions.DefinitionByPath[definition.FullPath()]; definition.Node.Multi && ok {
				definitions.DefinitionByPath[definition.FullPath()].Values = append(
					definitions.DefinitionByPath[definition.FullPath()].Values,
					value,
				)
			} else if definition.Node.Multi {
				definitions.add(&definition)
				definition.Values = []any{value}
			} else {
				definitions.add(&definition)
				definition.Value = value
			}
		} else {
			log.Warnf("Could not determine purpose of line: '%s'", line)
		}
	}

	return nil
	/*
	* create a node stack indexing by current path
	* for each line:
	* get type of line and the node's name
	* check the current node's children for a matching name
	* add to/modify definition from a stack as well?
	 */
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

func makeNewDefinition(nodes []*Node, parentDefinition *Definition, name string) Definition {
	path := []string{}
	if parentDefinition != nil {
		path = append(parentDefinition.Path, parentDefinition.Name)
	}
	definition := Definition{
		Name: name,
		Path: path,
	}

	for _, node := range nodes {
		if node.IsTag {
			definition.Node = node
			break
		} else if node.Name == name {
			definition.Node = node
			break
		}
	}

	return definition
}

func getDefinitionName(line string) string {
	// Could either be `firewall {` or `name some-firewall {`
	// Regardless, value is second to last
	parts := strings.Split(
		strings.TrimSpace(line), " ",
	)

	return parts[len(parts)-2]
}
