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
