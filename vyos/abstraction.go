package vyos

import "github.com/ammesonb/ubiquiti-config-generator/abstraction"

func FromNetworkAbstraction(nodes *Node, network *abstraction.Network) *Definitions {
	definitions := initDefinitions()

	dhcpDefinition := generateDefinitionTree(nodes, []string{"service", "dhcp-server", "shared-network-name"})
	// TODO: rest of DHCP definition
	// TODO: static host mappings
	// TODO: interface, if set
	// TODO: firewall names
	// TODO: assign firewall to interface
	// TODO: firewall rules
	// TODO: firewall rule numbering

	definitions.add(dhcpDefinition)

	return definitions
}

func FromPortGroupAbstraction(nodes *Node, group abstraction.PortGroup) *Definitions {
	// Have to explicitly add each port to a new slice, cannot convert/assert any other way :(
	var ports []any
	for _, port := range group.Ports {
		ports = append(ports, port)
	}

	definitions := initDefinitions()
	startingPath := []string{"firewall", "group", "port-group"}
	groupDefinition := generateDefinitionTree(nodes, startingPath)
	groupDefinition.Children[0].Children[0].Value = group.Name
	groupDefinition.Children[0].Children[0].Children = []*Definition{
		{
			Name:     "description",
			Path:     append(startingPath, group.Name),
			Node:     nodes.FindChild(append(startingPath, "node.tag", "description")),
			Value:    group.Description,
			Children: []*Definition{},
		},
		{
			Name:     "port",
			Path:     []string{"firewall", "group", "port-group", group.Name},
			Node:     nodes.FindChild(append(startingPath, "node.tag", "port")),
			Values:   ports,
			Children: []*Definition{},
		},
	}

	definitions.add(groupDefinition)

	return definitions
}

/*
Name:     "group",
Path:     []string{"firewall"},
Node:     nodes.FindChild([]string{"firewall", "group"}),
Children: []*Definition{},
*/
