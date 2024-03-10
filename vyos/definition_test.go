package vyos

import (
	"github.com/ammesonb/ubiquiti-config-generator/config"
	"github.com/ammesonb/ubiquiti-config-generator/utils"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestDefinition_Diff(t *testing.T) {
	def := Definition{
		Path:  []string{},
		Name:  utils.DYNAMIC_NODE,
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
		Name:  utils.DYNAMIC_NODE,
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

func TestDiffDefinition(t *testing.T) {
	def := Definition{
		Name:    utils.DYNAMIC_NODE,
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

	assert.Len(t, diffs, 9, "Should have 9 differences, got %d", len(diffs))
}

func TestMerge(t *testing.T) {
	nodes := GetGeneratedNodes(t)
	definitions := initDefinitions()

	definitions.add(&Definition{
		Name:       "interfaces",
		Path:       []string{},
		Node:       nodes.FindChild([]string{"interfaces"}),
		Comment:    "",
		Value:      nil,
		Values:     nil,
		Children:   []*Definition{},
		ParentNode: nil,
	})

	definitions.add(&Definition{
		Name:       "firewall",
		Path:       []string{},
		Node:       nodes.FindChild([]string{"firewall"}),
		Comment:    "",
		Value:      nil,
		Values:     nil,
		Children:   []*Definition{},
		ParentNode: nil,
	})

	definitions.add(&Definition{
		Name:       "all-ping",
		Path:       []string{"firewall"},
		Node:       nodes.FindChild([]string{"firewall", "all-ping"}),
		Comment:    "",
		Value:      "enable",
		Values:     nil,
		Children:   []*Definition{},
		ParentNode: nodes.FindChild([]string{"firewall"}),
	})
	definitions.add(&Definition{
		Name:       "broadcast-ping",
		Path:       []string{"firewall"},
		Node:       nodes.FindChild([]string{"firewall", "broadcast-ping"}),
		Comment:    "",
		Value:      "disable",
		Values:     nil,
		Children:   []*Definition{},
		ParentNode: nodes.FindChild([]string{"firewall"}),
	})
	definitions.add(&Definition{
		Name:       "group",
		Path:       []string{"firewall"},
		Node:       nodes.FindChild([]string{"firewall", "group"}),
		Comment:    "",
		Value:      nil,
		Values:     nil,
		Children:   []*Definition{},
		ParentNode: nodes.FindChild([]string{"firewall"}),
	})
	definitions.add(&Definition{
		Name:       "port-group",
		Path:       []string{"firewall", "group"},
		Node:       nodes.FindChild([]string{"firewall", "group", "port-group"}),
		Comment:    "",
		Value:      "test-port-group",
		Values:     nil,
		Children:   []*Definition{},
		ParentNode: nodes.FindChild([]string{"firewall", "group"}),
	})
	definitions.add(&Definition{
		Name:       "description",
		Path:       []string{"firewall", "group", "port-group", "test-port-group"},
		Node:       nodes.FindChild([]string{"firewall", "group", "port-group", utils.DYNAMIC_NODE, "description"}),
		Comment:    "",
		Value:      "a test port group",
		Values:     nil,
		Children:   []*Definition{},
		ParentNode: nodes.FindChild([]string{"firewall", "group", "port-group"}),
	})
	definitions.add(&Definition{
		Name:       "port",
		Path:       []string{"firewall", "group", "port-group", "test-port-group"},
		Node:       nodes.FindChild([]string{"firewall", "group", "port-group", utils.DYNAMIC_NODE, "port"}),
		Comment:    "",
		Value:      nil,
		Values:     []any{53, 123},
		Children:   []*Definition{},
		ParentNode: nodes.FindChild([]string{"firewall", "group", "port-group"}),
	})

	others := initDefinitions()

	domainName := "ubiquiti"
	others.add(&Definition{
		Name:       "system",
		Path:       []string{},
		Node:       nodes.FindChild([]string{"system"}),
		Comment:    "",
		Value:      nil,
		Values:     nil,
		Children:   []*Definition{},
		ParentNode: nil,
	})
	others.add(&Definition{
		Name:       "domain-name",
		Path:       []string{"system"},
		Node:       nodes.FindChild([]string{"system", "domain-name"}),
		Comment:    "",
		Value:      domainName,
		Values:     nil,
		Children:   []*Definition{},
		ParentNode: nodes.FindChild([]string{"system"}),
	})

	others.add(&Definition{
		Name:       "interfaces",
		Path:       []string{},
		Node:       nodes.FindChild([]string{"interfaces"}),
		Comment:    "",
		Value:      nil,
		Values:     nil,
		Children:   []*Definition{},
		ParentNode: nil,
	})
	others.add(&Definition{
		Name:       "ethernet",
		Path:       []string{"interfaces"},
		Node:       nodes.FindChild([]string{"interfaces", "ethernet"}),
		Comment:    "",
		Value:      "eth0",
		Values:     nil,
		Children:   []*Definition{},
		ParentNode: nodes.FindChild([]string{"interfaces"}),
	})
	others.add(&Definition{
		Name:       "address",
		Path:       []string{"interfaces", "ethernet", "eth0"},
		Node:       nodes.FindChild([]string{"interfaces", "ethernet", utils.DYNAMIC_NODE, "address"}),
		Comment:    "",
		Value:      "192.168.0.1",
		Values:     nil,
		Children:   []*Definition{},
		ParentNode: nodes.FindChild([]string{"interfaces", "ethernet"}),
	})
	others.add(&Definition{
		Name:       "vif",
		Path:       []string{"interfaces", "ethernet", "eth0"},
		Node:       nodes.FindChild([]string{"interfaces", "ethernet", utils.DYNAMIC_NODE, "vif"}),
		Comment:    "",
		Value:      99,
		Values:     nil,
		Children:   []*Definition{},
		ParentNode: nodes.FindChild([]string{"interfaces", "ethernet"}),
	})

	others.add(&Definition{
		Name:       "firewall",
		Path:       []string{},
		Node:       nodes.FindChild([]string{"firewall"}),
		Comment:    "",
		Value:      nil,
		Values:     nil,
		Children:   []*Definition{},
		ParentNode: nil,
	})
	others.add(&Definition{
		Name:       "all-ping",
		Path:       []string{"firewall"},
		Node:       nodes.FindChild([]string{"firewall", "all-ping"}),
		Comment:    "",
		Value:      "enable",
		Values:     nil,
		Children:   []*Definition{},
		ParentNode: nodes.FindChild([]string{"firewall"}),
	})
	others.add(&Definition{
		Name:       "ip-src-route",
		Path:       []string{"firewall"},
		Node:       nodes.FindChild([]string{"firewall", "ip-src-route"}),
		Comment:    "",
		Value:      "disable",
		Values:     nil,
		Children:   []*Definition{},
		ParentNode: nodes.FindChild([]string{"firewall"}),
	})
	others.add(&Definition{
		Name:       "group",
		Path:       []string{"firewall"},
		Node:       nodes.FindChild([]string{"firewall", "group"}),
		Comment:    "",
		Value:      nil,
		Values:     nil,
		Children:   []*Definition{},
		ParentNode: nodes.FindChild([]string{"firewall"}),
	})
	others.add(&Definition{
		Name:       "port-group",
		Path:       []string{"firewall", "group"},
		Node:       nodes.FindChild([]string{"firewall", "group", "port-group"}),
		Comment:    "",
		Value:      "test-port-group",
		Values:     nil,
		Children:   []*Definition{},
		ParentNode: nodes.FindChild([]string{"firewall", "group"}),
	})
	others.add(&Definition{
		Name:       "port",
		Path:       []string{"firewall", "group", "port-group", "test-port-group"},
		Node:       nodes.FindChild([]string{"firewall", "group", "port-group", utils.DYNAMIC_NODE, "port"}),
		Comment:    "",
		Value:      nil,
		Values:     []any{53, 123},
		Children:   []*Definition{},
		ParentNode: nodes.FindChild([]string{"firewall", "group", "port-group"}),
	})
	others.add(&Definition{
		Name:       "port-group",
		Path:       []string{"firewall", "group"},
		Node:       nodes.FindChild([]string{"firewall", "group", "port-group"}),
		Comment:    "",
		Value:      "http-port-group",
		Values:     nil,
		Children:   []*Definition{},
		ParentNode: nodes.FindChild([]string{"firewall", "group"}),
	})
	others.add(&Definition{
		Name:       "description",
		Path:       []string{"firewall", "group", "port-group", "http-port-group"},
		Node:       nodes.FindChild([]string{"firewall", "group", "port-group", utils.DYNAMIC_NODE, "description"}),
		Comment:    "",
		Value:      "HTTP ports",
		Values:     nil,
		Children:   []*Definition{},
		ParentNode: nodes.FindChild([]string{"firewall", "group", "port-group"}),
	})
	others.add(&Definition{
		Name:       "port",
		Path:       []string{"firewall", "group", "port-group", "http-port-group"},
		Node:       nodes.FindChild([]string{"firewall", "group", "port-group", utils.DYNAMIC_NODE, "port"}),
		Comment:    "",
		Value:      nil,
		Values:     []any{80, 443},
		Children:   []*Definition{},
		ParentNode: nodes.FindChild([]string{"firewall", "group", "port-group"}),
	})

	err := definitions.merge(others)
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
	nodes := GetGeneratedNodes(t)
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
		Node:     nodes.FindChild([]string{"firewall", "group", "port-group", utils.DYNAMIC_NODE, "description"}),
		Comment:  "",
		Value:    "a test port group",
		Values:   nil,
		Children: []*Definition{},
	})
	definitions.add(&Definition{
		Name:     "port",
		Path:     []string{"firewall", "group", "port-group", "test-port-group"},
		Node:     nodes.FindChild([]string{"firewall", "group", "port-group", utils.DYNAMIC_NODE, "port"}),
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

	err := definitions.merge(others)
	assert.ErrorContainsf(t, err, "differences between nodes at path firewall/all-ping", "Expected all-ping to have conflicting value")
}

func TestGenerateDefinitionTree(t *testing.T) {
	nodes := GetGeneratedNodes(t)

	path := utils.MakeVyosPath()
	path.Append(
		utils.MakeVyosPC("firewall"),
		utils.MakeVyosPC("group"),
		utils.MakeVyosPC("port-group"),
		utils.MakeVyosDynamicPC("test-port-group"),
	)
	generated := generateSparseDefinitionTree(nodes, path)

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
						Name:       "port-group",
						Path:       []string{"firewall", "group"},
						Node:       nodes.FindChild([]string{"firewall", "group", "port-group"}),
						Value:      "test-port-group",
						ParentNode: nodes.FindChild([]string{"firewall", "group"}),
					},
				},
				ParentNode: nodes.FindChild([]string{"firewall"}),
			},
		},
		ParentNode: nil,
	}

	diffs := expected.Diff(generated)
	if len(diffs) > 0 {
		t.Errorf("Generated tree did not match expected: %s", strings.Join(diffs, ", "))
	}
}

