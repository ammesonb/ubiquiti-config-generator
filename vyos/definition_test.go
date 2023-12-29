package vyos

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiffNode(t *testing.T) {
	def := Definition{
		Path:  []string{"foo"},
		Name:  "node.tag",
		Value: "node1",
		Node: &Node{
			Name:  "node1",
			Type:  "txt",
			IsTag: true,
			Multi: true,
			Path:  "foo",
		},
	}

	diffs := def.diffNode(&Definition{
		Path: []string{"bar"},
		Name: "node2",
		Node: &Node{
			Name:  "node2",
			Type:  "int",
			IsTag: false,
			Multi: false,
			Path:  "bar",
		},
	})

	assert.Len(t, diffs, 5, "Should have 5 differences, got %d", diffs)
}

func TestDiffDefinition(t *testing.T) {
	def := Definition{
		Name:    "node.tag",
		Path:    []string{"foo"},
		Comment: "foo",
		Value:   "node1",
		Values:  []any{80},
		Node:    nil,
	}

	diffs := def.diffDefinition(&Definition{
		Name:    "foo",
		Path:    []string{"bar"},
		Comment: "bar",
		Value:   "node2",
		Values:  []any{443},
		Node:    nil,
	})

	assert.Len(t, diffs, 7, "Should have 7 differences, got %d", len(diffs))
}

