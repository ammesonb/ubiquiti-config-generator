package vyos

import (
	"reflect"
	"strconv"

	"github.com/ammesonb/ubiquiti-config-generator/abstraction"
	"github.com/ammesonb/ubiquiti-config-generator/config"
)

func FromNetworkAbstraction(nodes *Node, network *abstraction.Network, deviceConfig *config.DeviceConfig) *Definitions {
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

	subnetNodePath := append(dhcpPath, DYNAMIC_NODE, "subnet", DYNAMIC_NODE)
	for _, subnet := range network.Subnets {
		subnetPath := append(dhcpPath, network.Name, "subnet", subnet.CIDR)
		addSubnetNetworkValues(nodes, definitions, subnet, subnetPath, subnetNodePath)

		for _, host := range subnet.Hosts {
			addHostToSubnet(nodes, definitions, host, subnetPath, subnetNodePath)
			addHostToAddressGroups(nodes, definitions, host)
			addFirewallRules(nodes, definitions, network, host)
		}
	}

	return definitions
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
func addSubnetNetworkValues(nodes *Node, definitions *Definitions, subnet abstraction.Subnet, path []string, nodePath []string) {
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

func addFirewallRules(nodes *Node, definitions *Definitions, network *abstraction.Network, host *abstraction.Host) {
	for fromPort, toPort := range host.ForwardPorts {
		addForwardPort(nodes, definitions, host, network.InboundInterface, fromPort, toPort)
	}

	// TODO: connections
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
