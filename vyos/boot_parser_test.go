package vyos

import (
	"os"
	"testing"
)

func TestLineDetection(t *testing.T) {
	comment := "   /* This is a comment */"
	tagNodeOpen := "    name some-firewall {"
	tagNodeClose := "    }"
	normalNodeOpen := "services {"
	value := "    port 8080"
	quotedValue := `    description "Firewall rule {20}"`

	t.Run("Comment", func(t *testing.T) {
		verifyLine(t, comment, true, false, false, false)
	})

	t.Run("Open Tag Node", func(t *testing.T) {
		verifyLine(t, tagNodeOpen, false, true, false, false)

		description, name := splitNameFromValue(tagNodeOpen)
		if description != "name" {
			t.Errorf(
				"Incorrect description for tag node open: '%s'", description,
			)
		}
		if name != "some-firewall" {
			t.Errorf(
				"Incorrect node name for tag node open: '%s'", name,
			)
		}
	})

	t.Run("Close Tag Node", func(t *testing.T) {
		verifyLine(t, tagNodeClose, false, false, true, false)
	})

	t.Run("Open Normal Node", func(t *testing.T) {
		verifyLine(t, normalNodeOpen, false, true, false, false)
	})

	t.Run("Value", func(t *testing.T) {
		verifyLine(t, value, false, false, false, true)

		nodeName, nodeValue := splitNameFromValue(value)
		if nodeName != "port" {
			t.Errorf(
				"Incorrect name for node: '%s'", nodeName,
			)
		}
		if nodeValue != "8080" {
			t.Errorf(
				"Incorrect value for node: '%s'", nodeValue,
			)
		}
	})

	t.Run("Quoted value", func(t *testing.T) {
		verifyLine(t, quotedValue, false, false, false, true)

	})
}

func verifyLine(t *testing.T, line string, isComment bool, isNode bool, closesNode bool, hasValue bool) {
	if lineIsComment(line) != isComment {
		negate := ""
		if !isComment {
			negate = "not "
		}
		t.Errorf("Line should be %sdetected as comment", negate)
	}
	if lineCreatesNode(line) != isNode {
		negate := ""
		if !isNode {
			negate = "not "
		}
		t.Errorf("Line should %screate a node", negate)
	}
	if lineEndsNode(line) != closesNode {
		negate := ""
		if !closesNode {
			negate = "not "
		}
		t.Errorf("Line should %sclose a node", negate)
	}
	if lineHasValue(line) != hasValue {
		negate := ""
		if !closesNode {
			negate = "not "
		}
		t.Errorf("Line should %shave a value", negate)
	}
}

func TestSplitName(t *testing.T) {
	tagNodeOpen := "    name some-firewall {"
	quotedValue := `    description "Firewall rule {20}"`

	t.Run("Open Tag Node", func(t *testing.T) {
		description, name := splitNameFromValue(tagNodeOpen)
		if description != "name" {
			t.Errorf(
				"Incorrect description for tag node open: '%s'", description,
			)
		}
		if name != "some-firewall" {
			t.Errorf(
				"Incorrect node name for tag node open: '%s'", name,
			)
		}
	})

	t.Run("Quoted value", func(t *testing.T) {
		nodeName, nodeValue := splitNameFromValue(quotedValue)
		if nodeName != "description" {
			t.Errorf(
				"Incorrect name for quoted node: '%s'", nodeName,
			)
		}
		if nodeValue != "Firewall rule {20}" {
			t.Errorf(
				"Incorrect value for node: '%s'", nodeValue,
			)
		}
	})
}

func TestDefinitionName(t *testing.T) {
	name := getDefinitionName("    firewall {")
	if name != "firewall" {
		t.Errorf("Got incorrect name '%s', expected 'firewall'", name)
	}

	name = getDefinitionName("    name some-firewall {")
	if name != "name" {
		t.Errorf("Got incorrect tag attribute '%s', expected 'name'", name)
	}
}

func TestGetTagDefinitionName(t *testing.T) {
	name := getTagDefinitionName("    name some-firewall {")
	if name != "some-firewall" {
		t.Errorf("Got incorrect tag name '%s', expected 'some-firewall'", name)
	}
}