func TestMerge(t *testing.T) {
	nodes, err := GetGeneratedNodes()
	if err != nil {
		t.Errorf("Failed to get generated nodes: %v", err)
	}
	definitions := initDefinitions()

	definitions.add(&Definition{
		Name:     "interfaces",
		Path:     []string{},
		Node:     nodes.FindChild([]string{"interfaces"}),
		Comment:  "",
		Value:    nil,
		Values:   nil,
		Children: []*Definition{},
	})

	definitions.add(&Definition{
		Name:     "firewall",
		Path:     []string{},
		Node:     nodes.FindChild([]string{"firewall"}),
		Comment:  "",
		Value:    nil,
		Values:   nil,
		Children: []*Definition{},
	})

	definitions.add(&Definition{
		Name:     "all-ping",
		Path:     []string{"firewall"},
		Node:     nodes.FindChild([]string{"firewall", "all-ping"}),
		Comment:  "",
		Value:    "enable",
		Values:   nil,
		Children: []*Definition{},
	})
	definitions.add(&Definition{
		Name:     "broadcast-ping",
		Path:     []string{"firewall"},
		Node:     nodes.FindChild([]string{"firewall", "broadcast-ping"}),
		Comment:  "",
		Value:    "disable",
		Values:   nil,
		Children: []*Definition{},
	})
	definitions.add(&Definition{
		Name:     "group",
		Path:     []string{"firewall"},
		Node:     nodes.FindChild([]string{"firewall", "group"}),
		Comment:  "",
		Value:    nil,
		Values:   nil,
		Children: []*Definition{},
	})
	definitions.add(&Definition{
		Name:     "port-group",
		Path:     []string{"firewall", "group"},
		Node:     nodes.FindChild([]string{"firewall", "group", "port-group"}),
		Comment:  "",
		Value:    "test-port-group",
		Values:   nil,
		Children: []*Definition{},
	})
	definitions.add(&Definition{
		Name:     "description",
		Path:     []string{"firewall", "group", "port-group", "test-port-group"},
		Node:     nodes.FindChild([]string{"firewall", "group", "port-group", "node.tag", "description"}),
		Comment:  "",
		Value:    "a test port group",
		Values:   nil,
		Children: []*Definition{},
	})
	definitions.add(&Definition{
		Name:     "port",
		Path:     []string{"firewall", "group", "port-group", "test-port-group"},
		Node:     nodes.FindChild([]string{"firewall", "group", "port-group", "node.tag", "port"}),
		Comment:  "",
		Value:    nil,
		Values:   []any{53, 123},
		Children: []*Definition{},
	})

	others := initDefinitions()

	domainName := "ubiquiti"
	others.add(&Definition{
		Name:     "system",
		Path:     []string{},
		Node:     nodes.FindChild([]string{"system"}),
		Comment:  "",
		Value:    nil,
		Values:   nil,
		Children: []*Definition{},
	})
	others.add(&Definition{
		Name:     "domain-name",
		Path:     []string{"system"},
		Node:     nodes.FindChild([]string{"system", "domain-name"}),
		Comment:  "",
		Value:    domainName,
		Values:   nil,
		Children: []*Definition{},
	})

	others.add(&Definition{
		Name:     "interfaces",
		Path:     []string{},
		Node:     nodes.FindChild([]string{"interfaces"}),
		Comment:  "",
		Value:    nil,
		Values:   nil,
		Children: []*Definition{},
	})
	others.add(&Definition{
		Name:     "ethernet",
		Path:     []string{"interfaces"},
		Node:     nodes.FindChild([]string{"interfaces", "ethernet"}),
		Comment:  "",
		Value:    "eth0",
		Values:   nil,
		Children: []*Definition{},
	})
	others.add(&Definition{
		Name:     "address",
		Path:     []string{"interfaces", "ethernet", "eth0"},
		Node:     nodes.FindChild([]string{"interfaces", "ethernet", "node.tag", "address"}),
		Comment:  "",
		Value:    "192.168.0.1",
		Values:   nil,
		Children: []*Definition{},
	})
	others.add(&Definition{
		Name:     "vif",
		Path:     []string{"interfaces", "ethernet", "eth0"},
		Node:     nodes.FindChild([]string{"interfaces", "ethernet", "node.tag", "vif"}),
		Comment:  "",
		Value:    99,
		Values:   nil,
		Children: []*Definition{},
	})

	others.add(&Definition{
		Name:     "firewall",
		Path:     []string{},
		Node:     nodes.FindChild([]string{"firewall"}),
		Comment:  "",
		Value:    nil,
		Values:   nil,
		Children: []*Definition{},
	})
	others.add(&Definition{
		Name:     "all-ping",
		Path:     []string{"firewall"},
		Node:     nodes.FindChild([]string{"firewall", "all-ping"}),
		Comment:  "",
		Value:    "enable",
		Values:   nil,
		Children: []*Definition{},
	})
	others.add(&Definition{
		Name:     "ip-src-route",
		Path:     []string{"firewall"},
		Node:     nodes.FindChild([]string{"firewall", "ip-src-route"}),
		Comment:  "",
		Value:    "disable",
		Values:   nil,
		Children: []*Definition{},
	})
	others.add(&Definition{
		Name:     "group",
		Path:     []string{"firewall"},
		Node:     nodes.FindChild([]string{"firewall", "group"}),
		Comment:  "",
		Value:    nil,
		Values:   nil,
		Children: []*Definition{},
	})
	others.add(&Definition{
		Name:     "port-group",
		Path:     []string{"firewall", "group"},
		Node:     nodes.FindChild([]string{"firewall", "group", "port-group"}),
		Comment:  "",
		Value:    "test-port-group",
		Values:   nil,
		Children: []*Definition{},
	})
	others.add(&Definition{
		Name:     "port",
		Path:     []string{"firewall", "group", "port-group", "test-port-group"},
		Node:     nodes.FindChild([]string{"firewall", "group", "port-group", "node.tag", "port"}),
		Comment:  "",
		Value:    nil,
		Values:   []any{53, 123},
		Children: []*Definition{},
	})
	others.add(&Definition{
		Name:     "port-group",
		Path:     []string{"firewall", "group"},
		Node:     nodes.FindChild([]string{"firewall", "group", "port-group"}),
		Comment:  "",
		Value:    "http-port-group",
		Values:   nil,
		Children: []*Definition{},
	})
	others.add(&Definition{
		Name:     "description",
		Path:     []string{"firewall", "group", "port-group", "http-port-group"},
		Node:     nodes.FindChild([]string{"firewall", "group", "port-group", "node.tag", "description"}),
		Comment:  "",
		Value:    "HTTP ports",
		Values:   nil,
		Children: []*Definition{},
	})
	others.add(&Definition{
		Name:     "port",
		Path:     []string{"firewall", "group", "port-group", "http-port-group"},
		Node:     nodes.FindChild([]string{"firewall", "group", "port-group", "node.tag", "port"}),
		Comment:  "",
		Value:    nil,
		Values:   []any{80, 443},
		Children: []*Definition{},
	})

	err = definitions.merge(others)
	assert.NoError(t, err, "Definitions should merge successfully")

	assert.Equal(
		t,
		domainName,
		definitions.FindChild([]any{"system", "domain-name"}).Value,
		"Added system domain name should match",
	)

	assert.Equal(
		t,
		"enable",
		definitions.FindChild([]any{"firewall", "all-ping"}).Value,
		"All ping should stay enabled",
	)

	assert.Equal(
		t,
		"disable",
		definitions.FindChild([]any{"firewall", "broadcast-ping"}).Value,
		"Broadcast ping should be disabled",
	)

	assert.Equal(
		t,
		"disable",
		definitions.FindChild([]any{"firewall", "ip-src-route"}).Value,
		"IP source route should be disabled",
	)

	assert.Len(t, definitions.FindChild([]any{"firewall", "group"}).Children, 2, "Two port groups")

	httpPortGroup := definitions.FindChild([]any{"firewall", "group", "port-group", "http-port-group"})
	assert.NotNil(t, httpPortGroup, "HTTP port group should be added")

	if httpPortGroup == nil {
		t.FailNow()
	}

	assert.Equal(
		t,
		"HTTP ports",
		definitions.FindChild([]any{"firewall", "group", "port-group", "http-port-group", "description"}).Value,
		"Expected correct HTTP port group description",
	)

	assert.Len(t, httpPortGroup.Children, 2, "Expected exactly 2 children: port and description")

	assert.Equal(
		t,
		[]any{80, 443},
		definitions.FindChild([]any{"firewall", "group", "port-group", "http-port-group", "port"}).Values,
		"Expected correct HTTP ports",
	)

	assert.Equal(
		t,
		"192.168.0.1",
		definitions.FindChild([]any{"interfaces", "ethernet", "eth0", "address"}).Value,
		"Expected correct ethernet address",
	)

	assert.Equal(
		t,
		99,
		definitions.FindChild([]any{"interfaces", "ethernet", "eth0", "vif", 99}).Value,
		"Expected correct ethernet VIF",
	)
}