func TestGenerateMultiDynamicNodeDefinitionTree(t *testing.T) {
	nodes := GetGeneratedNodes(t)

	path := utils.MakeVyosPath()
	path.Append(
		utils.MakeVyosPC("interfaces"),
		utils.MakeVyosPC("ethernet"),
		utils.MakeVyosDynamicPC("eth1"),
		utils.MakeVyosPC("vif"),
		utils.MakeVyosDynamicPC("10"),
	)
	generated := generateSparseDefinitionTree(nodes, path)

	expected := &Definition{
		Name: "interfaces",
		Path: []string{},
		Node: nodes.FindChild([]string{"interfaces"}),
		Children: []*Definition{
			{
				Name:  "ethernet",
				Path:  []string{"interfaces"},
				Node:  nodes.FindChild([]string{"interfaces", "ethernet"}),
				Value: "eth1",
				Children: []*Definition{
					{
						Name:       "vif",
						Path:       []string{"interfaces", "ethernet", "eth1"},
						Node:       nodes.FindChild([]string{"interfaces", "ethernet", utils.DYNAMIC_NODE, "vif"}),
						Value:      "10",
						ParentNode: nodes.FindChild([]string{"interfaces", "ethernet"}),
					},
				},
				ParentNode: nodes.FindChild([]string{"interfaces"}),
			},
		},
		ParentNode: nil,
	}

	diffs := expected.Diff(generated)
	if len(diffs) > 0 {
		t.Errorf("Generated tree did not match expected: %s", strings.Join(diffs, ", "))
	}
}