/*
	Tests:

- Multiple root nodes like firewall + service
- Multi node
- Tag nodes
- Comments
- Compound test combining above elements like a normal config would
*/
func TestParseBootDefinitions(t *testing.T) {
	rootNode, err := GetGeneratedNodes()
	if err != nil {
		t.Error(err.Error())
		t.FailNow()
	}

	firewall := Definition{
		Name: "firewall",
		Path: []string{},
		Node: rootNode.ChildNodes["firewall"],
		Children: []*Definition{
			{
				Name:     "all-ping",
				Path:     []string{"firewall"},
				Node:     rootNode.FindChild([]string{"firewall", "all-ping"}),
				Value:    "enable",
				Children: []*Definition{},
			},
			{
				Name:     "broadcast-ping",
				Path:     []string{"firewall"},
				Node:     rootNode.FindChild([]string{"firewall", "broadcast-ping"}),
				Value:    "disable",
				Children: []*Definition{},
			},
			{
				Name: "group",
				Path: []string{"firewall"},
				Node: rootNode.FindChild([]string{"firewall", "group"}),
				Children: []*Definition{
					{
						Name:    "address-group",
						Value:   "admin",
						Comment: "/* Reserved hosts for admin stuff */",
						Path:    []string{"firewall", "group"},
						Node: rootNode.FindChild([]string{
							"firewall", "group", "address-group",
						}),
						Children: []*Definition{
							{
								Name: "description",
								Path: []string{"firewall", "group", "address-group", "admin"},
								Node: rootNode.FindChild([]string{
									"firewall", "group", "address-group", "node.tag", "description",
								}),
								Value:    "admin",
								Children: []*Definition{},
							},
							{
								Name: "address",
								Path: []string{"firewall", "group", "address-group", "admin"},
								Node: rootNode.FindChild([]string{
									"firewall", "group", "address-group", "node.tag", "address",
								}),
								Values: []any{
									"192.168.0.1",
									"192.168.0.2",
									"192.168.1.1",
								},
								Children: []*Definition{},
							},
						},
					},
				},
			},
			{
				Name:     "log-martians",
				Path:     []string{"firewall"},
				Node:     rootNode.FindChild([]string{"firewall", "log-martians"}),
				Value:    "enable",
				Children: []*Definition{},
			},
			{
				Name:  "name",
				Path:  []string{"firewall"},
				Value: "WAN_IN",
				Node:  rootNode.FindChild([]string{"firewall", "name"}),
				Children: []*Definition{
					{
						Name:  "default-action",
						Value: "drop",
						Path:  []string{"firewall", "name", "WAN_IN"},
						Node: rootNode.FindChild([]string{
							"firewall", "name", "node.tag", "default-action",
						}),
						Children: []*Definition{},
					},
					{
						Name:  "rule",
						Value: "100",
						Path:  []string{"firewall", "name", "WAN_IN"},
						Node: rootNode.FindChild([]string{
							"firewall", "name", "node.tag", "rule",
						}),
						Children: []*Definition{
							{
								Name:  "action",
								Value: "accept",
								Path:  []string{"firewall", "name", "WAN_IN", "rule", "100"},
								Node: rootNode.FindChild([]string{
									"firewall", "name", "node.tag", "rule", "node.tag", "action",
								}),
								Children: []*Definition{},
							},
							{
								Name:  "description",
								Value: "Allow 'IGMP'",
								Path:  []string{"firewall", "name", "WAN_IN", "rule", "100"},
								Node: rootNode.FindChild([]string{
									"firewall", "name", "node.tag", "rule", "node.tag", "description",
								}),
								Children: []*Definition{},
							},
							{
								Name:  "log",
								Value: "disable",
								Path:  []string{"firewall", "name", "WAN_IN", "rule", "100"},
								Node: rootNode.FindChild([]string{
									"firewall", "name", "node.tag", "rule", "node.tag", "log",
								}),
								Children: []*Definition{},
							},
							{
								Name:  "protocol",
								Value: "igmp",
								Path:  []string{"firewall", "name", "WAN_IN", "rule", "100"},
								Node: rootNode.FindChild([]string{
									"firewall", "name", "node.tag", "rule", "node.tag", "protocol",
								}),
								Children: []*Definition{},
							},
						},
					},
				},
			},
		},
	}

	interfaces := Definition{
		Name: "interfaces",
		Path: []string{},
		Node: rootNode.ChildNodes["interfaces"],
		Children: []*Definition{
			{
				Name:  "ethernet",
				Path:  []string{"interfaces"},
				Node:  rootNode.FindChild([]string{"interfaces", "ethernet"}),
				Value: "eth0",
				Children: []*Definition{
					{
						Name: "address",
						Path: []string{"interfaces", "ethernet", "eth0"},
						Node: rootNode.FindChild([]string{
							"interfaces", "ethernet", "node.tag", "address",
						}),
						Values:   []any{"dhcp"},
						Children: []*Definition{},
					},
					{
						Name: "description",
						Path: []string{"interfaces", "ethernet", "eth0"},
						Node: rootNode.FindChild([]string{
							"interfaces", "ethernet", "node.tag", "description",
						}),
						Value:    "UPLINK",
						Children: []*Definition{},
					},
					{
						Name: "duplex",
						Path: []string{"interfaces", "ethernet", "eth0"},
						Node: rootNode.FindChild([]string{
							"interfaces", "ethernet", "node.tag", "duplex",
						}),
						Value:    "auto",
						Children: []*Definition{},
					},
					{
						Name: "firewall",
						Path: []string{"interfaces", "ethernet", "eth0"},
						Node: rootNode.FindChild([]string{
							"interfaces", "ethernet", "node.tag", "firewall",
						}),
						Children: []*Definition{
							{
								Name: "in",
								Path: []string{"interfaces", "ethernet", "eth0", "firewall"},
								Node: rootNode.FindChild([]string{
									"interfaces", "ethernet", "node.tag", "firewall", "in",
								}),
								Children: []*Definition{
									{
										Name: "name",
										Path: []string{"interfaces", "ethernet", "eth0", "firewall", "in"},
										Node: rootNode.FindChild([]string{
											"interfaces", "ethernet", "node.tag", "firewall", "in", "name",
										}),
										Value:    "WAN-IN",
										Children: []*Definition{},
									},
								},
							},
						},
					},
					{
						Name:     "speed",
						Path:     []string{"interfaces", "ethernet", "eth0"},
						Node:     rootNode.FindChild([]string{"interfaces", "ethernet", "node.tag", "speed"}),
						Value:    "auto",
						Children: []*Definition{},
					},
				},
			},
			{
				Name:  "ethernet",
				Path:  []string{"interfaces"},
				Node:  rootNode.FindChild([]string{"interfaces", "ethernet"}),
				Value: "eth1",
				Children: []*Definition{
					{
						Name: "address",
						Path: []string{"interfaces", "ethernet", "eth1"},
						Node: rootNode.FindChild([]string{
							"interfaces", "ethernet", "node.tag", "address",
						}),
						Values:   []any{"192.168.0.1/24"},
						Children: []*Definition{},
					},
					{
						Name: "description",
						Path: []string{"interfaces", "ethernet", "eth1"},
						Node: rootNode.FindChild([]string{
							"interfaces", "ethernet", "node.tag", "description",
						}),
						Value:    "HOUSE",
						Children: []*Definition{},
					},
					{
						Name: "duplex",
						Path: []string{"interfaces", "ethernet", "eth1"},
						Node: rootNode.FindChild([]string{
							"interfaces", "ethernet", "node.tag", "duplex",
						}),
						Value:    "auto",
						Children: []*Definition{},
					},
					{
						Name:     "speed",
						Path:     []string{"interfaces", "ethernet", "eth1"},
						Node:     rootNode.FindChild([]string{"interfaces", "ethernet", "node.tag", "speed"}),
						Value:    "auto",
						Children: []*Definition{},
					},
				},
			},
			{
				Name:     "loopback",
				Path:     []string{"interfaces"},
				Node:     rootNode.FindChild([]string{"interfaces", "loopback"}),
				Value:    "lo",
				Children: []*Definition{},
			},
			{
				Name:  "switch",
				Path:  []string{"interfaces"},
				Node:  rootNode.FindChild([]string{"interfaces", "switch"}),
				Value: "switch0",
				Children: []*Definition{
					{
						Name: "mtu",
						Path: []string{"interfaces", "switch", "switch0"},
						Node: rootNode.FindChild([]string{
							"interfaces", "switch", "node.tag", "mtu",
						}),
						Value:    "1500",
						Children: []*Definition{},
					},
				},
			},
		},
	}

	t.Run("Firewall", func(t *testing.T) {
		testFirewallBoot(t, rootNode, &firewall)
		testInterfacesBoot(t, rootNode, &interfaces)
	})
}

