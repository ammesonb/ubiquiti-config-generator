package vyos

import (
	"io"
	"regexp"
	"strings"
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
func ParseBootDefinitions(reader *io.Reader, definitions *Definitions, nodes *Node) error {
	return nil
	/*
	* create a node stack indexing by current path
	* for each line:
	* get type of line and the node's name
	* check the current node's children for a matching name
	* add to/modify definition from a stack as well?
	 */
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

// Splits either a key/value pair for a node or the description/name pair for a tagged node
// Since tagged nodes open scopes, also need to strip the open brace at the end of the line
// Presumably any value using it will be quoted, so should not be trimmed
func splitNameFromValue(line string) (string, string) {
	valueDefSplit := regexp.MustCompile(
		`^[[:space:]]*([[:^space:]]+) "?(.*)"?`,
	)
	parts := valueDefSplit.FindAllStringSubmatch(line, -1)

	return strings.TrimSpace(parts[0][1]),
		strings.TrimSpace(
			strings.TrimRight(
				strings.TrimSpace(parts[0][2]), `{"`),
		)
}
