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
*/

// Node represents a configurable path or entry in the VyOS template directory
type Node struct {
	Name string
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

	// An allowed value, list or possibly a command
	// If command format, may be able to use val_help instead
	// ex. list: vpn/ipsec/remote-access/ike-settings/esp-group/node.def
	// ex. command: firewall/name/node.tag/default-action/node.def
	// ex. command: system/ip/arp/table-size/node.def
	// ex. command: interfaces/switch/node.tag/redirect.node.def
	// ex. command: interfaces/switch/node.tag/switch-port/interface/node.def
	// ex. command: firewall/name/node.tag/rules/node.tag/protocol/node.def
	Allowed []string

	// The VyOS command to run to validate
	// ex. script: firewall/name/node.def
	// ex: list: firewall/name/node.tag/default-action/node.def
	// ex. list: vpn/ipsec/logging/log-modes/node.def
	// ex: arithmetic vpn/ipsec/esp-group/node.tag/proposal/node.def
	Command string

	// RegEx pattern
	// ex. zone-policy/zone/node.def
	Pattern string
}