func testFirewallBoot(t *testing.T, rootNode *Node, expected *Definition) {
	file, err := os.Open("../vyos_test/firewall.boot")
	if err != nil {
		t.Errorf("Failed to read firewall boot data: %+v", err)
		t.FailNow()
	}

	definitions := initDefinitions()
	ParseBootDefinitions(file, definitions, rootNode)

	if len(definitions.Definitions) > 1 {
		t.Errorf("Got %d root definitions, expected one", len(definitions.Definitions))
		t.FailNow()
	}

	for _, mismatch := range expected.Diff(definitions.Definitions[0]) {
		t.Error(mismatch)
	}
}

func testInterfacesBoot(t *testing.T, rootNode *Node, expected *Definition) {
	file, err := os.Open("../vyos_test/interfaces.boot")
	if err != nil {
		t.Errorf("Failed to read InterfacES boot data: %+v", err)
		t.FailNow()
	}

	definitions := initDefinitions()
	ParseBootDefinitions(file, definitions, rootNode)

	if len(definitions.Definitions) > 1 {
		t.Errorf("Got %d root definitions, expected one", len(definitions.Definitions))
		t.FailNow()
	}

	for _, mismatch := range expected.Diff(definitions.Definitions[0]) {
		t.Error(mismatch)
	}
}