func TestGeneratePopulatedDefinitionTreeFromRoot(t *testing.T) {
	nodes := GetGeneratedNodes(t)

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
						Name:       "default-action",
						Path:       []string{"firewall", "name", "foo"},
						Node:       nodes.FindChild([]string{"firewall", "name", utils.DYNAMIC_NODE, "default-action"}),
						Value:      "drop",
						Children:   []*Definition{},
						ParentNode: nodes.FindChild([]string{"firewall", "name"}),
					},
				},
				ParentNode: nodes.FindChild([]string{"firewall"}),
			},
			{
				Name:  "name",
				Path:  []string{"firewall"},
				Node:  nodes.FindChild([]string{"firewall", "name"}),
				Value: "bar",
				Children: []*Definition{
					{
						Name:       "default-action",
						Path:       []string{"firewall", "name", "bar"},
						Node:       nodes.FindChild([]string{"firewall", "name", utils.DYNAMIC_NODE, "default-action"}),
						Value:      "accept",
						Children:   []*Definition{},
						ParentNode: nodes.FindChild([]string{"firewall", "name"}),
					},
				},
				ParentNode: nodes.FindChild([]string{"firewall"}),
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
		utils.MakeVyosPath(),
		nil,
	)

	assert.Empty(t, expected.Diff(generated), "Definitions should generate as expected")
}

