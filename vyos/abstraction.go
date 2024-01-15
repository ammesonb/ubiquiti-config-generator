package vyos

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/ammesonb/ubiquiti-config-generator/abstraction"
	"github.com/ammesonb/ubiquiti-config-generator/config"
	"github.com/ammesonb/ubiquiti-config-generator/validation"
)

func FromNetworkAbstraction(nodes *Node, network *abstraction.Network, deviceConfig *config.DeviceConfig) (*Definitions, []error) {
	definitions := initDefinitions()

	dhcpPath := []string{"service", "dhcp-server", "shared-network-name"}
	dhcpDefinition := generateSparseDefinitionTree(nodes, dhcpPath)
	dhcpDefinition.Children[0].Children[0].Children[0].Value = network.Name
	dhcpDefinition.Children[0].Children[0].Children[0].Children[0].Value = network.Description

	definitions.add(
		generatePopulatedDefinitionTree(
			nodes,
			BasicDefinition{
				Name:  DYNAMIC_NODE,
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

	if network.Interface != nil {
		configureNetworkInterface(
			nodes,
			definitions,
			network.Interface,
			[]string{"interfaces", "ethernet"},
		)
	}

	errors := make([]error, 0)
	subnetNodePath := append(dhcpPath, DYNAMIC_NODE, "subnet", DYNAMIC_NODE)
	for _, subnet := range network.Subnets {
		subnetPath := append(dhcpPath, network.Name, "subnet", subnet.CIDR)
		addSubnetNetworkValues(nodes, definitions, subnet, subnetPath, subnetNodePath)

		for _, host := range subnet.Hosts {
			addHostToSubnet(nodes, definitions, host, subnetPath, subnetNodePath)
			addHostToAddressGroups(nodes, definitions, host)
			if err := addFirewallRules(nodes, definitions, network, subnet, host); err != nil {
				errors = append(errors, err...)
			}
		}
	}

	return definitions, errors
}

// configureNetworkInterface sets properties of an interface including address, firewalls, etc
func configureNetworkInterface(nodes *Node, definitions *Definitions, iface *abstraction.Interface, path []string) {
	nodePath := append(path, DYNAMIC_NODE)
	path = append(path, iface.Name)

	definitions.addValue(nodes, path, nodePath, "duplex", iface.Duplex)
	definitions.addValue(nodes, path, nodePath, "speed", iface.Speed)

	if iface.Vif != nil {
		// For interfaces with virtual interfaces/VLANs, set the description to CARRIER automatically
		definitions.addValue(nodes, path, nodePath, "description", "CARRIER")
		path = append(path, "vif", strconv.Itoa(int(*iface.Vif)))
		nodePath = append(nodePath, "vif", DYNAMIC_NODE)
	}

	definitions.addValue(nodes, path, nodePath, "description", iface.Description)
	definitions.addValue(nodes, path, nodePath, "address", iface.Address)

	addExtraValues(nodes, definitions, path, nodePath, iface.Extra)

	path = append(path, "firewall", "in")
	nodePath = append(nodePath, "firewall", "in")
	definitions.addValue(nodes, path, nodePath, "name", iface.InboundFirewall)

	path[len(path)-1] = "out"
	nodePath[len(nodePath)-1] = "out"
	definitions.addValue(nodes, path, nodePath, "name", iface.OutboundFirewall)

	path[len(path)-1] = "local"
	nodePath[len(nodePath)-1] = "local"
	definitions.addValue(nodes, path, nodePath, "name", iface.LocalFirewall)
}

// addSubnetNetworkValues configures the given subnet definitions using the provided nodes and paths
func addSubnetNetworkValues(nodes *Node, definitions *Definitions, subnet *abstraction.Subnet, path []string, nodePath []string) {
	definitions.add(
		generatePopulatedDefinitionTree(
			nodes,
			BasicDefinition{
				Name: "start",
				Children: []BasicDefinition{
					{
						Name:  DYNAMIC_NODE,
						Value: subnet.DHCPStart,
						Children: []BasicDefinition{
							{
								Name: "stop",
								Children: []BasicDefinition{
									{
										Name:  DYNAMIC_NODE,
										Value: subnet.DHCPEnd,
									},
								},
							},
						},
					},
				},
			},
			path,
			nodePath,
		),
	)

	definitions.addValue(nodes, path, nodePath, "lease", subnet.DHCPLease)
	definitions.addListValue(nodes, path, nodePath, "dns-server", config.SliceStrToAny(subnet.DNSServers))
	definitions.addValue(nodes, path, nodePath, "domain-name", subnet.DomainName)
	definitions.addValue(nodes, path, nodePath, "default-router", subnet.DefaultRouter)

	addExtraValues(nodes, definitions, path, nodePath, subnet.Extra)
}

// addHostToSubnet configures the reserved IP address for a host with the given MAC address
func addHostToSubnet(nodes *Node, definitions *Definitions, host *abstraction.Host, path []string, nodePath []string) {
	nodePath = append(nodePath, "static-mapping", DYNAMIC_NODE)
	path = append(path, "static-mapping", host.Name)

	definitions.addValue(nodes, path, nodePath, "ip-address", host.Address)
	definitions.addValue(nodes, path, nodePath, "mac-address", host.MAC)
}

func addExtraValues(nodes *Node, definitions *Definitions, path []string, nodePath []string, extra map[string]any) {
	for key, val := range extra {
		valType := reflect.TypeOf(val).Kind()
		if valType == reflect.Slice || valType == reflect.Array {
			definitions.addListValue(nodes, path, nodePath, key, val.([]any))
		} else {
			definitions.addValue(nodes, path, nodePath, key, val)
		}
	}
}

// addHostToAddressGroups will add the address of the host to any specified groups it should have
func addHostToAddressGroups(nodes *Node, definitions *Definitions, host *abstraction.Host) {
	if host.AddressGroups == nil {
		return
	}

	path := []string{"firewall", "group", "address-group"}
	nodePath := []string{"firewall", "group", "address-group", "node.tag"}
	for _, group := range host.AddressGroups {
		definitions.appendToListValue(nodes, append(path, group), nodePath, "address", host.Address)
	}
}

func addFirewallRules(nodes *Node, definitions *Definitions, network *abstraction.Network, subnet *abstraction.Subnet, host *abstraction.Host) []error {
	for fromPort, toPort := range host.ForwardPorts {
		addForwardPort(nodes, definitions, host, network.InboundInterface, fromPort, toPort)
	}

	if len(host.Connections) > 0 && network.Interface == nil {
		return []error{fmt.Errorf("network %s must have interface defined in order to set firewall rules on host %s", network.Name, host.Name)}
	}

	errors := make([]error, 0)
	for _, connection := range host.Connections {
		if err := addConnection(
			nodes,
			definitions,
			network,
			getConnectionFirewall(network, subnet, host, connection),
			connection,
		); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

func addForwardPort(nodes *Node, definitions *Definitions, host *abstraction.Host, inboundInterface string, from int32, to int32) {
	ruleNumber := abstraction.GetCounter(abstraction.NAT_COUNTER).Next()

	path := []string{"service", "nat", "rule", strconv.Itoa(ruleNumber)}
	nodePath := []string{"service", "nat", "rule", "node.tag"}

	definitions.addValue(nodes, path, nodePath, "type", "destination")
	definitions.addValue(nodes, path, nodePath, "protocol", "tcp_udp")
	definitions.addValue(nodes, path, nodePath, "inbound-interface", inboundInterface)
	definitions.addValue(nodes, append(path, "destination"), append(nodePath, "destination"), "port", from)

	path = append(path, "inside-address")
	nodePath = append(nodePath, "inside-address")
	definitions.addValue(nodes, path, nodePath, "address", host.Address)
	definitions.addValue(nodes, path, nodePath, "port", to)
}

func addConnection(
	nodes *Node,
	definitions *Definitions,
	network *abstraction.Network,
	firewall string,
	connection abstraction.FirewallConnection,
) error {
	if firewall == "" {
		return fmt.Errorf(
			"unable to determine firewall direction for connection %s in network %s, no local address detected in connections",
			network.Name,
			connection.Description,
		)
	}

	rule := abstraction.GetCounter(firewall).Next()

	rulePath := []string{
		"firewall", "name", firewall, "rule", strconv.Itoa(rule),
	}
	ruleNodePath := []string{
		"firewall", "name", DYNAMIC_NODE, "rule", DYNAMIC_NODE,
	}

	var action, log string
	if connection.Allow {
		action = "accept"
	} else {
		action = "reject"
	}

	if connection.Log {
		log = "enable"
	} else {
		log = "disable"
	}

	definitions.addValue(nodes, rulePath, ruleNodePath, "action", action)
	definitions.addValue(nodes, rulePath, ruleNodePath, "description", connection.Description)
	definitions.addValue(nodes, rulePath, ruleNodePath, "protocol", connection.Protocol)
	definitions.addValue(nodes, rulePath, ruleNodePath, "log", log)

	addConnectionDetail(
		definitions,
		nodes,
		append(rulePath, "source"),
		append(ruleNodePath, "source"),
		connection.Source,
	)
	addConnectionDetail(
		definitions,
		nodes,
		append(rulePath, "destination"),
		append(ruleNodePath, "destination"),
		connection.Destination,
	)

	return nil
}

func getConnectionFirewall(network *abstraction.Network, subnet *abstraction.Subnet, host *abstraction.Host, connection abstraction.FirewallConnection) string {
	var sourceLocal, destLocal bool
	// Must define an address for source to be local
	sourceValid := connection.Source != nil && connection.Source.Address != nil
	if !sourceValid {
		sourceLocal = false
	} else {
		// since address may be an address group, ignore errors about invalid IPs
		sourceLocal, _ = validation.IsAddressInSubnet(*connection.Source.Address, subnet.CIDR)
		sourceLocal = config.InSlice(connection.Source.Address, config.SliceStrToAny(host.AddressGroups)) || sourceLocal
	}

	// Must define an address for destination to be local
	destValid := connection.Destination != nil && connection.Destination.Address != nil
	if !destValid {
		// since address may be an address group, ignore errors about invalid IPs
		destLocal, _ = validation.IsAddressInSubnet(*connection.Destination.Address, subnet.CIDR)
		destLocal = config.InSlice(connection.Destination.Address, config.SliceStrToAny(host.AddressGroups)) || destLocal
	}

	if sourceLocal && destLocal {
		// local if both source and destination addresses are in the network subnet
		return network.Interface.LocalFirewall
	} else if sourceLocal {
		// inbound if source matches host or address group of host, since traffic is _inbound_ to the router port
		return network.Interface.InboundFirewall
	} else if destLocal {
		// outbound if destination matches host or address group of host, since traffic is _outbound_ from the router port
		return network.Interface.OutboundFirewall
	}

	return ""
}

func addConnectionDetail(definitions *Definitions, nodes *Node, path []string, nodePath []string, connection *abstraction.ConnectionDetail) {
	if connection != nil {
		if connection.Address != nil {
			if validation.IsValidAddress(*connection.Address) {
				// For valid IP address, just add it directly
				definitions.addValue(nodes, path, nodePath, "address", connection.Address)
			} else if connection.Address != nil {
				// Group needs to nest further
				definitions.addValue(
					nodes,
					append(path, "group"),
					append(nodePath, "group"),
					"address-group",
					connection.Address,
				)
			}
		}

		if connection.Port != nil {
			port, err := strconv.Atoi(*connection.Port)
			if err == nil {
				definitions.addValue(nodes, path, nodePath, "port", port)
			} else {
				// Group needs to nest further
				definitions.addValue(
					nodes,
					append(path, "group"),
					append(nodePath, "group"),
					"port-group",
					connection.Port,
				)
			}

		}
	}
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
			Node:     nodes.FindChild(append(startingPath, DYNAMIC_NODE, "description")),
			Value:    group.Description,
			Children: []*Definition{},
		},
		{
			Name:     "port",
			Path:     []string{"firewall", "group", "port-group", group.Name},
			Node:     nodes.FindChild(append(startingPath, DYNAMIC_NODE, "port")),
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
