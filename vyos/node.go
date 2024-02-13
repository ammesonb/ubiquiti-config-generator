package vyos

import (
	"fmt"
	"github.com/ammesonb/ubiquiti-config-generator/utils"
	"strings"

	"github.com/ammesonb/ubiquiti-config-generator/console_logger"
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
		* However, in some cases may not yet be defined yet - e.g. firewall rule group address
		  where this now creates a new group
	* $VAR(@) needs to be replaced for both commands/expressions/etc and help text
*/

// Node represents a configurable path or entry in the VyOS template directory
type Node struct {
	Name string `yaml:"Name"`
	/* Type of the node, can be:
	- bool
	- u32
	- txt
	- macaddr
	- ipv4
	- ipv6
	- ipv4net
	- ipv6net
	*/
	// VyOS type for node, like txt or u32
	// Can be omitted, like for services/ssh/disable-password-authentication,
	// where it is used as a boolean - present is true, omitted is false
	Type string `yaml:"Type"`
	// Describes the node's purpose
	Help string `yaml:"Help"`
	// Describes help for the various possible node values
	ValuesHelp []string `yaml:"ValHelp"`
	// If the node is a tag, e.g. allows multiple named entries like firewalls
	// or rule numbers
	IsTag bool `yaml:"IsTag"`
	// If can have multiple options, like with ports in a port group
	// ex. firewall/groups/port-group/node.tag/port/node.def
	Multi bool `yaml:"Multi"`

	// The full path to this node
	Path string `yaml:"Path"`

	ChildNodes map[string]*Node `yaml:"ChildNodes"`

	// Can have multiple validations, usually with patterns
	Constraints []NodeConstraint `yaml:"Constraints"`
}

// Children returns an unordered list of nodes this one contains
func (node *Node) Children() []*Node {
	children := []*Node{}

	for _, child := range node.ChildNodes {
		children = append(children, child)
	}

	return children
}

// FindChild recursively delves into this node's children along the given path
func (node *Node) FindChild(path []string) *Node {
	children := []*Node{node}
	for _, step := range path {
		child, ok := utils.Last(children).ChildNodes[step]
		if !ok {
			console_logger.DefaultLogger().Debugf(
				"Could not find node for step '%s' of path '%s'",
				step,
				strings.Join(path, "/"),
			)
			return nil
		}

		children = append(children, child)
	}

	return utils.Last(children)
}

func (node *Node) ParentPath() string {
	parts := strings.Split(node.Path, "/")
	// parent is all but last entry of the path parts
	parts = utils.AllExcept(parts, 1)
	// If last entry is a placeholder, trim that too
	if utils.Last(parts) == utils.DYNAMIC_NODE {
		parts = utils.AllExcept(parts, 1)
	}

	return strings.Join(parts, "/")
}

func (node *Node) diffNode(other *Node) []string {
	differences := []string{}

	// Skip node child check since that could be expensive
	if node.Name != other.Name {
		differences = append(
			differences,
			fmt.Sprintf(
				"%s: Node 'Name' should be '%s' but got '%s'",
				node.Path,
				node.Name,
				other.Name,
			),
		)
	}
	if node.Type != other.Type {
		differences = append(
			differences,
			fmt.Sprintf(
				"%s: Node 'Type' should be '%s' but got '%s'",
				node.Path,
				node.Type,
				other.Type,
			),
		)
	}
	if node.IsTag != other.IsTag {
		differences = append(
			differences,
			fmt.Sprintf(
				"%s: Node 'IsTag' should be '%t' but got '%t'",
				node.Path,
				node.IsTag,
				other.IsTag,
			),
		)
	}
	if node.Multi != other.Multi {
		differences = append(
			differences,
			fmt.Sprintf(
				"%s: Node 'Multi' should be '%t' but got '%t'",
				node.Path,
				node.Multi,
				other.Multi,
			),
		)
	}
	if node.Path != other.Path {
		differences = append(
			differences,
			fmt.Sprintf(
				"%s: Node 'Path' should be '%s' but got '%s'",
				node.Path,
				node.Path,
				other.Path,
			),
		)
	}

	return differences
}

// NodeConstraint contains a set of values, command, or pattern the value of the node must satisfy
type NodeConstraint struct {
	// Friendly reason for what this validation checks
	FailureReason string `yaml:"FailureReason"`

	// A list of possible options, explicit and static values
	// ex. list: vpn/ipsec/logging/log-modes/node.def
	// ex. list: firewall/name/node.tag/default-action/node.def
	Options []string `yaml:"Options"`

	// A list of disallowed values
	// ex. system/login/user/node.tag/group/node.def
	NegatedOptions []string `yaml:"NegatedOptions"`

	// A command that will generate the possible options
	// This will frequently miss new values, such as a new firewall name or group, so will only show as a warning not a blocking error
	// ex. allowed: vpn/ipsec/remote-access/ike-settings/esp-group/node.def
	// ex. allowed: system/ip/arp/table-size/node.def
	// ex. allowed: interfaces/switch/node.tag/redirect.node.def
	// ex. allowed: interfaces/switch/node.tag/switch-port/interface/node.def
	// ex. allowed: firewall/name/node.tag/rules/node.tag/protocol/node.def
	OptionsCommand string `yaml:"OptionsCommand"`

	// The VyOS command to run to validate
	// Will take an $VAR(@) parameter somewhere to verify a value is valid
	// ex. exec: firewall/name/node.def
	// ex. allowed: system/ip/arp/table-size/node.def
	// ex. allowed: interfaces/switch/node.tag/switch-port/interface/node.def
	ValidateCommand string `yaml:"ValidateCommand"`

	// RegEx pattern
	// Note that it fails if it does NOT match this pattern
	// e.g. pattern looks for cases where the value does not start with a "-", but
	//      that is actually indicating the value is VALID, so if we want to detect
	//			invalid instances we need to negate that check
	// ex. zone-policy/zone/node.def
	Pattern string `yaml:"Pattern"`

	// Like pattern, but negated
	NegatedPattern string `yaml:"NegatedPattern"`

	// Minimum/maximum values for the node
	// ex: vpn/ipsec/esp-group/node.tag/proposal/node.def
	MinBound *int `yaml:"MinBound"`
	MaxBound *int `yaml:"MaxBound"`
}

// ConstraintKey represents the name of a possible VyOS constraint
type ConstraintKey string

const (
	// FailureReason is a human-readable failure message for the constraint
	FailureReason ConstraintKey = "FailureReason"
	// Options contains a complete list of possible values
	Options ConstraintKey = "Options"
	// NegatedOptions contains a complete list of disallowed values
	NegatedOptions ConstraintKey = "NegatedOptions"
	// OptionsCommand generates a list of values, but may not include newly-generated options,
	// like with firewall name, address groups, interfaces, etc
	OptionsCommand ConstraintKey = "OptionsCommand"
	// ValidateCommand is a function to execute with a value to verifu validity
	ValidateCommand ConstraintKey = "ValidateCommand"
	// Pattern is a RegExp pattern to validate a value against
	Pattern ConstraintKey = "Pattern"
	// NegatedPattern is a negated RegExp pattern to validate a value against
	NegatedPattern ConstraintKey = "NegatedPattern"
	// MinBound is the lowest a numerical value can be
	MinBound ConstraintKey = "MinBound"
	// MaxBound is the largest a numerical value can be
	MaxBound ConstraintKey = "MaxBound"
)

var UnsetMinBound = -123456789
var UnsetMaxBound = 123456789

// GetProperty Dynamically looks up the value of a particular constraint value by its key
func (n *NodeConstraint) GetProperty(field ConstraintKey) interface{} {
	switch field {
	case FailureReason:
		return n.FailureReason
	case Options:
		return n.Options
	case NegatedOptions:
		return n.NegatedOptions
	case OptionsCommand:
		return n.OptionsCommand
	case ValidateCommand:
		return n.ValidateCommand
	case Pattern:
		return n.Pattern
	case NegatedPattern:
		return n.NegatedPattern
	case MinBound:
		if n.MinBound == nil {
			// Random value, should only be used for testing anyways
			console_logger.DefaultLogger().Warn("Requested unset minimum bound")
			return UnsetMinBound
		}
		return *n.MinBound
	case MaxBound:
		if n.MaxBound == nil {
			// Random value, should only be used for testing anyways
			console_logger.DefaultLogger().Warn("Requested unset maximum bound")
			return UnsetMaxBound
		}
		return *n.MaxBound
	}

	return nil
}
