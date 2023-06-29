package vyos

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
	Name string
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
	Type string
	Help string
	// If the node is a tag, e.g. allows multiple named entries like firewalls
	// or rule numbers
	IsTag bool
	// If can have multiple options, like with ports in a port group
	// ex. firewall/groups/port-group/node.tag/port/node.def
	Multi bool

	Children map[string]*Node

	// Can have multiple validations, usually with patterns
	Constraints []NodeConstraint
}

// NodeConstraint contains a set of values, command, or pattern the value of the node must satisfy
type NodeConstraint struct {
	// Friendly reason for what this validation checks
	FailureReason string

	// A list of possible options, explicit and static values
	// ex. list: vpn/ipsec/logging/log-modes/node.def
	// ex. list: firewall/name/node.tag/default-action/node.def
	Options []string

	// A command that will generate the possible options
	// This will frequently miss new values, such as a new firewall name or group, so will only show as a warning not a blocking error
	// ex. allowed: vpn/ipsec/remote-access/ike-settings/esp-group/node.def
	// ex. allowed: system/ip/arp/table-size/node.def
	// ex. allowed: interfaces/switch/node.tag/redirect.node.def - TODO also has syntax expression, not sure how to represent
	// ex. allowed: interfaces/switch/node.tag/switch-port/interface/node.def
	// ex. allowed: firewall/name/node.tag/rules/node.tag/protocol/node.def
	OptionsCommand string

	// The VyOS command to run to validate
	// Will take an $VAR(@) parameter somewhere to verify a value is valid
	// ex. exec: firewall/name/node.def
	// ex. allowed: system/ip/arp/table-size/node.def
	// ex. allowed: interfaces/switch/node.tag/switch-port/interface/node.def
	ValidateCommand string

	// RegEx pattern
	// Note that it fails if it does NOT match this pattern
	// e.g. pattern looks for cases where the value does not start with a "-", but
	//      that is actually indicating the value is VALID, so if we want to detect
	//			invalid instances we need to negate that check
	// ex. zone-policy/zone/node.def
	Pattern string

	// Minimum/maximum values for the node
	// ex: vpn/ipsec/esp-group/node.tag/proposal/node.def
	MinBound int
	MaxBound int
}

// ConstraintKey represents the name of a possible VyOS constraint
type ConstraintKey string

const (
	// FailureReason is a human-readable failure message for the constraint
	FailureReason ConstraintKey = "FailureReason"
	// Options contains a complete list of possible values
	Options ConstraintKey = "Options"
	// OptionsCommand generates a list of values, but may not include newly-generated options,
	// like with firewall name, address groups, interfaces, etc
	OptionsCommand ConstraintKey = "OptionsCommand"
	// ValidateCommand is a function to execute with a value to verifu validity
	ValidateCommand ConstraintKey = "ValidateCommand"
	// Pattern is a RegExp pattern to validate a value against
	Pattern ConstraintKey = "Pattern"
	// MinBound is the lowest a numerical value can be
	MinBound ConstraintKey = "MinBound"
	// MaxBound is the largest a numerical value can be
	MaxBound ConstraintKey = "MaxBound"
)

// GetProperty Dynamically looks up the value of a particular constraint value by its key
func (n *NodeConstraint) GetProperty(field ConstraintKey) interface{} {
	switch field {
	case FailureReason:
		return n.FailureReason
	case Options:
		return n.Options
	case OptionsCommand:
		return n.OptionsCommand
	case ValidateCommand:
		return n.ValidateCommand
	case Pattern:
		return n.Pattern
	case MinBound:
		return n.MinBound
	case MaxBound:
		return n.MaxBound
	}

	return nil
}
