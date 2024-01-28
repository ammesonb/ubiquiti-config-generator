package vyos

import (
	"github.com/ammesonb/ubiquiti-config-generator/config"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestDefinition_Diff(t *testing.T) {
	def := Definition{
		Path:  []string{},
		Name:  DYNAMIC_NODE,
		Value: "foo",
		Node: &Node{
			Name:  "node1",
			Type:  "txt",
			IsTag: true,
			Multi: true,
			Path:  "",
		},
		Children: []*Definition{
			{
				Name: "bar",
				Path: []string{"foo"},
			},
		},
	}

	diffs := def.Diff(&Definition{
		Path:  []string{},
		Name:  DYNAMIC_NODE,
		Value: "bar",
		Node: &Node{
			Name:  "node1",
			Type:  "txt",
			IsTag: true,
			Multi: true,
			Path:  "",
		},
		Children: []*Definition{},
	})

	assert.Len(t, diffs, 2, "Value mismatch and children mismatch")
	assert.Contains(t, diffs[0], "'Value' should be")
	assert.Contains(t, diffs[1], "Should have 1 children but got 0")
}

func TestDiffNode(t *testing.T) {
	def := Definition{
		Path:  []string{"foo"},
		Name:  DYNAMIC_NODE,
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
		Name:    DYNAMIC_NODE,
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
		Node:     nodes.FindChild([]string{"firewall", "group", "port-group", DYNAMIC_NODE, "description"}),
		Comment:  "",
		Value:    "a test port group",
		Values:   nil,
		Children: []*Definition{},
	})
	definitions.add(&Definition{
		Name:     "port",
		Path:     []string{"firewall", "group", "port-group", "test-port-group"},
		Node:     nodes.FindChild([]string{"firewall", "group", "port-group", DYNAMIC_NODE, "port"}),
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
		Node:     nodes.FindChild([]string{"interfaces", "ethernet", DYNAMIC_NODE, "address"}),
		Comment:  "",
		Value:    "192.168.0.1",
		Values:   nil,
		Children: []*Definition{},
	})
	others.add(&Definition{
		Name:     "vif",
		Path:     []string{"interfaces", "ethernet", "eth0"},
		Node:     nodes.FindChild([]string{"interfaces", "ethernet", DYNAMIC_NODE, "vif"}),
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
		Node:     nodes.FindChild([]string{"firewall", "group", "port-group", DYNAMIC_NODE, "port"}),
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
		Node:     nodes.FindChild([]string{"firewall", "group", "port-group", DYNAMIC_NODE, "description"}),
		Comment:  "",
		Value:    "HTTP ports",
		Values:   nil,
		Children: []*Definition{},
	})
	others.add(&Definition{
		Name:     "port",
		Path:     []string{"firewall", "group", "port-group", "http-port-group"},
		Node:     nodes.FindChild([]string{"firewall", "group", "port-group", DYNAMIC_NODE, "port"}),
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
		Node:     nodes.FindChild([]string{"firewall", "group", "port-group", DYNAMIC_NODE, "description"}),
		Comment:  "",
		Value:    "a test port group",
		Values:   nil,
		Children: []*Definition{},
	})
	definitions.add(&Definition{
		Name:     "port",
		Path:     []string{"firewall", "group", "port-group", "test-port-group"},
		Node:     nodes.FindChild([]string{"firewall", "group", "port-group", DYNAMIC_NODE, "port"}),
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

func TestGenerateDefinitionTree(t *testing.T) {
	nodes, err := GetGeneratedNodes()
	if err != nil {
		t.Errorf("Failed to generate fixtures: %v", err)
	}

	generated := generateSparseDefinitionTree(nodes, []string{"firewall", "group", "port-group"})

	expected := &Definition{
		Name: "firewall",
		Path: []string{},
		Node: nodes.FindChild([]string{"firewall"}),
		Children: []*Definition{
			{
				Name: "group",
				Path: []string{"firewall"},
				Node: nodes.FindChild([]string{"firewall", "group"}),
				Children: []*Definition{
					{
						Name:     "port-group",
						Path:     []string{"firewall", "group"},
						Node:     nodes.FindChild([]string{"firewall", "group", "port-group"}),
						Children: []*Definition{},
					},
				},
			},
		},
	}

	diffs := expected.Diff(generated)
	if len(diffs) > 0 {
		t.Errorf("Generated tree did not match expected: %s", strings.Join(diffs, ", "))
	}
}

func TestGeneratePopulatedDefinitionTree(t *testing.T) {
	nodes, err := GetGeneratedNodes()
	assert.NoError(t, err, "Generating nodes should not fail")

	expected := &Definition{
		Name: "firewall",
		Path: []string{},
		Node: nodes.FindChild([]string{"firewall"}),
		Children: []*Definition{
			{
				Name:  "name",
				Path:  []string{"firewall"},
				Node:  nodes.FindChild([]string{"firewall", "name"}),
				Value: "foo",
				Children: []*Definition{
					{
						Name:     "default-action",
						Path:     []string{"firewall", "name", "foo"},
						Node:     nodes.FindChild([]string{"firewall", "name", DYNAMIC_NODE, "default-action"}),
						Value:    "drop",
						Children: []*Definition{},
					},
				},
			},
			{
				Name:  "name",
				Path:  []string{"firewall"},
				Node:  nodes.FindChild([]string{"firewall", "name"}),
				Value: "bar",
				Children: []*Definition{
					{
						Name:     "default-action",
						Path:     []string{"firewall", "name", "bar"},
						Node:     nodes.FindChild([]string{"firewall", "name", DYNAMIC_NODE, "default-action"}),
						Value:    "accept",
						Children: []*Definition{},
					},
				},
			},
		},
	}

	generated := generatePopulatedDefinitionTree(
		nodes,
		BasicDefinition{
			Name: "firewall",
			Children: []BasicDefinition{
				{
					Name:  "name",
					Value: "foo",
					Children: []BasicDefinition{
						{
							Name:  "default-action",
							Value: "drop",
						},
					},
				},
				{
					Name:  "name",
					Value: "bar",
					Children: []BasicDefinition{
						{
							Name:  "default-action",
							Value: "accept",
						},
					},
				},
			},
		},
		[]string{},
		[]string{},
	)

	assert.Empty(t, expected.Diff(generated), "Definitions should generate as expected")
}

func TestDefinitions_FindChild(t *testing.T) {
	nodes, err := GetGeneratedNodes()
	assert.NoError(t, err, "Generating nodes should not fail")

	defs := initDefinitions()
	defs.add(
		generatePopulatedDefinitionTree(
			nodes,
			BasicDefinition{
				Name: "firewall",
				Children: []BasicDefinition{
					{
						Name:  "name",
						Value: "foo",
						Children: []BasicDefinition{
							{
								Name:  "default-action",
								Value: "drop",
							},
						},
					},
					{
						Name:  "name",
						Value: "bar",
						Children: []BasicDefinition{
							{
								Name:  "default-action",
								Value: "accept",
							},
						},
					},
				},
			},
			[]string{},
			[]string{},
		),
	)

	assert.Nil(t, defs.FindChild([]any{"interface"}), "Missing root element is nil")
	foo := defs.FindChild([]any{"firewall", "name", "foo"})
	assert.Equal(t, foo.Name, "name")
	assert.Equal(t, foo.Value, "foo")
	assert.Len(t, foo.Children, 1, "Foo firewall has one child")
	assert.Equal(t, foo.Children[0].Value, "drop")

	barAction := defs.FindChild([]any{"firewall", "name", "bar", "default-action"})
	assert.Equal(t, barAction.Name, "default-action")
	assert.Equal(t, barAction.Value, "accept")

	assert.Nil(t, defs.FindChild([]any{"firewall", "name", "nonexistent", "default-action"}))
}

func TestAddValue(t *testing.T) {
	nodes, err := GetGeneratedNodes()
	assert.NoError(t, err, "Generating nodes should not fail")

	defs := initDefinitions()

	defs.add(
		generatePopulatedDefinitionTree(
			nodes,
			BasicDefinition{
				Name: "firewall",
				Children: []BasicDefinition{
					{
						Name: "group",
						Children: []BasicDefinition{
							{
								Name:     "port-group",
								Value:    "web-ports",
								Children: []BasicDefinition{},
							},
						},
					},
				},
			},
			[]string{},
			[]string{},
		),
	)

	path := []string{"firewall", "group", "port-group", "web-ports"}
	nodePath := []string{"firewall", "group", "port-group", DYNAMIC_NODE}

	description := "Ports used by web servers"
	defs.addValue(nodes, path, nodePath, "description", description)
	defs.appendToListValue(nodes, path, nodePath, "port", 80)
	defs.appendToListValue(nodes, path, nodePath, "port", 443)

	descriptionDef := defs.FindChild(config.SliceStrToAny(append(path, "description")))
	assert.NotNil(t, descriptionDef)
	assert.Equal(t, descriptionDef.Value, description)
	assert.Nil(t, descriptionDef.Values)

	portsDef := defs.FindChild(config.SliceStrToAny(append(path, "port")))
	assert.NotNil(t, portsDef)
	assert.Len(t, portsDef.Values, 2, "Port 80 and 443 added")
	assert.Equal(t, portsDef.Values, []any{80, 443})
	assert.Nil(t, portsDef.Value)
}

func TestAddExisting(t *testing.T) {
	nodes, err := GetGeneratedNodes()
	assert.NoError(t, err, "Generating nodes should not fail")

	defs := initDefinitions()
	definition := generatePopulatedDefinitionTree(
		nodes,
		BasicDefinition{
			Name: "firewall",
			Children: []BasicDefinition{
				{
					Name:  "all-ping",
					Value: true,
				},
			},
		},
		[]string{},
		[]string{},
	)

	defs.add(definition)
	pingPath := []any{"firewall", "all-ping"}
	assert.True(t, defs.FindChild(pingPath).Value.(bool))

	definition.Children[0].Value = false

	defs.add(definition)
	assert.False(t, defs.FindChild(pingPath).Value.(bool), "Existing path value should be overridden")
}
