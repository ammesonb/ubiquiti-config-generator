package vyos

import (
	"reflect"

	"github.com/ammesonb/ubiquiti-config-generator/abstraction"
	"github.com/ammesonb/ubiquiti-config-generator/config"
)

func FromNetworkAbstraction(nodes *Node, network *abstraction.Network) *Definitions {
	definitions := initDefinitions()

	dhcpPath := []string{"service", "dhcp-server", "shared-network-name"}
	dhcpDefinition := generateSparseDefinitionTree(nodes, dhcpPath)
	dhcpDefinition.Children[0].Children[0].Children[0].Value = network.Name
	dhcpDefinition.Children[0].Children[0].Children[0].Children[0].Value = network.Description

	definitions.add(
		generatePopulatedDefinitionTree(
			nodes,
			BasicDefinition{
				Name:  "node.tag",
				Value: network.Name,
				Children: []BasicDefinition{
					{
						Name:  "authoritative",
						Value: network.Authoritative,
					},
					{
						Name:  "description",
						Value: network.Description,
					},
				},
			},
			dhcpPath,
			dhcpPath,
		),
	)

	subnetNodePath := append(dhcpPath, "node.tag", "subnet", "node.tag")
	for _, subnet := range network.Subnets {
		subnetPath := append(dhcpPath, network.Name, "subnet", subnet.CIDR)

		definitions.add(
			generatePopulatedDefinitionTree(
				nodes,
				BasicDefinition{
					Name: "start",
					Children: []BasicDefinition{
						{
							Name:  "node.tag",
							Value: subnet.DHCPStart,
							Children: []BasicDefinition{
								{
									Name: "stop",
									Children: []BasicDefinition{
										{
											Name:  "node.tag",
											Value: subnet.DHCPEnd,
										},
									},
								},
							},
						},
					},
				},
				subnetPath,
				subnetNodePath,
			),
		)

		definitions.add(&Definition{
			Name:  "lease",
			Path:  subnetPath,
			Node:  nodes.FindChild(append(subnetNodePath, "lease")),
			Value: subnet.DHCPLease,
		})

		definitions.add(&Definition{
			Name:   "dns-server",
			Path:   subnetPath,
			Node:   nodes.FindChild(append(subnetNodePath, "dns-server")),
			Values: config.SliceStrToAny(subnet.DNSServers),
		})

		definitions.add(&Definition{
			Name:  "domain-name",
			Path:  subnetPath,
			Node:  nodes.FindChild(append(subnetNodePath, "domain-name")),
			Value: subnet.DomainName,
		})

		definitions.add(&Definition{
			Name:  "default-router",
			Path:  subnetPath,
			Node:  nodes.FindChild(append(subnetNodePath, "default-router")),
			Value: subnet.DefaultRouter,
		})

		for key, val := range subnet.Extra {
			valType := reflect.TypeOf(val).Kind()
			if valType == reflect.Slice || valType == reflect.Array {
				definitions.add(&Definition{
					Name:   key,
					Path:   subnetPath,
					Node:   nodes.FindChild(append(subnetNodePath, key)),
					Values: val.([]any),
				})
			} else {
				definitions.add(&Definition{
					Name:  key,
					Path:  subnetPath,
					Node:  nodes.FindChild(append(subnetNodePath, key)),
					Value: val,
				})
			}

		}
	}

	// TODO: static host mappings
	// TODO: interface, if set
	// TODO: firewall names
	// TODO: assign firewall to interface
	// TODO: firewall rules
	// TODO: firewall rule numbering

	return definitions
}

func FromPortGroupAbstraction(nodes *Node, group abstraction.PortGroup) *Definitions {
	// Have to explicitly add each port to a new slice, cannot convert/assert any other way :(
	definitions := initDefinitions()
	startingPath := []string{"firewall", "group", "port-group"}
	groupDefinition := generateSparseDefinitionTree(nodes, startingPath)
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
			Values:   config.SliceIntToAny(group.Ports),
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
