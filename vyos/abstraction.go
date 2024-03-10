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

var (
	errGenPortGroupTree      = "failed to create tree for port group %s"
	errGenDHCPTree           = "failed to generate DHCP service tree"
	errConfigInterface       = "failed to configure network interface"
	errGenInterfaceTree      = "failed to generate ethernet interface %s tree"
	errGenEthVifTree         = "failed to generate ethernet VIF %s.%d tree"
	errGenSubnetTree         = "failed to generate subnet %s tree"
	errGenHostTree           = "failed to create tree for host %s"
	errGenInFwTree           = "failed to generate inbound firewall tree"
	errGenOutFwTree          = "failed to generate outbound firewall tree"
	errGenLocalFwTree        = "failed to generate local firewall tree"
	errGenAddrGroupTree      = "failed to generate tree for address group %s"
	errFwRequiresInterface   = "network %s must have interface defined in order to set firewall rules on host %s"
	errGenNatRuleTree        = "failed to create NAT rule to forward port %d to %s"
	errUnknownFirewall       = "could not identify firewall for connection '%s' on host '%s'"
	errGenDestinationNatTree = "failed to create destination NAT to forward port %d to %s"
	errGenInsideNatTree      = "failed to create inside NAT to forward port %d to %s"
	errGenFwRuleTree         = "failed to create tree for firewall %s, rule %d"
	errGenFwSrcTree          = "failed to create tree for firewall %s rule %d source details for %s"
	errGenFwDestTree         = "failed to create tree for firewall %s rule %d destination details for %s"
	errGenFwAddrGroupTree    = "failed to create address group firewall path for %s"
	errGenFwPortGroupTree    = "failed to create port group firewall path for %s"
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
	if err := definitions.ensureTree(nodes, dhcpPath); err != nil {
		return nil, []error{utils.ErrWithParent(errGenDHCPTree, err)}
	}

	definitions.addValue(nodes, dhcpPath, "authoritative", network.Authoritative)
	definitions.addValue(nodes, dhcpPath, "description", network.Description)

	if network.Interface != nil {
		if err := configureNetworkInterface(
			nodes,
			definitions,
			network.Interface,
		); err != nil {
			return nil, []error{utils.ErrWithParent(errConfigInterface, err)}
		}
	}

	errors := make([]error, 0)
	subnetBase := dhcpPath.Extend(utils.MakeVyosPC("subnet"))
	for _, subnet := range network.Subnets {
		subnetPath := subnetBase.Extend(utils.MakeVyosDynamicPC(subnet.CIDR))
		if err := definitions.ensureTree(nodes, subnetPath); err != nil {
			errors = append(errors, utils.ErrWithCtxParent(errGenSubnetTree, subnet.CIDR, err))
			continue
		}
		addSubnetNetworkValues(nodes, definitions, subnet, subnetPath)

		for _, host := range subnet.Hosts {
			if err := addHostToSubnet(nodes, definitions, host, subnetPath); err != nil {
				errors = append(errors, err)
				continue
			}
			if errs := addHostToAddressGroups(nodes, definitions, host); errs != nil {
				errors = append(errors, errs...)
			}
			if err := addFirewallRules(nodes, definitions, network, subnet, host); err != nil {
				errors = append(errors, err...)
			}
		}
	}

	return definitions, errors
}

// configureNetworkInterface sets properties of an interface including address, firewalls, etc
func configureNetworkInterface(nodes *Node, definitions *Definitions, iface *abstraction.Interface) error {
	path := utils.MakeVyosPath()
	path.Append(
		utils.MakeVyosPC("interfaces"),
		utils.MakeVyosPC("ethernet"),
		utils.MakeVyosDynamicPC(iface.Name),
	)
	// initialize the interfaces/ethernet tree, and override the dynamic node with the proper value
	if err := definitions.ensureTree(nodes, path); err != nil {
		return utils.ErrWithCtxParent(errGenInterfaceTree, iface.Name, err)
	}

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
		if err := definitions.ensureTree(nodes, path); err != nil {
			return utils.ErrWithVarCtxParent(
				errGenEthVifTree,
				err,
				iface.Name,
				*iface.Vif,
			)
		}
	}

	definitions.addValue(nodes, path, "description", iface.Description)
	definitions.addValue(nodes, path, "address", iface.Address)

	addExtraValues(nodes, definitions, path, iface.Extra)

	inPath := path.Extend(
		utils.MakeVyosPC("firewall"),
		utils.MakeVyosPC("in"),
	)
	if err := definitions.ensureTree(nodes, inPath); err != nil {
		return utils.ErrWithParent(errGenInFwTree, err)
	}
	definitions.addValue(nodes, inPath, "name", iface.InboundFirewall)

	outPath := inPath.DivergeFrom(1, utils.MakeVyosPC("out"))
	if err := definitions.ensureTree(nodes, outPath); err != nil {
		return utils.ErrWithParent(errGenOutFwTree, err)
	}
	definitions.addValue(nodes, outPath, "name", iface.OutboundFirewall)

	localPath := inPath.DivergeFrom(1, utils.MakeVyosPC("local"))
	if err := definitions.ensureTree(nodes, localPath); err != nil {
		return utils.ErrWithParent(errGenLocalFwTree, err)
	}
	definitions.addValue(nodes, localPath, "name", iface.LocalFirewall)

	return nil
}