func TestMergeConflictingDefinition(t *testing.T) {
	nodes, err := GetGeneratedNodes()
	if err != nil {
		t.Errorf("Failed to get generated nodes: %v", err)
	}
	definitions := initDefinitions()

	definitions.add(&Definition{
		Name:     "firewall",
		Path:     []string{},
		Node:     nodes.FindChild([]string{"firewall"}),
		Comment:  "",
		Value:    nil,
		Values:   nil,
		Children: []*Definition{},
	})

	definitions.add(&Definition{
		Name:     "all-ping",
		Path:     []string{"firewall"},
		Node:     nodes.FindChild([]string{"firewall", "all-ping"}),
		Comment:  "",
		Value:    "enable",
		Values:   nil,
		Children: []*Definition{},
	})
	definitions.add(&Definition{
		Name:     "broadcast-ping",
		Path:     []string{"firewall"},
		Node:     nodes.FindChild([]string{"firewall", "broadcast-ping"}),
		Comment:  "",
		Value:    "disable",
		Values:   nil,
		Children: []*Definition{},
	})
	definitions.add(&Definition{
		Name:     "group",
		Path:     []string{"firewall"},
		Node:     nodes.FindChild([]string{"firewall", "group"}),
		Comment:  "",
		Value:    nil,
		Values:   nil,
		Children: []*Definition{},
	})
	definitions.add(&Definition{
		Name:     "port-group",
		Path:     []string{"firewall", "group"},
		Node:     nodes.FindChild([]string{"firewall", "group", "port-group"}),
		Comment:  "",
		Value:    "test-port-group",
		Values:   nil,
		Children: []*Definition{},
	})
	definitions.add(&Definition{
		Name:     "description",
		Path:     []string{"firewall", "group", "port-group", "test-port-group"},
		Node:     nodes.FindChild([]string{"firewall", "group", "port-group", "node.tag", "description"}),
		Comment:  "",
		Value:    "a test port group",
		Values:   nil,
		Children: []*Definition{},
	})
	definitions.add(&Definition{
		Name:     "port",
		Path:     []string{"firewall", "group", "port-group", "test-port-group"},
		Node:     nodes.FindChild([]string{"firewall", "group", "port-group", "node.tag", "port"}),
		Comment:  "",
		Value:    nil,
		Values:   []any{53, 123},
		Children: []*Definition{},
	})

	others := initDefinitions()

	domainName := "ubiquiti"
	others.add(&Definition{
		Name:     "system",
		Path:     []string{},
		Node:     nodes.FindChild([]string{"system"}),
		Comment:  "",
		Value:    nil,
		Values:   nil,
		Children: []*Definition{},
	})
	others.add(&Definition{
		Name:     "domain-name",
		Path:     []string{"system"},
		Node:     nodes.FindChild([]string{"system", "domain-name"}),
		Comment:  "",
		Value:    domainName,
		Values:   nil,
		Children: []*Definition{},
	})

	others.add(&Definition{
		Name:     "firewall",
		Path:     []string{},
		Node:     nodes.FindChild([]string{"firewall"}),
		Comment:  "",
		Value:    nil,
		Values:   nil,
		Children: []*Definition{},
	})
	others.add(&Definition{
		Name:     "all-ping",
		Path:     []string{"firewall"},
		Node:     nodes.FindChild([]string{"firewall", "all-ping"}),
		Comment:  "",
		Value:    "disable",
		Values:   nil,
		Children: []*Definition{},
	})

	err = definitions.merge(others)
	assert.ErrorContainsf(t, err, "differences between nodes at path firewall/all-ping", "Expected all-ping to have conflicting value")
}