func TestGeneratePopulatedDefinitionTreeNested(t *testing.T) {
	nodes := GetGeneratedNodes(t)

	path := utils.MakeVyosPath()
	path.Append(
		utils.MakeVyosPC("service"),
		utils.MakeVyosPC("dhcp-server"),
		utils.MakeVyosPC("shared-network-name"),
		utils.MakeVyosDynamicPC("test-service"),
		utils.MakeVyosPC("subnet"),
		utils.MakeVyosDynamicPC("10.0.0.0/24"),
	)

	expected := &Definition{
		Name:  "start",
		Path:  path.Path,
		Node:  nodes.FindChild(utils.CopySliceWith(path.NodePath, "start")),
		Value: "10.0.0.240",
		Children: []*Definition{
			{
				Name:       "stop",
				Path:       utils.CopySliceWith(path.Path, "start", "10.0.0.240"),
				Node:       nodes.FindChild(utils.CopySliceWith(path.NodePath, "start", utils.DYNAMIC_NODE, "stop")),
				Value:      "10.0.0.255",
				Children:   []*Definition{},
				ParentNode: nodes.FindChild(utils.CopySliceWith(path.NodePath, "start")),
			},
		},
		ParentNode: nodes.FindChild(utils.AllExcept(path.NodePath, 1)),
	}

	generated := generatePopulatedDefinitionTree(
		nodes,
		BasicDefinition{
			Name:  "start",
			Value: "10.0.0.240",
			Children: []BasicDefinition{
				{
					Name:  "stop",
					Value: "10.0.0.255",
				},
			},
		},
		path,
		// Skip the tag placeholder for the parent, since that is what would be done normally
		nodes.FindChild(utils.AllExcept(path.NodePath, 1)),
	)

	assert.Empty(t, expected.Diff(generated), "Definitions should generate as expected")
}
func TestDefinitions_FindChild(t *testing.T) {
	nodes := GetGeneratedNodes(t)

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
			utils.MakeVyosPath(),
			nil,
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

	assert.Equal(t, foo.ParentNode.Name, "firewall")
	assert.False(t, foo.ParentNode.IsTag)
	assert.Equal(t, barAction.ParentNode.Name, "name")
	assert.True(t, barAction.ParentNode.IsTag)
}

