package vyos

import (
	"fmt"
	"github.com/ammesonb/ubiquiti-config-generator/utils"
	"reflect"
	"sort"
	"strconv"

	"github.com/ammesonb/ubiquiti-config-generator/abstraction"
	"github.com/ammesonb/ubiquiti-config-generator/config"
	"github.com/ammesonb/ubiquiti-config-generator/validation"
)

func FromNetworkAbstraction(nodes *Node, network *abstraction.Network) (*Definitions, []error) {
	definitions := initDefinitions()

	dhcpPath := utils.MakeVyosPath()
	dhcpPath.Append(
		utils.MakeVyosPC("service"),
		utils.MakeVyosPC("dhcp-server"),
		utils.MakeVyosPC("shared-network-name"),
		utils.MakeVyosDynamicPC(network.Name),
	)
	dhcpDefinition := generateSparseDefinitionTree(nodes, dhcpPath)
	definitions.add(dhcpDefinition)

	definitions.addValue(nodes, dhcpPath, "authoritative", network.Authoritative)
	definitions.addValue(nodes, dhcpPath, "description", network.Description)

	if network.Interface != nil {
		configureNetworkInterface(
			nodes,
			definitions,
			network.Interface,
		)
	}

	errors := make([]error, 0)
	subnetBase := dhcpPath.Extend(utils.MakeVyosPC("subnet"))
	for _, subnet := range network.Subnets {
		subnetPath := subnetBase.Extend(utils.MakeVyosDynamicPC(subnet.CIDR))
		definitions.add(generateSparseDefinitionTree(nodes, subnetPath))
		addSubnetNetworkValues(nodes, definitions, subnet, subnetPath)

		for _, host := range subnet.Hosts {
			addHostToSubnet(nodes, definitions, host, subnetPath)
			addHostToAddressGroups(nodes, definitions, host)
			if err := addFirewallRules(nodes, definitions, network, subnet, host); err != nil {
				errors = append(errors, err...)
			}
		}
	}

	return definitions, errors
}

// configureNetworkInterface sets properties of an interface including address, firewalls, etc
func configureNetworkInterface(nodes *Node, definitions *Definitions, iface *abstraction.Interface) {
	path := utils.MakeVyosPath()
	path.Append(
		utils.MakeVyosPC("interfaces"),
		utils.MakeVyosPC("ethernet"),
		utils.MakeVyosDynamicPC(iface.Name),
	)
	// initialize the interfaces/ethernet tree, and override the dynamic node with the proper value
	ethInterfaces := generateSparseDefinitionTree(nodes, path)
	definitions.add(ethInterfaces)

	definitions.addValue(nodes, path, "duplex", iface.Duplex)
	definitions.addValue(nodes, path, "speed", iface.Speed)

	if iface.Vif != nil {
		// For interfaces with virtual interfaces/VLANs, set the description to CARRIER automatically
		definitions.addValue(nodes, path, "description", "CARRIER")

		// Initialize the vif dynamic node, updating path in place since
		path.Append(
			utils.MakeVyosPC("vif"),
			utils.MakeVyosDynamicPC(strconv.Itoa(int(*iface.Vif))),
		)
		vifDef := generateSparseDefinitionTree(nodes, path)
		definitions.add(vifDef)
	}

	definitions.addValue(nodes, path, "description", iface.Description)
	definitions.addValue(nodes, path, "address", iface.Address)

	addExtraValues(nodes, definitions, path, iface.Extra)

	inPath := path.Extend(
		utils.MakeVyosPC("firewall"),
		utils.MakeVyosPC("in"),
	)
	definitions.add(generateSparseDefinitionTree(nodes, inPath))
	definitions.addValue(nodes, inPath, "name", iface.InboundFirewall)

	outPath := inPath.DivergeFrom(1, utils.MakeVyosPC("out"))
	definitions.add(generateSparseDefinitionTree(nodes, outPath))
	definitions.addValue(nodes, outPath, "name", iface.OutboundFirewall)

	localPath := inPath.DivergeFrom(1, utils.MakeVyosPC("local"))
	definitions.add(generateSparseDefinitionTree(nodes, localPath))
	definitions.addValue(nodes, localPath, "name", iface.LocalFirewall)
}

