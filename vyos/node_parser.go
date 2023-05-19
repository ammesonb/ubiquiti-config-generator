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

// ParseNodeDef takes a template path and converts it into a list of nodes for analysis/validation
func ParseNodeDef(templatesPath string) ([]Node, error) {
	return nil, nil
}
