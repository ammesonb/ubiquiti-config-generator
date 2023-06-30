package vyos

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

/** Config notes:
  * REFERENCE: https://docs.vyos.io/en/crux/contributing/vyos_cli.html#mapping-old-node-def-style-to-new-xml-definitions
  * Parser needs to read multiple lines, though looks like blank will mean new property
  * :val_help specifies possible values for a given node, only _one_ of which may be used
    * Represents both enums or possible formats
    * e.g. for ports in firewall port group, could be service like http, number, or number range (1000-1050)
    * may have an underscore prefixing it
  * :multi means multiple values for the _current_ node, e.g. ports in a firewall port group NOT rule numbers, etc
  * node.tag is a directory but represents any given name
  * node.def contains values for the given node
    * But if the parent is a node.tag, then that applies to the path name instead
  * Validation should be applied to the configuration after homogenization into the VyOS configuration
  * allowed will be list of acceptable values - no syntax expression for these
    * seems to be identical to the val_help parameters if multi-line
    * may also be a command (or local -a array, same thing)?
      e.g. interfaces/switch/node.tag/redirect/node.def
      e.g. interfaces/switch/node.tag/switch-port/interface/node.def
*/

// ParseNodeDef takes a template path and converts it into a list of nodes for analysis/validation
func ParseNodeDef(templatesPath string) (*Node, error) {
	// ReadDir returns relative paths, so /etc will return hosts, passwd, shadow, etc
	// Not including the parent `/etc/` prefix
	entries, err := ioutil.ReadDir(templatesPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read files in %s: %s", templatesPath, err.Error())
	}

	node := &Node{
		Name:  filepath.Base(templatesPath),
		IsTag: false,
		Multi: false,
		Path:  templatesPath,

		Children:    make(map[string]*Node),
		Constraints: []NodeConstraint{},
	}

	// If node.def in entries, update node from this path
	// For directories, recurse and extend nodes
	for _, entry := range entries {
		if entry.Name() == "node.def" {
			reader, err := openDefinitionFile(filepath.Join(templatesPath, entry.Name()))
			if err != nil {
				return nil, fmt.Errorf("Failed opening node.def: %s", err.Error())
			}

			if err = parseDefinition(reader, node); err != nil {
				return nil, err
			}
			reader.Close()
			continue
		} else if !entry.IsDir() {
			continue
		}

		childNode, err := ParseNodeDef(filepath.Join(templatesPath, entry.Name()))
		if err != nil {
			return nil, err
		}

		node.Children[entry.Name()] = childNode
	}

	return node, nil
}