// addSubnetNetworkValues configures the given subnet definitions using the provided nodes and paths
func addSubnetNetworkValues(nodes *Node, definitions *Definitions, subnet *abstraction.Subnet, path *utils.VyosPath) {
	definitions.add(generateSparseDefinitionTree(nodes, path))
	definitions.add(
		generatePopulatedDefinitionTree(
			nodes,
			BasicDefinition{
				Name:  "start",
				Value: subnet.DHCPStart,
				Children: []BasicDefinition{
					{
						Name:  "stop",
						Value: subnet.DHCPEnd,
					},
				},
			},
			path,
			// Get the parent node, stripping the final placeholder from the dynamic tag placeholder
			nodes.FindChild(utils.AllExcept(path.NodePath, 1)),
		),
	)

	definitions.addValue(nodes, path, "lease", subnet.DHCPLease)
	definitions.addListValue(nodes, path, "dns-server", config.SliceStrToAny(subnet.DNSServers))
	definitions.addValue(nodes, path, "domain-name", subnet.DomainName)
	definitions.addValue(nodes, path, "default-router", subnet.DefaultRouter)

	addExtraValues(nodes, definitions, path, subnet.Extra)
}

// addHostToSubnet configures the reserved IP address for a host with the given MAC address
func addHostToSubnet(nodes *Node, definitions *Definitions, host *abstraction.Host, path *utils.VyosPath) {
	hostPath := path.Extend(
		utils.MakeVyosPC("static-mapping"),
		utils.MakeVyosDynamicPC(host.Name),
	)

	definitions.add(generateSparseDefinitionTree(nodes, hostPath))

	definitions.addValue(nodes, hostPath, "ip-address", host.Address)
	definitions.addValue(nodes, hostPath, "mac-address", host.MAC)
}

func addExtraValues(nodes *Node, definitions *Definitions, path *utils.VyosPath, extra map[string]any) {
	for key, val := range extra {
		valType := reflect.TypeOf(val).Kind()
		if valType == reflect.Slice || valType == reflect.Array {
			definitions.addListValue(nodes, path, key, val.([]any))
		} else {
			definitions.addValue(nodes, path, key, val)
		}
	}
}

// addHostToAddressGroups will add the address of the host to any specified groups it should have
func addHostToAddressGroups(nodes *Node, definitions *Definitions, host *abstraction.Host) {
	if host.AddressGroups == nil {
		return
	}

	path := utils.MakeVyosPath()
	path.Append(
		utils.MakeVyosPC("firewall"),
		utils.MakeVyosPC("group"),
		utils.MakeVyosPC("address-group"),
	)
	for _, group := range host.AddressGroups {
		groupPath := path.Extend(utils.MakeVyosDynamicPC(group))
		definitions.add(generateSparseDefinitionTree(nodes, groupPath))
		definitions.appendToListValue(nodes, groupPath, "address", host.Address)
	}
}

