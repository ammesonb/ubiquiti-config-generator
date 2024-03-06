package vyos

import (
	"github.com/ammesonb/ubiquiti-config-generator/utils"
	"os"
	"strings"
	"testing"
)

func TestLineDetection(t *testing.T) {
	comment := "   /* This is a comment */"
	tagNodeOpen := "    name some-firewall {"
	tagNodeClose := "    }"
	normalNodeOpen := "service {"
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
	rootNode := GetGeneratedNodes(t)

	firewall := Definition{
		Name:       "firewall",
		Path:       []string{},
		Node:       rootNode.ChildNodes["firewall"],
		ParentNode: nil,
		Children: []*Definition{
			{
				Name:       "all-ping",
				Path:       []string{"firewall"},
				Node:       rootNode.FindChild([]string{"firewall", "all-ping"}),
				ParentNode: rootNode.FindChild([]string{"firewall"}),
				Value:      "enable",
				Children:   []*Definition{},
			},
			{
				Name:       "broadcast-ping",
				Path:       []string{"firewall"},
				Node:       rootNode.FindChild([]string{"firewall", "broadcast-ping"}),
				ParentNode: rootNode.FindChild([]string{"firewall"}),
				Value:      "disable",
				Children:   []*Definition{},
			},
			{
				Name:       "group",
				Path:       []string{"firewall"},
				Node:       rootNode.FindChild([]string{"firewall", "group"}),
				ParentNode: rootNode.FindChild([]string{"firewall"}),
				Children: []*Definition{
					{
						Name:    "address-group",
						Value:   "admin",
						Comment: "/* Reserved hosts for admin stuff */",
						Path:    []string{"firewall", "group"},
						Node: rootNode.FindChild([]string{
							"firewall", "group", "address-group",
						}),
						ParentNode: rootNode.FindChild([]string{"firewall", "group"}),
						Children: []*Definition{
							{
								Name: "description",
								Path: []string{"firewall", "group", "address-group", "admin"},
								Node: rootNode.FindChild([]string{
									"firewall", "group", "address-group", utils.DYNAMIC_NODE, "description",
								}),
								ParentNode: rootNode.FindChild([]string{"firewall", "group", "address-group"}),
								Value:      "admin",
								Children:   []*Definition{},
							},
							{
								Name: "address",
								Path: []string{"firewall", "group", "address-group", "admin"},
								Node: rootNode.FindChild([]string{
									"firewall", "group", "address-group", utils.DYNAMIC_NODE, "address",
								}),
								ParentNode: rootNode.FindChild([]string{"firewall", "group", "address-group"}),
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
				Name:       "log-martians",
				Path:       []string{"firewall"},
				Node:       rootNode.FindChild([]string{"firewall", "log-martians"}),
				ParentNode: rootNode.FindChild([]string{"firewall"}),
				Value:      "enable",
				Children:   []*Definition{},
			},
			{
				Name:       "name",
				Path:       []string{"firewall"},
				Value:      "WAN_IN",
				Node:       rootNode.FindChild([]string{"firewall", "name"}),
				ParentNode: rootNode.FindChild([]string{"firewall"}),
				Children: []*Definition{
					{
						Name:  "default-action",
						Value: "drop",
						Path:  []string{"firewall", "name", "WAN_IN"},
						Node: rootNode.FindChild([]string{
							"firewall", "name", utils.DYNAMIC_NODE, "default-action",
						}),
						ParentNode: rootNode.FindChild([]string{"firewall", "name"}),
						Children:   []*Definition{},
					},
					{
						Name:  "rule",
						Value: "100",
						Path:  []string{"firewall", "name", "WAN_IN"},
						Node: rootNode.FindChild([]string{
							"firewall", "name", utils.DYNAMIC_NODE, "rule",
						}),
						ParentNode: rootNode.FindChild([]string{"firewall", "name"}),
						Children: []*Definition{
							{
								Name:  "action",
								Value: "accept",
								Path:  []string{"firewall", "name", "WAN_IN", "rule", "100"},
								Node: rootNode.FindChild([]string{
									"firewall", "name", utils.DYNAMIC_NODE, "rule", utils.DYNAMIC_NODE, "action",
								}),
								ParentNode: rootNode.FindChild([]string{"firewall", "name", utils.DYNAMIC_NODE, "rule"}),
								Children:   []*Definition{},
							},
							{
								Name:  "description",
								Value: "Allow 'IGMP'",
								Path:  []string{"firewall", "name", "WAN_IN", "rule", "100"},
								Node: rootNode.FindChild([]string{
									"firewall", "name", utils.DYNAMIC_NODE, "rule", utils.DYNAMIC_NODE, "description",
								}),
								ParentNode: rootNode.FindChild([]string{"firewall", "name", utils.DYNAMIC_NODE, "rule"}),
								Children:   []*Definition{},
							},
							{
								Name:  "log",
								Value: "disable",
								Path:  []string{"firewall", "name", "WAN_IN", "rule", "100"},
								Node: rootNode.FindChild([]string{
									"firewall", "name", utils.DYNAMIC_NODE, "rule", utils.DYNAMIC_NODE, "log",
								}),
								ParentNode: rootNode.FindChild([]string{"firewall", "name", utils.DYNAMIC_NODE, "rule"}),
								Children:   []*Definition{},
							},
							{
								Name:  "protocol",
								Value: "igmp",
								Path:  []string{"firewall", "name", "WAN_IN", "rule", "100"},
								Node: rootNode.FindChild([]string{
									"firewall", "name", utils.DYNAMIC_NODE, "rule", utils.DYNAMIC_NODE, "protocol",
								}),
								ParentNode: rootNode.FindChild([]string{"firewall", "name", utils.DYNAMIC_NODE, "rule"}),
								Children:   []*Definition{},
							},
						},
					},
				},
			},
		},
	}

	interfaces := Definition{
		Name:       "interfaces",
		Path:       []string{},
		Node:       rootNode.ChildNodes["interfaces"],
		ParentNode: nil,
		Children: []*Definition{
			{
				Name:       "ethernet",
				Path:       []string{"interfaces"},
				Node:       rootNode.FindChild([]string{"interfaces", "ethernet"}),
				ParentNode: rootNode.FindChild([]string{"interfaces"}),
				Value:      "eth0",
				Children: []*Definition{
					{
						Name: "address",
						Path: []string{"interfaces", "ethernet", "eth0"},
						Node: rootNode.FindChild([]string{
							"interfaces", "ethernet", utils.DYNAMIC_NODE, "address",
						}),
						ParentNode: rootNode.FindChild([]string{"interfaces", "ethernet"}),
						Values:     []any{"dhcp"},
						Children:   []*Definition{},
					},
					{
						Name: "description",
						Path: []string{"interfaces", "ethernet", "eth0"},
						Node: rootNode.FindChild([]string{
							"interfaces", "ethernet", utils.DYNAMIC_NODE, "description",
						}),
						ParentNode: rootNode.FindChild([]string{"interfaces", "ethernet"}),
						Value:      "UPLINK",
						Children:   []*Definition{},
					},
					{
						Name: "duplex",
						Path: []string{"interfaces", "ethernet", "eth0"},
						Node: rootNode.FindChild([]string{
							"interfaces", "ethernet", utils.DYNAMIC_NODE, "duplex",
						}),
						ParentNode: rootNode.FindChild([]string{"interfaces", "ethernet"}),
						Value:      "auto",
						Children:   []*Definition{},
					},
					{
						Name: "firewall",
						Path: []string{"interfaces", "ethernet", "eth0"},
						Node: rootNode.FindChild([]string{
							"interfaces", "ethernet", utils.DYNAMIC_NODE, "firewall",
						}),
						ParentNode: rootNode.FindChild([]string{"interfaces", "ethernet"}),
						Children: []*Definition{
							{
								Name: "in",
								Path: []string{"interfaces", "ethernet", "eth0", "firewall"},
								Node: rootNode.FindChild([]string{
									"interfaces", "ethernet", utils.DYNAMIC_NODE, "firewall", "in",
								}),
								ParentNode: rootNode.FindChild([]string{"interfaces", "ethernet", utils.DYNAMIC_NODE, "firewall"}),
								Children: []*Definition{
									{
										Name: "name",
										Path: []string{"interfaces", "ethernet", "eth0", "firewall", "in"},
										Node: rootNode.FindChild([]string{
											"interfaces", "ethernet", utils.DYNAMIC_NODE, "firewall", "in", "name",
										}),
										ParentNode: rootNode.FindChild([]string{"interfaces", "ethernet", utils.DYNAMIC_NODE, "firewall", "in"}),
										Value:      "WAN-IN",
										Children:   []*Definition{},
									},
								},
							},
						},
					},
					{
						Name:       "speed",
						Path:       []string{"interfaces", "ethernet", "eth0"},
						Node:       rootNode.FindChild([]string{"interfaces", "ethernet", utils.DYNAMIC_NODE, "speed"}),
						ParentNode: rootNode.FindChild([]string{"interfaces", "ethernet"}),
						Value:      "auto",
						Children:   []*Definition{},
					},
				},
			},
			{
				Name:       "ethernet",
				Path:       []string{"interfaces"},
				Node:       rootNode.FindChild([]string{"interfaces", "ethernet"}),
				ParentNode: rootNode.FindChild([]string{"interfaces"}),
				Value:      "eth1",
				Children: []*Definition{
					{
						Name: "address",
						Path: []string{"interfaces", "ethernet", "eth1"},
						Node: rootNode.FindChild([]string{
							"interfaces", "ethernet", utils.DYNAMIC_NODE, "address",
						}),
						ParentNode: rootNode.FindChild([]string{"interfaces", "ethernet"}),
						Values:     []any{"192.168.0.1/24"},
						Children:   []*Definition{},
					},
					{
						Name: "description",
						Path: []string{"interfaces", "ethernet", "eth1"},
						Node: rootNode.FindChild([]string{
							"interfaces", "ethernet", utils.DYNAMIC_NODE, "description",
						}),
						ParentNode: rootNode.FindChild([]string{"interfaces", "ethernet"}),
						Value:      "HOUSE",
						Children:   []*Definition{},
					},
					{
						Name: "duplex",
						Path: []string{"interfaces", "ethernet", "eth1"},
						Node: rootNode.FindChild([]string{
							"interfaces", "ethernet", utils.DYNAMIC_NODE, "duplex",
						}),
						ParentNode: rootNode.FindChild([]string{"interfaces", "ethernet"}),
						Value:      "auto",
						Children:   []*Definition{},
					},
					{
						Name:       "speed",
						Path:       []string{"interfaces", "ethernet", "eth1"},
						Node:       rootNode.FindChild([]string{"interfaces", "ethernet", utils.DYNAMIC_NODE, "speed"}),
						ParentNode: rootNode.FindChild([]string{"interfaces", "ethernet"}),
						Value:      "auto",
						Children:   []*Definition{},
					},
				},
			},
			{
				Name:       "loopback",
				Path:       []string{"interfaces"},
				Node:       rootNode.FindChild([]string{"interfaces", "loopback"}),
				ParentNode: rootNode.FindChild([]string{"interfaces"}),
				Value:      "lo",
				Children:   []*Definition{},
			},
			{
				Name:       "switch",
				Path:       []string{"interfaces"},
				Node:       rootNode.FindChild([]string{"interfaces", "switch"}),
				ParentNode: rootNode.FindChild([]string{"interfaces"}),
				Value:      "switch0",
				Children: []*Definition{
					{
						Name: "mtu",
						Path: []string{"interfaces", "switch", "switch0"},
						Node: rootNode.FindChild([]string{
							"interfaces", "switch", utils.DYNAMIC_NODE, "mtu",
						}),
						ParentNode: rootNode.FindChild([]string{"interfaces", "switch"}),
						Value:      "1500",
						Children:   []*Definition{},
					},
				},
			},
		},
	}

	serviceSSH := Definition{
		Name:       "service",
		Path:       []string{},
		Node:       rootNode.ChildNodes["service"],
		ParentNode: nil,
		Children: []*Definition{
			{
				Name:       "ssh",
				Path:       []string{"service"},
				Node:       rootNode.FindChild([]string{"service", "ssh"}),
				ParentNode: rootNode.FindChild([]string{"service"}),
				Children: []*Definition{
					{
						Name: "disable-password-authentication",
						Path: []string{"service", "ssh"},
						Node: rootNode.FindChild([]string{
							"service", "ssh", "disable-password-authentication",
						}),
						ParentNode: rootNode.FindChild([]string{"service", "ssh"}),
						Children:   []*Definition{},
					},
					{
						Name: "port",
						Path: []string{"service", "ssh"},
						Node: rootNode.FindChild([]string{
							"service", "ssh", "port",
						}),
						ParentNode: rootNode.FindChild([]string{"service", "ssh"}),
						Value:      "22",
						Children:   []*Definition{},
					},
					{
						Name: "protocol-version",
						Path: []string{"service", "ssh"},
						Node: rootNode.FindChild([]string{
							"service", "ssh", "protocol-version",
						}),
						ParentNode: rootNode.FindChild([]string{"service", "ssh"}),
						Value:      "v2",
						Children:   []*Definition{},
					},
				},
			},
		},
	}

	firewallBoot := "firewall.boot"
	interfacesBoot := "interfaces.boot"
	serviceSSHBoot := "service-ssh.boot"
	t.Run("Firewall", func(t *testing.T) {
		testSingleBoot(t, rootNode, firewallBoot, &firewall)
	})

	t.Run("Interfaces", func(t *testing.T) {
		testSingleBoot(t, rootNode, interfacesBoot, &interfaces)
	})

	t.Run("serviceSSH", func(t *testing.T) {
		testSingleBoot(t, rootNode, serviceSSHBoot, &serviceSSH)
	})

	t.Run("Combined", func(t *testing.T) {
		testCombinedBoot(
			t,
			rootNode,
			[]string{firewallBoot, interfacesBoot, serviceSSHBoot},
			[]*Definition{&firewall, &interfaces, &serviceSSH},
		)
	})
}

func testSingleBoot(t *testing.T, rootNode *Node, filename string, expected *Definition) {
	file, err := os.Open("../vyos_test/" + filename)
	if err != nil {
		t.Errorf("Failed to read boot data for %s: %+v", filename, err)
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

func testCombinedBoot(t *testing.T, rootNode *Node, files []string, expected []*Definition) {
	data := ""
	for _, name := range files {
		bootData, err := os.ReadFile("../vyos_test/" + name)
		if err != nil {
			t.Errorf("Failed to read boot data for %s: %+v", name, err)
			t.FailNow()
		}
		data += string(bootData)
		data += "\n"
	}

	reader := strings.NewReader(data)

	definitions := initDefinitions()
	ParseBootDefinitions(reader, definitions, rootNode)

	if len(expected) != len(definitions.Definitions) {
		t.Errorf(
			"Expected %d definitions but found %d",
			len(expected),
			len(definitions.Definitions),
		)
	}
	for idx, def := range expected {
		for _, mismatch := range def.Diff(definitions.Definitions[idx]) {
			t.Error(mismatch)
		}
	}
}