func TestAddValue(t *testing.T) {
	nodes := GetGeneratedNodes(t)

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
			utils.MakeVyosPath(),
			nil,
		),
	)

	path := utils.MakeVyosPath()
	path.Append(
		utils.MakeVyosPC("firewall"),
		utils.MakeVyosPC("group"),
		utils.MakeVyosPC("port-group"),
		utils.MakeVyosDynamicPC("web-ports"),
	)

	description := "Ports used by web servers"
	defs.addValue(nodes, path, "description", description)
	defs.appendToListValue(nodes, path, "port", 80)
	defs.appendToListValue(nodes, path, "port", 443)

	descriptionDef := defs.FindChild(config.SliceStrToAny(utils.CopySliceWith(path.Path, "description")))
	assert.NotNil(t, descriptionDef)
	assert.Equal(t, descriptionDef.Value, description)
	assert.Nil(t, descriptionDef.Values)

	assert.Equal(t, descriptionDef.ParentNode.Name, "port-group")
	assert.True(t, descriptionDef.ParentNode.IsTag)

	portsDef := defs.FindChild(config.SliceStrToAny(utils.CopySliceWith(path.Path, "port")))
	assert.NotNil(t, portsDef)
	assert.Len(t, portsDef.Values, 2, "Port 80 and 443 added")
	assert.Equal(t, portsDef.Values, []any{80, 443})
	assert.Nil(t, portsDef.Value)

	assert.Equal(t, portsDef.ParentNode.Name, "port-group")
	assert.True(t, portsDef.ParentNode.IsTag)
}

func TestAddExisting(t *testing.T) {
	nodes := GetGeneratedNodes(t)

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
		utils.MakeVyosPath(),
		nil,
	)

	defs.add(definition)
	pingPath := []any{"firewall", "all-ping"}
	assert.True(t, defs.FindChild(pingPath).Value.(bool))

	definition.Children[0].Value = false

	defs.add(definition)
	assert.False(t, defs.FindChild(pingPath).Value.(bool), "Existing path value should be overridden")
}

func TestAddDoesNotOverwrite(t *testing.T) {
	nodes := GetGeneratedNodes(t)

	defs := initDefinitions()
	ethPath := utils.MakeVyosPath()
	ethPath.Append(
		utils.MakeVyosPC("interfaces"),
		utils.MakeVyosPC("ethernet"),
		utils.MakeVyosDynamicPC("eth0"),
	)
	defs.add(generateSparseDefinitionTree(nodes, ethPath))
	defs.addValue(nodes, ethPath, "description", "foo")
	defs.addValue(nodes, ethPath, "duplex", "auto")
	defs.addValue(nodes, ethPath, "speed", "auto")

	vifPath := ethPath.Extend(utils.MakeVyosPC("vif"), utils.MakeVyosDynamicPC("10"))
	defs.add(generateSparseDefinitionTree(nodes, vifPath))

	eth := defs.FindChild(config.SliceStrToAny(ethPath.Path))
	assert.NotNil(t, eth)
	if eth == nil {
		t.FailNow()
	}
	assert.Len(t, eth.Children, 4, "Should have four children: description, duplex, speed, vif")
}

func TestEnsureTree(t *testing.T) {
	nodes := GetGeneratedNodes(t)
	defs := initDefinitions()

	goodPath := utils.MakeVyosPath()
	goodPath.Append(
		utils.MakeVyosPC("firewall"),
		utils.MakeVyosPC("name"),
		utils.MakeVyosDynamicPC("test-firewall"),
		utils.MakeVyosPC("rule"),
		utils.MakeVyosDynamicPC("100"),
	)
	assert.Nil(t, defs.ensureTree(nodes, goodPath))

	badPath := utils.MakeVyosPath()
	badPath.Append(
		utils.MakeVyosPC("firewall"),
		utils.MakeVyosPC("name"),
	)
	assert.ErrorIs(
		t,
		defs.ensureTree(nodes, badPath),
		utils.ErrWithCtx(errUnmatchedDynamicNode, "firewall/name"),
	)
	assert.ErrorIs(
		t,
		defs.ensureTree(&Node{}, badPath),
		utils.ErrWithCtx(errNonexistentNode, "firewall"),
	)
	badPath.Path = append(badPath.Path, "foo")
	assert.ErrorIs(
		t,
		defs.ensureTree(nodes, badPath),
		utils.ErrWithVarCtx(errDiffLength, 3, 2),
	)
}