func openDefinitionFile(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func parseDefinition(reader io.Reader, node *Node) error {
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	scope := ""
	value := ""
	helpValues := []string{}
	allowed := ""
	expression := ""
	for scanner.Scan() {
		line := scanner.Text()

		// value needs to be concatenated for each line beginning with space, tab, etc
		// then when done, assign to relevant part of the node and clear the value array

		if len(line) == 0 {
			// Blank lines end scope
			addOption(scope, value, &helpValues, &allowed, &expression, node)
			scope = ""
			value = ""
		} else if !unicode.IsSpace(rune(line[0])) && strings.Contains(line, ":") {
			// If a character starts the line and the line contains a colon, this is a new option
			// Close out the old scope if set, since this one will override it
			if scope != "" {
				addOption(scope, value, &helpValues, &allowed, &expression, node)
			}

			scope = strings.Split(line, ":")[0]
			value = line
		} else if unicode.IsSpace(rune(line[0])) {
			// If a space, then just simply append this line to the option value
			value += "\n" + line
		}
	}

	if len(value) > 0 {
		addOption(scope, value, &helpValues, &allowed, &expression, node)
	}

	if !parseConstraints(node, expression) {
		fmt.Printf("Expression in node %s did not match any parser: %s\n", node.Path, expression)
	}

	return nil
}

func parseConstraints(node *Node, expression string) bool {
	if expression == "" {
		return true
	}

	helpSplit := regexp.MustCompile(` ?; ?\\?[[:space:]]*"`)
	if len(helpSplit.FindAllStringIndex(expression, -1)) > 1 {
		fmt.Printf("Got extra semicolons in value: %s\n", expression)
	}

	parts := helpSplit.Split(expression, 2)
	expr := strings.TrimSpace(parts[0])
	help := ""
	if len(parts) > 1 {
		help = stripCommand(parts[1])
	}

	done := skipPath(expression)
	done = done || checkRange(expr, help, node)
	done = done || addPattern(expr, help, node)
	done = done || addNegatedPattern(expr, help, node)
	done = done || addExec(expr, help, node)
	done = done || addExprList(expr, help, node)
	done = done || addNegatedExprList(expr, help, node)
	done = done || addAllowedCommand(expr, help, node)
	done = done || addNumericalInfinity(expr, help, node)

	return done
}

func addOption(
	scope string,
	value string,
	helpValues *[]string,
	allowed *string,
	expression *string,
	node *Node,
) {
	expr := regexp.MustCompile("^[[:word:]:]+:")
	value = strings.TrimSpace(expr.ReplaceAllString(value, ""))
	switch scope {
	case "tag":
		node.IsTag = true
	case "multi":
		node.Multi = true
	case "type":
		node.Type = strings.TrimSpace(value)
	case "help":
		node.Help = strings.TrimSpace(value)
	case "val_help":
		*helpValues = append(
			// For val_help, strip the description after the semi colon
			*helpValues, strings.TrimSpace(value),
		)
	case "allowed":
		*allowed = value
	case "syntax":
		*expression = value
	}
}

func stripCommand(command string) string {
	return strings.TrimSpace(
		strings.Trim(
			strings.TrimSpace(
				strings.Trim(
					command,
					`"\`,
				),
			),
			`"\`,
		),
	)
}

func skipPath(expression string) bool {
	// These check values of parent nodes, which is tricky to evaluate
	// in this structure since it is frequently mixed in with commands
	// and other syntax, so since there are relatively few of them skip them for now
	return strings.Contains(expression, "VAR(../")
}

func checkRange(expression string, help string, node *Node) bool {
	/*
		The expression captures something of the form VAR >= min && VAR <= max
		But spaces are sometimes omitted, and rarely both terms have parens around them
		(VAR>=min)&&(VAR<=max)
		Also sometimes only a min or max value is present, so check for them individually instead of combined
	*/
	min := checkMinBound(expression, node)
	max := checkMaxBound(expression, node)

	if min == nil && max == nil {
		return false
	}

	node.Constraints = append(
		node.Constraints,
		NodeConstraint{
			FailureReason: help,
			MinBound:      min,
			MaxBound:      max,
		},
	)
	return true
}

func checkMinBound(expression string, node *Node) *int {
	boundsExpr := regexp.MustCompile(`\(?\$VAR\(@\) ?(>=?) ?([0-9]+)\)?`)
	result := boundsExpr.FindStringSubmatch(expression)
	if len(result) < 2 {
		return nil
	}

	// Since regex match will always only contain numbers, can assume no errors
	min, _ := strconv.Atoi(result[2])
	if result[1] == ">" {
		min++
	}
	return &min
}

func checkMaxBound(expression string, node *Node) *int {
	boundsExpr := regexp.MustCompile(`\(?\$VAR\(@\) ?(<=?) ?([0-9]+)\)?`)
	result := boundsExpr.FindStringSubmatch(expression)
	if len(result) < 2 {
		return nil
	}

	// Since regex match will always only contain numbers, can assume no errors
	max, _ := strconv.Atoi(result[2])
	if result[1] == "<" {
		max--
	}
	return &max
}

func addPattern(expression string, help string, node *Node) bool {
	if !strings.HasPrefix(expression, "pattern ") {
		return false
	}

	node.Constraints = append(node.Constraints,
		NodeConstraint{
			FailureReason: help,
			Pattern: strings.Split(
				strings.Split(expression, "pattern ")[1],
				"\"",
			)[1],
		})

	return true
}

func addNegatedPattern(expression string, help string, node *Node) bool {
	if !strings.HasPrefix(expression, "! pattern ") {
		return false
	}

	node.Constraints = append(node.Constraints,
		NodeConstraint{
			FailureReason: help,
			NegatedPattern: strings.Split(
				strings.Split(expression, "! pattern ")[1],
				"\"",
			)[1],
		})

	return true
}

func addExec(expression string, help string, node *Node) bool {
	if !strings.HasPrefix(expression, "exec ") {
		return false
	}

	node.Constraints = append(node.Constraints,
		NodeConstraint{
			// Help will frequently be blank for commands since the command will output
			// the failure reason
			FailureReason: help,
			// Commands are usually contained in quotes on the left and right,
			// so strip those
			ValidateCommand: stripCommand(
				strings.Split(expression, "exec ")[1],
			),
		})

	return true
}

func addExprList(expression string, help string, node *Node) bool {
	if !strings.HasPrefix(expression, "$VAR(@) in ") {
		return false
	}

	options := make([]string, 0)

	optionRegex := regexp.MustCompile(`"([[:alnum:]]+)"`)
	for _, option := range optionRegex.FindAllStringSubmatch(
		strings.Split(expression, " in ")[1],
		-1,
	) {
		options = append(options, option[1])
	}

	node.Constraints = append(node.Constraints,
		NodeConstraint{
			FailureReason: help,
			Options:       options,
		})

	return true
}

func addNegatedExprList(expression string, help string, node *Node) bool {
	if !strings.HasPrefix(expression, "! $VAR(@) in ") {
		return false
	}

	options := make([]string, 0)

	optionRegex := regexp.MustCompile(`"([[:alnum:]]+)"`)
	for _, option := range optionRegex.FindAllStringSubmatch(
		strings.Split(expression, " in ")[1],
		-1,
	) {
		options = append(options, option[1])
	}

	node.Constraints = append(node.Constraints,
		NodeConstraint{
			FailureReason:  help,
			NegatedOptions: options,
		})

	return true
}
func addAllowedCommand(expression string, help string, node *Node) bool {
	if !strings.HasPrefix(expression, "allowed:") {
		return false
	}

	command := strings.Split(expression, "allowed:")[1]

	node.Constraints = append(node.Constraints,
		NodeConstraint{
			FailureReason: help,
			// Allowed commands appear to not use double quotes, so only need to strip
			// spaces, not backslashes and other punctuation
			OptionsCommand: strings.TrimSpace(command),
		})

	return true
}

func addNumericalInfinity(expression string, help string, node *Node) bool {
	if expression != `($VAR(@) == "infinity" || (pattern $VAR(@) "[0-9]*"))` {
		return false
	}

	node.Constraints = append(node.Constraints,
		NodeConstraint{
			FailureReason: help,
			Options:       []string{"infinity"},
			Pattern:       "[0-9]*",
		})

	return true
}

// TODO: this likely breaks it
// interfaces/switch/node.tag/redirect/node.def
// syntax:expression: $VAR(@) != "$VAR(../@)" ;\
//         "interface $VAR(../@): redirect to same interface not allowed"
