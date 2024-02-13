package vyos

import (
	"fmt"
	"github.com/ammesonb/ubiquiti-config-generator/abstraction"
	"github.com/ammesonb/ubiquiti-config-generator/config"
	"github.com/ammesonb/ubiquiti-config-generator/utils"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestFromPortGroupAbstraction(t *testing.T) {
	nodes, err := GetGeneratedNodes()
	assert.NoError(t, err, "Node generation should succeed")

	name := "port-group"
	description := "test empty port group"
	group := abstraction.PortGroup{
		Name:        name,
		Description: description,
		Ports:       []int{},
	}
	defs := FromPortGroupAbstraction(nodes, group)

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
	defs = FromPortGroupAbstraction(nodes, group)

	assert.NotNil(t, defs.FindChild(groupPath), "Port group should exist")
	assert.Equal(t, description, defs.FindChild(append(groupPath, "description")).Value)
	assert.Len(t, defs.FindChild(append(groupPath, "port")).Values, 3)
	assert.ElementsMatch(t, []int{123, 53, 80}, defs.FindChild(append(groupPath, "port")).Values)
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
					"time-server": "10.0.0.1",
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
	nodes, err := GetGeneratedNodes()
	assert.NoError(t, err, "Node generation should succeed")

	network := getSampleNetwork()

	for _, firewall := range []string{
		network.Interface.InboundFirewall,
		network.Interface.OutboundFirewall,
		network.Interface.LocalFirewall,
		abstraction.NAT_COUNTER,
	} {
		abstraction.MakeCounter(firewall, network.FirewallRuleNumberStart, network.FirewallRuleNumberStart)
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
	expected.add(generateSparseDefinitionTree(
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

	expected.add(generateSparseDefinitionTree(nodes, dhcpPath))

	expected.addValue(nodes, dhcpPath, "authoritative", "enable")
	expected.addValue(nodes, dhcpPath, "description", "sample test network")

	zeroSubnet := dhcpPath.Extend(
		utils.MakeVyosPC("subnet"),
		utils.MakeVyosDynamicPC("10.0.0.0/24"),
	)
	expected.add(generateSparseDefinitionTree(nodes, zeroSubnet))
	expected.addValue(nodes, zeroSubnet, "domain-name", "home.lan")
	expected.addValue(nodes, zeroSubnet, "default-router", "10.0.0.1")
	expected.addValue(nodes, zeroSubnet, "lease", int32(86400))
	expected.addValue(nodes, zeroSubnet, "time-server", "10.0.0.1")
	expected.addListValue(nodes, zeroSubnet, "dns-server", []any{"8.8.8.8", "8.8.4.4"})
	zeroStartPath := zeroSubnet.Extend(
		utils.MakeVyosPC("start"),
		utils.MakeVyosDynamicPC("10.0.0.240"),
	)
	expected.add(
		generateSparseDefinitionTree(nodes, zeroStartPath),
	)
	expected.addValue(nodes, zeroStartPath, "stop", "10.0.0.255")

	zeroHostPath := zeroSubnet.Extend(utils.MakeVyosPC("static-mapping"))
	zeroHostOne := zeroHostPath.Extend(utils.MakeVyosDynamicPC("host-1"))
	zeroHostTwo := zeroHostPath.Extend(utils.MakeVyosDynamicPC("host-2"))
	expected.add(generateSparseDefinitionTree(nodes, zeroHostOne))
	expected.add(generateSparseDefinitionTree(nodes, zeroHostTwo))

	expected.addValue(nodes, zeroHostOne, "ip-address", "10.0.0.10")
	expected.addValue(nodes, zeroHostOne, "mac-address", "ab:cd:ef:12:34:56")
	expected.addValue(nodes, zeroHostTwo, "ip-address", "10.0.0.12")
	expected.addValue(nodes, zeroHostTwo, "mac-address", "12:34:56:ab:cd:ef")

	oneSubnet := zeroSubnet.DivergeFrom(
		1,
		utils.MakeVyosDynamicPC("10.1.0.0/24"),
	)
	expected.add(generateSparseDefinitionTree(nodes, oneSubnet))
	expected.addValue(nodes, oneSubnet, "domain-name", "guest.lan")
	expected.addValue(nodes, oneSubnet, "default-router", "10.0.0.1")
	expected.addValue(nodes, oneSubnet, "lease", int32(86400))
	expected.addValue(nodes, oneSubnet, "time-server", "10.0.0.1")
	expected.addListValue(nodes, oneSubnet, "dns-server", []any{"8.8.8.8", "8.8.4.4"})

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
		expected.add(generateSparseDefinitionTree(nodes, natRule))
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
		expected.add(generateSparseDefinitionTree(nodes, inside))
		expected.addValue(nodes, inside, "address", "10.0.0.10")
		expected.addValue(nodes, inside, "port", src)
		destination := natRule.Extend(utils.MakeVyosPC("destination"))
		expected.add(generateSparseDefinitionTree(nodes, destination))
		expected.addValue(nodes, destination, "port", dest)
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
	expected.add(generateSparseDefinitionTree(nodes, localFirewall))

	web := "web-ports"
	group1 := "group1"
	group2 := "group2"
	addFirewallRule(
		expected, nodes,
		outFirewall, "1000",
		"allow outbound web connections",
		true, false,
		"tcp_udp",
		nil, &web,
		&group1, nil,
	)
	addFirewallRule(
		expected, nodes,
		outFirewall, "1010",
		"block inbound web connections",
		false, true,
		"tcp_udp",
		nil, nil,
		&group1, &web,
	)
	addFirewallRule(
		expected, nodes,
		outFirewall, "1020",
		"allow group 1 and 2 to communicate",
		true, true,
		"all",
		&group2, nil,
		&group1, nil,
	)
	addFirewallRule(
		expected, nodes,
		inFirewall, "1000",
		"allow group 2 and 1 to communicate",
		true, true,
		"all",
		&group1, nil,
		&group2, nil,
	)
	hostAddr := "10.0.0.10"
	dnsPort := "53"
	addFirewallRule(
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

func addFirewallRule(
	definitions *Definitions,
	nodes *Node,
	path *utils.VyosPath,
	rule string,
	description string,
	allow bool,
	log bool,
	protocol string,
	srcAddr *string,
	srcPort *string,
	destAddr *string,
	destPort *string,
) {
	r := path.Extend(utils.MakeVyosDynamicPC(rule))
	definitions.add(generateSparseDefinitionTree(nodes, r))
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
	if srcAddr != nil || srcPort != nil {
		src := r.Extend(utils.MakeVyosPC("source"))
		definitions.add(generateSparseDefinitionTree(nodes, src))
		if srcAddr != nil {
			definitions.addValue(nodes, src, "address", *srcAddr)
		}
		if srcPort != nil {
			definitions.addValue(nodes, src, "port", *srcPort)
		}
	}
	if destAddr != nil || destPort != nil {
		dest := r.Extend(utils.MakeVyosPC("destination"))
		definitions.add(generateSparseDefinitionTree(nodes, dest))
		if destAddr != nil {
			definitions.addValue(nodes, dest, "address", *destAddr)
		}
		if destPort != nil {
			definitions.addValue(nodes, dest, "port", *destPort)
		}
	}
}