// addSubnetNetworkValues configures the given subnet definitions using the provided nodes and paths
func addSubnetNetworkValues(nodes *Node, definitions *Definitions, subnet *abstraction.Subnet, path *utils.VyosPath) {
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
func addHostToSubnet(nodes *Node, definitions *Definitions, host *abstraction.Host, path *utils.VyosPath) error {
	hostPath := path.Extend(
		utils.MakeVyosPC("static-mapping"),
		utils.MakeVyosDynamicPC(host.Name),
	)

	if err := definitions.ensureTree(nodes, hostPath); err != nil {
		return utils.ErrWithCtxParent(errGenHostTree, host.Name, err)
	}

	definitions.addValue(nodes, hostPath, "ip-address", host.Address)
	definitions.addValue(nodes, hostPath, "mac-address", host.MAC)

	return nil
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
func addHostToAddressGroups(nodes *Node, definitions *Definitions, host *abstraction.Host) []error {
	if host.AddressGroups == nil {
		return []error{}
	}

	path := utils.MakeVyosPath()
	path.Append(
		utils.MakeVyosPC("firewall"),
		utils.MakeVyosPC("group"),
		utils.MakeVyosPC("address-group"),
	)
	errors := make([]error, 0)
	for _, group := range host.AddressGroups {
		groupPath := path.Extend(utils.MakeVyosDynamicPC(group))
		if err := definitions.ensureTree(nodes, groupPath); err != nil {
			errors = append(errors, utils.ErrWithCtxParent(errGenAddrGroupTree, group, err))
			continue
		}
		definitions.appendToListValue(nodes, groupPath, "address", host.Address)
	}

	return errors
}

func addFirewallRules(nodes *Node, definitions *Definitions, network *abstraction.Network, subnet *abstraction.Subnet, host *abstraction.Host) []error {
	// ensure consistent ordering of ports
	fromPorts := make([]int, 0)
	for p := range host.ForwardPorts {
		fromPorts = append(fromPorts, int(p))
	}
	sort.Ints(fromPorts)

	errors := make([]error, 0)
	for _, fromPort := range fromPorts {
		from := int32(fromPort)
		if err := addForwardPort(nodes, definitions, host, network.InboundInterface, from, host.ForwardPorts[from]); err != nil {
			errors = append(errors, err)
		}
	}

	if len(host.Connections) > 0 && network.Interface == nil {
		return append(errors, utils.ErrWithVarCtx(errFwRequiresInterface, network.Name, host.Name))
	}

	for _, connection := range host.Connections {
		firewallName := getConnectionFirewall(network, subnet, host, connection)
		// if firewall name is blank, then some part of the rule definition is likely invalid since it does not seem to
		// apply to the host that it is written for
		if firewallName == "" {
			errors = append(
				errors,
				utils.ErrWithVarCtx(errUnknownFirewall, connection.Description, host.Name),
			)
			continue
		} else if err := addConnection(
			nodes,
			definitions,
			firewallName,
			connection,
		); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

func addForwardPort(nodes *Node, definitions *Definitions, host *abstraction.Host, inboundInterface string, from int32, to int32) error {
	ruleNumber := abstraction.GetCounter(abstraction.NAT_COUNTER).Next()

	path := utils.MakeVyosPath()
	path.Append(
		utils.MakeVyosPC("service"),
		utils.MakeVyosPC("nat"),
		utils.MakeVyosPC("rule"),
		utils.MakeVyosDynamicPC(strconv.Itoa(ruleNumber)),
	)

	if err := definitions.ensureTree(nodes, path); err != nil {
		return utils.ErrWithVarCtxParent(
			errGenNatRuleTree,
			err,
			from,
			host.Name,
		)
	}
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
	if err := definitions.ensureTree(nodes, destPath); err != nil {
		return utils.ErrWithVarCtx(
			errGenDestinationNatTree,
			from,
			host.Name,
		)
	}
	definitions.addValue(nodes, destPath, "port", from)

	insidePath := path.Extend(utils.MakeVyosPC("inside-address"))
	if err := definitions.ensureTree(nodes, insidePath); err != nil {
		return utils.ErrWithVarCtxParent(
			errGenInsideNatTree,
			err,
			from,
			host.Name,
		)
	}
	definitions.addValue(nodes, insidePath, "address", host.Address)
	definitions.addValue(nodes, insidePath, "port", to)

	return nil
}

func addConnection(
	nodes *Node,
	definitions *Definitions,
	firewall string,
	connection abstraction.FirewallConnection,
) error {
	rule := abstraction.GetCounter(firewall).Next()

	rulePath := utils.MakeVyosPath()
	rulePath.Append(
		utils.MakeVyosPC("firewall"),
		utils.MakeVyosPC("name"),
		utils.MakeVyosDynamicPC(firewall),
		utils.MakeVyosPC("rule"),
		utils.MakeVyosDynamicPC(strconv.Itoa(rule)),
	)

	if err := definitions.ensureTree(nodes, rulePath); err != nil {
		return utils.ErrWithVarCtxParent(errGenFwRuleTree, err, firewall, rule)
	}

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

	if err := addConnectionDetail(
		definitions,
		nodes,
		rulePath.Extend(utils.MakeVyosPC("source")),
		connection.Source,
	); err != nil {
		return utils.ErrWithVarCtxParent(errGenFwSrcTree, err, firewall, rule, connection.Description)
	}
	if err := addConnectionDetail(
		definitions,
		nodes,
		rulePath.Extend(utils.MakeVyosPC("destination")),
		connection.Destination,
	); err != nil {
		return utils.ErrWithVarCtxParent(errGenFwDestTree, err, firewall, rule, connection.Description)
	}

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

func addConnectionDetail(definitions *Definitions, nodes *Node, path *utils.VyosPath, connection *abstraction.ConnectionDetail) error {
	if connection != nil {
		if err := definitions.ensureTree(nodes, path); err != nil {
			return utils.ErrWithParent("failed to create connection tree", err)
		}
		if connection.Address != nil {
			if validation.IsValidAddress(*connection.Address) {
				// For valid IP address, just add it directly
				definitions.addValue(nodes, path, "address", *connection.Address)
			} else if connection.Address != nil {
				// Group needs to nest further
				groupPath := path.Extend(utils.MakeVyosPC("group"))
				if err := definitions.ensureTree(nodes, groupPath); err != nil {
					return utils.ErrWithCtxParent(errGenFwAddrGroupTree, *connection.Address, err)
				}
				definitions.addValue(
					nodes,
					groupPath,
					"address-group",
					*connection.Address,
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
				if err := definitions.ensureTree(nodes, groupPath); err != nil {
					return utils.ErrWithCtxParent(errGenFwPortGroupTree, *connection.Port, err)
				}
				definitions.addValue(
					nodes,
					groupPath,
					"port-group",
					*connection.Port,
				)
			}
		}
	}

	return nil
}

func FromPortGroupAbstraction(nodes *Node, group abstraction.PortGroup) (*Definitions, error) {
	// Have to explicitly add each port to a new slice, cannot convert/assert any other way :(
	definitions := initDefinitions()
	path := utils.MakeVyosPath()
	path.Append(
		utils.MakeVyosPC("firewall"),
		utils.MakeVyosPC("group"),
		utils.MakeVyosPC("port-group"),
		utils.MakeVyosDynamicPC(group.Name),
	)
	if err := definitions.ensureTree(nodes, path); err != nil {
		return nil, utils.ErrWithCtxParent(errGenPortGroupTree, group.Name, err)
	}

	definitions.addValue(
		nodes, path, "description", group.Description,
	)
	definitions.addListValue(
		nodes, path, "port", config.SliceIntToAny(group.Ports),
	)

	return definitions, nil
}