func addFirewallRules(nodes *Node, definitions *Definitions, network *abstraction.Network, subnet *abstraction.Subnet, host *abstraction.Host) []error {
	// ensure consistent ordering of ports
	fromPorts := make([]int, 0)
	for p := range host.ForwardPorts {
		fromPorts = append(fromPorts, int(p))
	}
	sort.Ints(fromPorts)

	for _, fromPort := range fromPorts {
		from := int32(fromPort)
		addForwardPort(nodes, definitions, host, network.InboundInterface, from, host.ForwardPorts[from])
	}

	if len(host.Connections) > 0 && network.Interface == nil {
		return []error{fmt.Errorf("network %s must have interface defined in order to set firewall rules on host %s", network.Name, host.Name)}
	}

	errors := make([]error, 0)
	for _, connection := range host.Connections {
		firewallName := getConnectionFirewall(network, subnet, host, connection)
		// if firewall name is blank, then some part of the rule definition is likely invalid since it does not seem to
		// apply to the host that it is written for
		if firewallName == "" {
			errors = append(
				errors,
				fmt.Errorf("could not identify firewall for connection '%s' on host '%s'", connection.Description, host.Name),
			)
			continue
		}

		if err := addConnection(
			nodes,
			definitions,
			network,
			firewallName,
			connection,
		); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

func addForwardPort(nodes *Node, definitions *Definitions, host *abstraction.Host, inboundInterface string, from int32, to int32) {
	ruleNumber := abstraction.GetCounter(abstraction.NAT_COUNTER).Next()

	path := utils.MakeVyosPath()
	path.Append(
		utils.MakeVyosPC("service"),
		utils.MakeVyosPC("nat"),
		utils.MakeVyosPC("rule"),
		utils.MakeVyosDynamicPC(strconv.Itoa(ruleNumber)),
	)

	definitions.add(generateSparseDefinitionTree(nodes, path))
	definitions.addValue(
		nodes,
		path,
		"description",
		fmt.Sprintf("Forward from port %d to %s port %d", from, host.Name, to),
	)
	definitions.addValue(nodes, path, "type", "destination")
	definitions.addValue(nodes, path, "protocol", "tcp_udp")
	definitions.addValue(nodes, path, "inbound-interface", inboundInterface)
	destPath := path.Extend(utils.MakeVyosPC("destination"))
	definitions.add(generateSparseDefinitionTree(nodes, destPath))
	definitions.addValue(nodes, destPath, "port", from)

	insidePath := path.Extend(utils.MakeVyosPC("inside-address"))
	definitions.add(generateSparseDefinitionTree(nodes, insidePath))
	definitions.addValue(nodes, insidePath, "address", host.Address)
	definitions.addValue(nodes, insidePath, "port", to)
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

	rulePath := utils.MakeVyosPath()
	rulePath.Append(
		utils.MakeVyosPC("firewall"),
		utils.MakeVyosPC("name"),
		utils.MakeVyosDynamicPC(firewall),
		utils.MakeVyosPC("rule"),
		utils.MakeVyosDynamicPC(strconv.Itoa(rule)),
	)

	definitions.add(generateSparseDefinitionTree(nodes, rulePath))

	var action, log string
	if connection.Allow {
		action = "accept"
	} else {
		action = "drop"
	}

	if connection.Log {
		log = "enable"
	} else {
		log = "disable"
	}

	definitions.addValue(nodes, rulePath, "action", action)
	definitions.addValue(nodes, rulePath, "description", connection.Description)
	definitions.addValue(nodes, rulePath, "protocol", connection.Protocol)
	definitions.addValue(nodes, rulePath, "log", log)

	addConnectionDetail(
		definitions,
		nodes,
		rulePath.Extend(utils.MakeVyosPC("source")),
		connection.Source,
	)
	addConnectionDetail(
		definitions,
		nodes,
		rulePath.Extend(utils.MakeVyosPC("destination")),
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
		sourceLocal = config.InSlice(*connection.Source.Address, config.SliceStrToAny(host.AddressGroups)) || sourceLocal
	}

	// Must define an address for destination to be local
	destValid := connection.Destination != nil && connection.Destination.Address != nil
	if destValid {
		// since address may be an address group, ignore errors about invalid IPs
		destLocal, _ = validation.IsAddressInSubnet(*connection.Destination.Address, subnet.CIDR)
		destLocal = config.InSlice(*connection.Destination.Address, config.SliceStrToAny(host.AddressGroups)) || destLocal
	}

	// after determining locality of source and destination, can determine appropriate firewall to use
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

func addConnectionDetail(definitions *Definitions, nodes *Node, path *utils.VyosPath, connection *abstraction.ConnectionDetail) {
	if connection != nil {
		definitions.add(generateSparseDefinitionTree(nodes, path))
		if connection.Address != nil {
			if validation.IsValidAddress(*connection.Address) {
				// For valid IP address, just add it directly
				definitions.addValue(nodes, path, "address", connection.Address)
			} else if connection.Address != nil {
				// Group needs to nest further
				groupPath := path.Extend(utils.MakeVyosPC("group"))
				definitions.add(generateSparseDefinitionTree(nodes, groupPath))
				definitions.addValue(
					nodes,
					groupPath,
					"address-group",
					connection.Address,
				)
			}
		}

		if connection.Port != nil {
			port, err := strconv.Atoi(*connection.Port)
			if err == nil {
				definitions.addValue(nodes, path, "port", port)
			} else {
				// Group needs to nest further
				groupPath := path.Extend(utils.MakeVyosPC("group"))
				definitions.add(generateSparseDefinitionTree(nodes, groupPath))
				definitions.addValue(
					nodes,
					groupPath,
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
	path := utils.MakeVyosPath()
	path.Append(
		utils.MakeVyosPC("firewall"),
		utils.MakeVyosPC("group"),
		utils.MakeVyosPC("port-group"),
		utils.MakeVyosDynamicPC(group.Name),
	)
	groupDefinition := generateSparseDefinitionTree(nodes, path)
	definitions.add(groupDefinition)

	definitions.addValue(
		nodes, path, "description", group.Description,
	)
	definitions.addListValue(
		nodes, path, "port", config.SliceIntToAny(group.Ports),
	)

	return definitions
}
