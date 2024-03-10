package vyos

import (
	"fmt"
	"github.com/ammesonb/ubiquiti-config-generator/abstraction"
	"github.com/ammesonb/ubiquiti-config-generator/config"
	"github.com/ammesonb/ubiquiti-config-generator/utils"
	"github.com/ammesonb/ubiquiti-config-generator/validation"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestFromPortGroupAbstraction(t *testing.T) {
	nodes := GetGeneratedNodes(t)

	name := "port-group"
	description := "test empty port group"
	group := abstraction.PortGroup{
		Name:        name,
		Description: description,
		Ports:       []int{},
	}
	defs, err := FromPortGroupAbstraction(nodes, group)

	assert.NoError(t, err)
	groupPath := []any{"firewall", "group", "port-group", name}
	assert.NotNil(t, defs.FindChild(groupPath), "Port group should exist")
	assert.Equal(t, description, defs.FindChild(append(groupPath, "description")).Value)
	assert.Empty(t, defs.FindChild(append(groupPath, "port")).Values)

	description = "test port group"
	group = abstraction.PortGroup{
		Name:        name,
		Description: description,
		Ports:       []int{123, 53, 80},
	}
	defs, err = FromPortGroupAbstraction(nodes, group)

	assert.NoError(t, err)
	assert.NotNil(t, defs.FindChild(groupPath), "Port group should exist")
	assert.Equal(t, description, defs.FindChild(append(groupPath, "description")).Value)
	assert.Len(t, defs.FindChild(append(groupPath, "port")).Values, 3)
	assert.ElementsMatch(t, []int{123, 53, 80}, defs.FindChild(append(groupPath, "port")).Values)

	groupN := nodes.FindChild([]string{"firewall", "group"})
	portGroup := groupN.ChildNodes["port-group"]
	delete(groupN.ChildNodes, "port-group")
	defs, err = FromPortGroupAbstraction(nodes, group)
	assert.ErrorIs(t, err, utils.ErrWithCtx(errGenPortGroupTree, name))
	assert.ErrorIs(t, err, utils.ErrWithCtx(errNonexistentNode, "firewall/group/port-group"))
	groupN.ChildNodes["port-group"] = portGroup
}

func getSampleNetwork() abstraction.Network {
	vif := int32(10)

	dns := []string{"8.8.8.8", "8.8.4.4"}

	hostOneAddr := "10.0.0.10"
	dnsPort := "53"
	webPorts := "web-ports"
	group1 := "group1"
	group2 := "group2"

	ruleStart := 1000
	ruleStep := 10

	return abstraction.Network{
		Name:        "test-network",
		Description: "sample test network",
		Interface: &abstraction.Interface{
			Name:             "eth1",
			Description:      "first port",
			Address:          "10.0.0.1",
			Vif:              &vif,
			Speed:            "auto",
			Duplex:           "auto",
			InboundFirewall:  "test-inbound",
			OutboundFirewall: "test-outbound",
			LocalFirewall:    "test-local",
			Extra: map[string]any{
				"disable-link-detect": "on",
			},
		},
		Authoritative: "enable",
		Subnets: []*abstraction.Subnet{
			{
				CIDR:          "10.0.0.0/24",
				DefaultRouter: "10.0.0.1",
				DomainName:    "home.lan",
				DNSServers:    dns,
				DHCPLease:     int32(86400),
				DHCPStart:     "10.0.0.240",
				DHCPEnd:       "10.0.0.255",
				Extra: map[string]any{
					"time-server": []any{"10.0.0.1", "10.0.0.2"},
				},
				Hosts: []*abstraction.Host{
					{
						Name:          "host-1",
						Address:       hostOneAddr,
						MAC:           "ab:cd:ef:12:34:56",
						AddressGroups: []string{"group1", "group2"},
						ForwardPorts:  map[int32]int32{8080: 80, 4430: 443, 123: 123},
						Connections:   []abstraction.FirewallConnection{},
					},
					{
						Name:          "host-2",
						Address:       "10.0.0.12",
						MAC:           "12:34:56:ab:cd:ef",
						AddressGroups: []string{"group1"},
						ForwardPorts:  map[int32]int32{},
						Connections: []abstraction.FirewallConnection{
							{
								Description: "allow outbound web connections",
								Allow:       true,
								Protocol:    "tcp_udp",
								Log:         false,
								Source: &abstraction.ConnectionDetail{
									Port: &webPorts,
								},
								Destination: &abstraction.ConnectionDetail{
									Address: &group1,
								},
							},
							{
								Description: "block inbound web connections",
								Allow:       false,
								Protocol:    "tcp_udp",
								Log:         true,
								Destination: &abstraction.ConnectionDetail{
									Address: &group1,
									Port:    &webPorts,
								},
							},
							{
								Description: "allow group 1 and 2 to communicate",
								Allow:       true,
								Protocol:    "all",
								Log:         true,
								Source: &abstraction.ConnectionDetail{
									Address: &group2,
								},
								Destination: &abstraction.ConnectionDetail{
									Address: &group1,
								},
							},
							{
								Description: "allow group 2 and 1 to communicate",
								Allow:       true,
								Protocol:    "all",
								Log:         true,
								Source: &abstraction.ConnectionDetail{
									Address: &group1,
								},
								Destination: &abstraction.ConnectionDetail{
									Address: &group2,
								},
							},
							{
								Description: "allow host to talk to DNS",
								Allow:       true,
								Protocol:    "all",
								Log:         true,
								Source: &abstraction.ConnectionDetail{
									Address: &hostOneAddr,
								},
								Destination: &abstraction.ConnectionDetail{
									Port: &dnsPort,
								},
							},
						},
					},
				},
			},
			{
				CIDR:          "10.1.0.0/24",
				DefaultRouter: "10.0.0.1",
				DomainName:    "guest.lan",
				DNSServers:    dns,
				DHCPLease:     int32(86400),
				DHCPStart:     "10.1.0.100",
				DHCPEnd:       "10.1.0.255",
				Extra:         map[string]any{},
				Hosts:         []*abstraction.Host{},
			},
		},
		InboundInterface:        "eth0",
		FirewallRuleNumberStart: ruleStart,
		FirewallRuleNumberStep:  ruleStep,
	}
}

func TestFromNetworkAbstraction(t *testing.T) {
	nodes := GetGeneratedNodes(t)

	network := getSampleNetwork()

	for _, firewall := range []string{
		network.Interface.InboundFirewall,
		network.Interface.OutboundFirewall,
		network.Interface.LocalFirewall,
		abstraction.NAT_COUNTER,
	} {
		abstraction.MakeCounter(firewall, network.FirewallRuleNumberStart, network.FirewallRuleNumberStep)
	}

	generated, errs := FromNetworkAbstraction(nodes, &network)
	assert.NotNil(t, generated, "Definitions parsed")
	assert.Empty(t, errs, "No errors")

	// set up expected interface definitions
	ethPath := utils.MakeVyosPath()
	ethPath.Append(
		utils.MakeVyosPC("interfaces"),
		utils.MakeVyosPC("ethernet"),
		utils.MakeVyosDynamicPC("eth1"),
	)
	ifaceFwPath := ethPath.Extend(
		utils.MakeVyosPC("vif"),
		utils.MakeVyosDynamicPC("10"),
		utils.MakeVyosPC("firewall"),
	)
	expected := initDefinitions()
	assert.NoError(t, expected.ensureTree(
		nodes,
		ifaceFwPath,
	))

	eth := expected.FindChild(config.SliceStrToAny(ethPath.Path))
	assert.NotNil(t, eth)
	if eth != nil {
		expected.add(&Definition{
			Name:       "description",
			Path:       ethPath.Path,
			Node:       nodes.FindChild(utils.CopySliceWith(ethPath.NodePath, "description")),
			Value:      "CARRIER",
			ParentNode: nodes.FindChild(utils.AllExcept(ethPath.NodePath, 1)),
		})
		expected.add(&Definition{
			Name:       "duplex",
			Path:       ethPath.Path,
			Node:       nodes.FindChild(utils.CopySliceWith(ethPath.NodePath, "duplex")),
			Value:      "auto",
			ParentNode: nodes.FindChild(utils.AllExcept(ethPath.NodePath, 1)),
		})
		expected.add(&Definition{
			Name:       "speed",
			Path:       ethPath.Path,
			Node:       nodes.FindChild(utils.CopySliceWith(ethPath.NodePath, "speed")),
			Value:      "auto",
			ParentNode: nodes.FindChild(utils.AllExcept(ethPath.NodePath, 1)),
		})
	}

	// expected virtual interface definitions
	vifPath := ifaceFwPath.DivergeFrom(1)
	vif := expected.FindChild(config.SliceStrToAny(vifPath.Path))
	assert.NotNil(t, vif)
	if vif != nil {
		expected.add(&Definition{
			Name:       "description",
			Path:       vifPath.Path,
			Node:       nodes.FindChild(utils.CopySliceWith(vifPath.NodePath, "description")),
			Value:      "first port",
			ParentNode: nodes.FindChild(utils.AllExcept(vifPath.NodePath, 1)),
		})
		expected.add(&Definition{
			Name:       "address",
			Path:       vifPath.Path,
			Node:       nodes.FindChild(utils.CopySliceWith(vifPath.NodePath, "address")),
			Value:      "10.0.0.1",
			ParentNode: nodes.FindChild(utils.AllExcept(vifPath.NodePath, 1)),
		})
		expected.add(&Definition{
			Name:       "disable-link-detect",
			Path:       vifPath.Path,
			Node:       nodes.FindChild(utils.CopySliceWith(vifPath.NodePath, "disable-link-detect")),
			Value:      "on",
			ParentNode: nodes.FindChild(utils.AllExcept(vifPath.NodePath, 1)),
		})
		expected.add(&Definition{
			Name: "in",
			Path: utils.CopySliceWith(vifPath.Path, "firewall"),
			Node: nodes.FindChild(utils.CopySliceWith(vifPath.NodePath, "firewall", "in")),
			Children: []*Definition{
				{
					Name:       "name",
					Path:       utils.CopySliceWith(vifPath.Path, "firewall", "in"),
					Node:       nodes.FindChild(utils.CopySliceWith(vifPath.NodePath, "firewall", "in", "name")),
					Value:      "test-inbound",
					ParentNode: nodes.FindChild(utils.CopySliceWith(vifPath.NodePath, "firewall", "in")),
				},
			},
			ParentNode: nodes.FindChild(utils.CopySliceWith(vifPath.NodePath, "firewall")),
		})
		expected.add(&Definition{
			Name: "out",
			Path: utils.CopySliceWith(vifPath.Path, "firewall"),
			Node: nodes.FindChild(utils.CopySliceWith(vifPath.NodePath, "firewall", "out")),
			Children: []*Definition{
				{
					Name:       "name",
					Path:       utils.CopySliceWith(vifPath.Path, "firewall", "out"),
					Node:       nodes.FindChild(utils.CopySliceWith(vifPath.NodePath, "firewall", "out", "name")),
					Value:      "test-outbound",
					ParentNode: nodes.FindChild(utils.CopySliceWith(vifPath.NodePath, "firewall", "out")),
				},
			},
			ParentNode: nodes.FindChild(utils.CopySliceWith(vifPath.NodePath, "firewall")),
		})
		expected.add(&Definition{
			Name: "local",
			Path: utils.CopySliceWith(vifPath.Path, "firewall"),
			Node: nodes.FindChild(utils.CopySliceWith(vifPath.NodePath, "firewall", "local")),
			Children: []*Definition{
				{
					Name:       "name",
					Path:       utils.CopySliceWith(vifPath.Path, "firewall", "local"),
					Node:       nodes.FindChild(utils.CopySliceWith(vifPath.NodePath, "firewall", "local", "name")),
					Value:      "test-local",
					ParentNode: nodes.FindChild(utils.CopySliceWith(vifPath.NodePath, "firewall", "local")),
				},
			},
			ParentNode: nodes.FindChild(utils.CopySliceWith(vifPath.NodePath, "firewall")),
		})
	}

	diffs := expected.FindChild([]any{"interfaces"}).Diff(generated.FindChild([]any{"interfaces"}))
	assert.Len(t, diffs, 0, "No differences between generated and expected interface definitions")
	if len(diffs) > 0 {
		for _, d := range diffs {
			fmt.Println(d)
		}
	}

	// DHCP server definitions
	dhcpPath := utils.MakeVyosPath()
	dhcpPath.Append(
		utils.MakeVyosPC("service"),
		utils.MakeVyosPC("dhcp-server"),
		utils.MakeVyosPC("shared-network-name"),
		utils.MakeVyosDynamicPC("test-network"),
	)

	assert.NoError(t, expected.ensureTree(nodes, dhcpPath))

	expected.addValue(nodes, dhcpPath, "authoritative", "enable")
	expected.addValue(nodes, dhcpPath, "description", "sample test network")

	zeroSubnet := dhcpPath.Extend(
		utils.MakeVyosPC("subnet"),
		utils.MakeVyosDynamicPC("10.0.0.0/24"),
	)
	assert.NoError(t, expected.ensureTree(nodes, zeroSubnet))
	expected.addValue(nodes, zeroSubnet, "domain-name", "home.lan")
	expected.addValue(nodes, zeroSubnet, "default-router", "10.0.0.1")
	expected.addValue(nodes, zeroSubnet, "lease", int32(86400))
	expected.addListValue(nodes, zeroSubnet, "time-server", []any{"10.0.0.1", "10.0.0.2"})
	expected.addListValue(nodes, zeroSubnet, "dns-server", []any{"8.8.8.8", "8.8.4.4"})
	zeroStartPath := zeroSubnet.Extend(
		utils.MakeVyosPC("start"),
		utils.MakeVyosDynamicPC("10.0.0.240"),
	)
	assert.NoError(t, expected.ensureTree(nodes, zeroStartPath))
	expected.addValue(nodes, zeroStartPath, "stop", "10.0.0.255")

	zeroHostPath := zeroSubnet.Extend(utils.MakeVyosPC("static-mapping"))
	zeroHostOne := zeroHostPath.Extend(utils.MakeVyosDynamicPC("host-1"))
	zeroHostTwo := zeroHostPath.Extend(utils.MakeVyosDynamicPC("host-2"))
	assert.NoError(t, expected.ensureTree(nodes, zeroHostOne))
	assert.NoError(t, expected.ensureTree(nodes, zeroHostTwo))

	expected.addValue(nodes, zeroHostOne, "ip-address", "10.0.0.10")
	expected.addValue(nodes, zeroHostOne, "mac-address", "ab:cd:ef:12:34:56")
	expected.addValue(nodes, zeroHostTwo, "ip-address", "10.0.0.12")
	expected.addValue(nodes, zeroHostTwo, "mac-address", "12:34:56:ab:cd:ef")

	oneSubnet := zeroSubnet.DivergeFrom(
		1,
		utils.MakeVyosDynamicPC("10.1.0.0/24"),
	)
	assert.NoError(t, expected.ensureTree(nodes, oneSubnet))
	expected.addValue(nodes, oneSubnet, "domain-name", "guest.lan")
	expected.addValue(nodes, oneSubnet, "default-router", "10.0.0.1")
	expected.addValue(nodes, oneSubnet, "lease", int32(86400))
	expected.addListValue(nodes, oneSubnet, "dns-server", []any{"8.8.8.8", "8.8.4.4"})

	oneStartPath := oneSubnet.Extend(utils.MakeVyosPC("start"), utils.MakeVyosDynamicPC("10.1.0.100"))
	assert.NoError(t, expected.ensureTree(nodes, oneStartPath))
	expected.addValue(nodes, oneStartPath, "stop", "10.1.0.255")

	nat := utils.MakeVyosPath()
	nat.Append(
		utils.MakeVyosPC("service"),
		utils.MakeVyosPC("nat"),
		utils.MakeVyosPC("rule"),
	)
	rule := 1000
	// ensure ordering, just like in code
	forwards := map[int32]int32{8080: 80, 4430: 443, 123: 123}
	for _, src := range []int32{123, 4430, 8080} {
		dest := forwards[src]
		natRule := nat.Extend(utils.MakeVyosDynamicPC(strconv.Itoa(rule)))
		assert.NoError(t, expected.ensureTree(nodes, natRule))
		expected.addValue(
			nodes,
			natRule,
			"description",
			fmt.Sprintf("Forward from port %d to host-1 port %d", src, dest),
		)
		expected.addValue(nodes, natRule, "inbound-interface", "eth0")
		expected.addValue(nodes, natRule, "protocol", "tcp_udp")
		expected.addValue(nodes, natRule, "type", "destination")
		inside := natRule.Extend(utils.MakeVyosPC("inside-address"))
		assert.NoError(t, expected.ensureTree(nodes, inside))
		expected.addValue(nodes, inside, "address", "10.0.0.10")
		expected.addValue(nodes, inside, "port", dest)
		destination := natRule.Extend(utils.MakeVyosPC("destination"))
		assert.NoError(t, expected.ensureTree(nodes, destination))
		expected.addValue(nodes, destination, "port", src)
		rule += 10
	}

	diffs = expected.FindChild([]any{"service"}).Diff(generated.FindChild([]any{"service"}))
	assert.Len(t, diffs, 0, "No differences between generated and expected service definitions")
	if len(diffs) > 0 {
		for _, d := range diffs {
			fmt.Println(d)
		}
	}

	inFirewall := utils.MakeVyosPath()
	inFirewall.Append(
		utils.MakeVyosPC("firewall"),
		utils.MakeVyosPC("name"),
		utils.MakeVyosDynamicPC("test-inbound"),
		utils.MakeVyosPC("rule"),
	)
	outFirewall := inFirewall.DivergeFrom(
		2,
		utils.MakeVyosDynamicPC("test-outbound"),
		utils.MakeVyosPC("rule"),
	)
	localFirewall := inFirewall.DivergeFrom(
		2,
		utils.MakeVyosDynamicPC("test-local"),
	)
	assert.NoError(t, expected.ensureTree(nodes, localFirewall))

	web := any("web-ports")
	group1 := "group1"
	group2 := "group2"
	addFirewallRule(
		t,
		expected, nodes,
		outFirewall, "1000",
		"allow outbound web connections",
		true, false,
		"tcp_udp",
		nil, &web,
		&group1, nil,
	)
	addFirewallRule(
		t,
		expected, nodes,
		outFirewall, "1010",
		"block inbound web connections",
		false, true,
		"tcp_udp",
		nil, nil,
		&group1, &web,
	)
	addFirewallRule(
		t,
		expected, nodes,
		outFirewall, "1020",
		"allow group 1 and 2 to communicate",
		true, true,
		"all",
		&group2, nil,
		&group1, nil,
	)
	addFirewallRule(
		t,
		expected, nodes,
		inFirewall, "1000",
		"allow group 2 and 1 to communicate",
		true, true,
		"all",
		&group1, nil,
		&group2, nil,
	)
	hostAddr := "10.0.0.10"
	dnsPort := any(53)
	addFirewallRule(
		t,
		expected, nodes,
		inFirewall, "1010",
		"allow host to talk to DNS",
		true, true,
		"all",
		&hostAddr, nil,
		nil, &dnsPort,
	)

	diffs = expected.FindChild([]any{"firewall"}).Diff(generated.FindChild([]any{"firewall"}))
	assert.Len(t, diffs, 0, "No differences between generated and expected firewall definitions")
	if len(diffs) > 0 {
		for _, d := range diffs {
			fmt.Println(d)
		}
	}
}

func TestFromNetworkAbstractionErrors(t *testing.T) {
	nodes := GetGeneratedNodes(t)

	network := getSampleNetwork()

	for _, firewall := range []string{
		network.Interface.InboundFirewall,
		network.Interface.OutboundFirewall,
		network.Interface.LocalFirewall,
		abstraction.NAT_COUNTER,
	} {
		abstraction.MakeCounter(firewall, network.FirewallRuleNumberStart, network.FirewallRuleNumberStep)
	}

	checkErrors := func(hasResult bool, expectedErrors [][]error) {
		generated, actualErrs := FromNetworkAbstraction(nodes, &network)
		if hasResult {
			assert.NotNil(t, generated)
		} else {
			assert.Nil(t, generated)
		}
		assert.Len(t, actualErrs, len(expectedErrors))
		for idx, errs := range expectedErrors {
			// Only test for errors we can check
			if idx > len(actualErrs) {
				break
			}

			for _, err := range errs {
				assert.ErrorIs(t, actualErrs[idx], err)
			}
		}
	}
	service := nodes.ChildNodes["service"]
	delete(nodes.ChildNodes, "service")
	checkErrors(false, [][]error{
		{
			utils.Err(errGenDHCPTree),
			utils.ErrWithCtx(errNonexistentNode, "service"),
		},
	})
	nodes.ChildNodes["service"] = service

	interfaces := nodes.ChildNodes["interfaces"]
	delete(nodes.ChildNodes, "interfaces")
	checkErrors(false, [][]error{
		{
			utils.Err(errConfigInterface),
			utils.ErrWithCtx(errGenInterfaceTree, "eth1"),
			utils.ErrWithCtx(errNonexistentNode, "interfaces"),
		},
	})
	nodes.ChildNodes["interfaces"] = interfaces

	ifaces := nodes.FindChild([]string{"interfaces", "ethernet", utils.DYNAMIC_NODE})
	vif := ifaces.ChildNodes["vif"]
	delete(ifaces.ChildNodes, "vif")
	checkErrors(false, [][]error{
		{
			utils.Err(errConfigInterface),
			utils.ErrWithVarCtx(errGenEthVifTree, "eth1", 10),
			utils.ErrWithCtx(errNonexistentNode, "interfaces/ethernet/"+utils.DYNAMIC_NODE+"/vif"),
		},
	})
	ifaces.ChildNodes["vif"] = vif

	fw := vif.ChildNodes[utils.DYNAMIC_NODE].ChildNodes["firewall"]
	for dir, err := range map[string]string{
		"in":    errGenInFwTree,
		"out":   errGenOutFwTree,
		"local": errGenLocalFwTree,
	} {
		node := fw.ChildNodes[dir]
		delete(fw.ChildNodes, dir)
		checkErrors(false, [][]error{
			{
				utils.Err(errConfigInterface),
				utils.Err(err),
			},
		})
		fw.ChildNodes[dir] = node
	}

	dhcp := service.FindChild([]string{"dhcp-server", "shared-network-name", utils.DYNAMIC_NODE})
	subnet := dhcp.ChildNodes["subnet"]
	delete(dhcp.ChildNodes, "subnet")
	checkErrors(true, [][]error{
		{
			utils.ErrWithCtx(errGenSubnetTree, "10.0.0.0/24"),
			utils.ErrWithCtx(
				errNonexistentNode,
				"service/dhcp-server/shared-network-name/"+utils.DYNAMIC_NODE+"/subnet",
			),
		},
		{
			utils.ErrWithCtx(errGenSubnetTree, "10.1.0.0/24"),
			utils.ErrWithCtx(
				errNonexistentNode,
				"service/dhcp-server/shared-network-name/"+utils.DYNAMIC_NODE+"/subnet",
			),
		},
	})
	dhcp.ChildNodes["subnet"] = subnet

	host := subnet.FindChild([]string{utils.DYNAMIC_NODE})
	static := host.ChildNodes["static-mapping"]
	delete(host.ChildNodes, "static-mapping")
	path := "service/dhcp-server/shared-network-name/" + utils.DYNAMIC_NODE + "/subnet/" + utils.DYNAMIC_NODE + "/static-mapping"
	checkErrors(true, [][]error{
		{
			utils.ErrWithCtx(errGenHostTree, "host-1"),
			utils.ErrWithCtx(
				errNonexistentNode,
				path,
			),
		},
		{
			utils.ErrWithCtx(errGenHostTree, "host-2"),
			utils.ErrWithCtx(
				errNonexistentNode,
				path,
			),
		},
	})
	host.ChildNodes["static-mapping"] = static
}

func addFirewallRule(
	t *testing.T,
	definitions *Definitions,
	nodes *Node,
	path *utils.VyosPath,
	rule string,
	description string,
	allow bool,
	log bool,
	protocol string,
	srcAddr *string,
	srcPort *any,
	destAddr *string,
	destPort *any,
) {
	t.Helper()
	r := path.Extend(utils.MakeVyosDynamicPC(rule))
	assert.NoError(t, definitions.ensureTree(nodes, r))
	definitions.addValue(nodes, r, "description", description)
	action := "accept"
	if !allow {
		action = "drop"
	}
	definitions.addValue(nodes, r, "action", action)
	doLog := "enable"
	if !log {
		doLog = "disable"
	}
	definitions.addValue(nodes, r, "log", doLog)
	definitions.addValue(nodes, r, "protocol", protocol)
	for _, conn := range []struct {
		Type    string
		Address *string
		Port    *any
	}{
		{Type: "source", Address: srcAddr, Port: srcPort},
		{Type: "destination", Address: destAddr, Port: destPort},
	} {
		if conn.Address == nil && conn.Port == nil {
			continue
		}
		fwConn := r.Extend(utils.MakeVyosPC(conn.Type))
		assert.NoError(t, definitions.ensureTree(nodes, fwConn))
		if conn.Address != nil {
			if validation.IsValidAddress(*conn.Address) {
				definitions.addValue(nodes, fwConn, "address", *conn.Address)
			} else {
				fwGroup := fwConn.Extend(utils.MakeVyosPC("group"))
				assert.NoError(t, definitions.ensureTree(nodes, fwGroup))
				definitions.addValue(nodes, fwGroup, "address-group", *conn.Address)
			}
		}
		if conn.Port != nil {
			if _, ok := (*conn.Port).(string); !ok {
				definitions.addValue(nodes, fwConn, "port", *conn.Port)
			} else {
				fwGroup := fwConn.Extend(utils.MakeVyosPC("group"))
				assert.NoError(t, definitions.ensureTree(nodes, fwGroup))
				definitions.addValue(nodes, fwGroup, "port-group", *conn.Port)
			}
		}

	}
}

func TestGetConnectionFirewall(t *testing.T) {
	inbound := "inbound"
	outbound := "outbound"
	local := "local"
	network := abstraction.Network{Interface: &abstraction.Interface{
		InboundFirewall:  inbound,
		OutboundFirewall: outbound,
		LocalFirewall:    local,
	}}
	subnet := abstraction.Subnet{CIDR: "10.0.0.0/24"}
	host := abstraction.Host{AddressGroups: []string{"address1", "address2"}}
	addr := "10.0.0.24"
	addr2 := "address2"

	conn := abstraction.FirewallConnection{}
	assert.Equal(t, "", getConnectionFirewall(&network, &subnet, &host, conn), "Unknown firewall")

	conn.Source = &abstraction.ConnectionDetail{Address: &addr}
	assert.Equal(t, inbound, getConnectionFirewall(&network, &subnet, &host, conn), "Inbound address")
	conn.Source.Address = &addr2
	assert.Equal(t, inbound, getConnectionFirewall(&network, &subnet, &host, conn), "Inbound address group")

	conn.Source = nil
	conn.Destination = &abstraction.ConnectionDetail{Address: &addr}
	assert.Equal(t, outbound, getConnectionFirewall(&network, &subnet, &host, conn), "Outbound address")
	conn.Destination.Address = &addr2
	assert.Equal(t, outbound, getConnectionFirewall(&network, &subnet, &host, conn), "Outbound address group")

	conn.Source = &abstraction.ConnectionDetail{Address: &addr}
	assert.Equal(t, local, getConnectionFirewall(&network, &subnet, &host, conn), "Local address")
	conn.Source.Address = &addr2
	assert.Equal(t, local, getConnectionFirewall(&network, &subnet, &host, conn), "Local address group")

}

func TestAddHostToAddressGroups(t *testing.T) {
	nodes := GetGeneratedNodes(t)
	definitions := initDefinitions()
	host := &abstraction.Host{}
	assert.Empty(t, addHostToAddressGroups(nodes, definitions, host))
	host.AddressGroups = []string{"address1", "address2"}

	group := nodes.FindChild([]string{"firewall", "group"})
	addrGroup := group.ChildNodes["address-group"]
	delete(group.ChildNodes, "address-group")

	errs := addHostToAddressGroups(nodes, definitions, host)
	assert.Len(t, errs, 2)
	assert.ErrorIs(t, errs[0], utils.ErrWithCtx(errGenAddrGroupTree, "address1"))
	assert.ErrorIs(t, errs[1], utils.ErrWithCtx(errGenAddrGroupTree, "address2"))

	group.ChildNodes["address-group"] = addrGroup
}

func TestAddFirewallRules(t *testing.T) {
	nodes := GetGeneratedNodes(t)
	definitions := initDefinitions()
	network := abstraction.Network{Name: "test-network"}
	subnet := abstraction.Subnet{CIDR: "10.0.0.0/24"}
	addr := "10.0.0.10"
	host := abstraction.Host{
		Name:         "a-host",
		Address:      addr,
		ForwardPorts: map[int32]int32{80: 80, 443: 443},
		Connections:  make([]abstraction.FirewallConnection, 1),
	}
	errs := addFirewallRules(nodes, definitions, &network, &subnet, &host)
	assert.Len(t, errs, 1)
	assert.ErrorIs(t, errs[0], utils.ErrWithVarCtx(errFwRequiresInterface, "test-network", "a-host"))

	nat := nodes.FindChild([]string{"service", "nat"})
	rule := nat.ChildNodes["rule"]
	delete(nat.ChildNodes, "rule")

	errs = addFirewallRules(nodes, definitions, &network, &subnet, &host)
	assert.Len(t, errs, 3)
	assert.ErrorIs(t, errs[0], utils.ErrWithVarCtx(errGenNatRuleTree, 80, "a-host"))
	assert.ErrorIs(t, errs[1], utils.ErrWithVarCtx(errGenNatRuleTree, 443, "a-host"))
	assert.ErrorIs(t, errs[2], utils.ErrWithVarCtx(errFwRequiresInterface, "test-network", "a-host"))
	nat.ChildNodes["rule"] = rule

	network.Interface = &abstraction.Interface{
		InboundFirewall:  "inbound",
		OutboundFirewall: "outbound",
		LocalFirewall:    "local",
	}
	abstraction.MakeCounter("inbound", 100, 10)
	abstraction.MakeCounter("outbound", 100, 10)
	abstraction.MakeCounter("local", 100, 10)

	port1 := "80"
	port2 := "443"
	host.Connections = []abstraction.FirewallConnection{
		{Description: "port 80", Source: &abstraction.ConnectionDetail{Port: &port1}},
	}
	errs = addFirewallRules(nodes, definitions, &network, &subnet, &host)
	assert.Len(t, errs, 1)
	assert.ErrorIs(t, errs[0], utils.ErrWithVarCtx(errUnknownFirewall, "port 80", "a-host"))

	fw := nodes.FindChild([]string{"firewall", "name", utils.DYNAMIC_NODE})
	rule = fw.ChildNodes["rule"]
	delete(fw.ChildNodes, "rule")
	host.Connections = []abstraction.FirewallConnection{
		{Description: "port 80", Source: &abstraction.ConnectionDetail{Address: &addr, Port: &port1}},
		{Description: "port 443", Source: &abstraction.ConnectionDetail{Address: &addr, Port: &port2}},
	}
	errs = addFirewallRules(nodes, definitions, &network, &subnet, &host)
	assert.Len(t, errs, 2)
	assert.ErrorIs(t, errs[0], utils.ErrWithVarCtx(errGenFwRuleTree, "inbound", 100))
	assert.ErrorIs(t, errs[1], utils.ErrWithVarCtx(errGenFwRuleTree, "inbound", 110))

	fw.ChildNodes["rule"] = rule
}

func TestAddForwardPortErrors(t *testing.T) {
	nodes := GetGeneratedNodes(t)
	definitions := initDefinitions()
	host := abstraction.Host{
		Name:         "a-host",
		ForwardPorts: map[int32]int32{80: 80, 443: 443},
		Connections:  make([]abstraction.FirewallConnection, 1),
	}

	nat := nodes.FindChild([]string{"service", "nat"})
	rule := nat.ChildNodes["rule"]
	delete(nat.ChildNodes, "rule")

	err := addForwardPort(nodes, definitions, &host, "inbound", 80, 80)
	assert.ErrorIs(t, err, utils.ErrWithVarCtx(errGenNatRuleTree, 80, "a-host"))

	nat.ChildNodes["rule"] = rule

	ruleEntry := rule.ChildNodes[utils.DYNAMIC_NODE]
	dest := ruleEntry.ChildNodes["destination"]
	delete(ruleEntry.ChildNodes, "destination")

	err = addForwardPort(nodes, definitions, &host, "inbound", 80, 443)
	assert.ErrorIs(t, err, utils.ErrWithVarCtx(errGenDestinationNatTree, 80, "a-host"))

	ruleEntry.ChildNodes["destination"] = dest

	addr := ruleEntry.ChildNodes["inside-address"]
	delete(ruleEntry.ChildNodes, "inside-address")

	err = addForwardPort(nodes, definitions, &host, "inbound", 80, 443)
	assert.ErrorIs(t, err, utils.ErrWithVarCtx(errGenInsideNatTree, 80, "a-host"))

	ruleEntry.ChildNodes["inside-address"] = addr
}

func TestAddConnectionErrors(t *testing.T) {
	nodes := GetGeneratedNodes(t)
	definitions := initDefinitions()
	abstraction.MakeCounter("inbound", 100, 10)

	fw := nodes.FindChild([]string{"firewall", "name", utils.DYNAMIC_NODE})
	rule := fw.ChildNodes["rule"]
	delete(fw.ChildNodes, "rule")

	conn := abstraction.FirewallConnection{
		Description: "test rule",
		Source:      &abstraction.ConnectionDetail{},
		Destination: &abstraction.ConnectionDetail{},
	}
	err := addConnection(nodes, definitions, "inbound", conn)
	assert.ErrorIs(t, err, utils.ErrWithVarCtx(errGenFwRuleTree, "inbound", 100))

	fw.ChildNodes["rule"] = rule

	ruleEntry := rule.ChildNodes[utils.DYNAMIC_NODE]
	src := ruleEntry.ChildNodes["source"]
	delete(ruleEntry.ChildNodes, "source")

	err = addConnection(nodes, definitions, "inbound", conn)
	assert.ErrorIs(t, err, utils.ErrWithVarCtx(errGenFwSrcTree, "inbound", 110, "test rule"))

	ruleEntry.ChildNodes["source"] = src

	dst := ruleEntry.ChildNodes["destination"]
	delete(ruleEntry.ChildNodes, "destination")

	err = addConnection(nodes, definitions, "inbound", conn)
	assert.ErrorIs(t, err, utils.ErrWithVarCtx(errGenFwDestTree, "inbound", 120, "test rule"))

	ruleEntry.ChildNodes["destination"] = dst

	fwGroup := "fw-group"
	conn.Source.Address = &fwGroup
	conn.Destination.Port = &fwGroup

	addrGroup := src.ChildNodes["group"]
	delete(src.ChildNodes, "group")

	err = addConnection(nodes, definitions, "inbound", conn)
	assert.ErrorIs(t, err, utils.ErrWithCtx(errGenFwAddrGroupTree, "fw-group"))
	assert.ErrorIs(t, err, utils.ErrWithVarCtx(errGenFwSrcTree, "inbound", 130, "test rule"))

	src.ChildNodes["group"] = addrGroup

	portGroup := dst.ChildNodes["group"]
	delete(dst.ChildNodes, "group")

	err = addConnection(nodes, definitions, "inbound", conn)
	assert.ErrorIs(t, err, utils.ErrWithCtx(errGenFwPortGroupTree, "fw-group"))
	assert.ErrorIs(t, err, utils.ErrWithVarCtx(errGenFwDestTree, "inbound", 140, "test rule"))

	dst.ChildNodes["group"] = portGroup
}
