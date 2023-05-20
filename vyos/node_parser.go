package vyos

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
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

		Children: make(map[string]*Node),
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

func parseDefinition(reader io.ReadCloser, node *Node) error {
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	scope := ""
	value := ""
	helpValues := &[]string{}
	allowed := ""
	expression := ""
	for scanner.Scan() {
		line := scanner.Text()

		// value needs to be concatenated for each line beginning with space, tab, etc
		// then when done, assign to relevant part of the node and clear the value array

		if !unicode.IsSpace(rune(line[0])) && strings.Contains(line, ":") {
			// If a character starts the line and the line contains a colon, this is a new option
			// Close out the old scope if set, since this one will override it
			if scope != "" {
				addOption(scope, value, helpValues, &allowed, &expression, node)
			}

			scope = strings.Split(line, ":")[0]
			value = line
		} else if len(line) == 0 {
			// Blank lines end scope
			addOption(scope, value, helpValues, &allowed, &expression, node)
			scope = ""
			value = ""
		} else if unicode.IsSpace(rune(line[0])) && strings.Contains(line, ":") {
			// If a space, then just simply append this line to the option value
			value += "\n" + line
		}

	}

	// TODO: constraints

	return nil
}

func addOption(
	scope string,
	value string,
	helpValues *[]string,
	allowed *string,
	expression *string,
	node *Node,
) {
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
			*helpValues, strings.TrimSpace(strings.Split(value, ";")[0]),
		)
	case "allowed":
		*allowed = value
	case "syntax":
		*expression = value
	}

}
