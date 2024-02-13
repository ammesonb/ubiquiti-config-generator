package vyos

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNode_FindChild(t *testing.T) {
	node := &Node{
		Name: "root",
		ChildNodes: map[string]*Node{
			"child": {
				Name: "child",
			},
		},
	}

	child := node.FindChild([]string{"missing"})
	assert.Nil(t, child, "No child found for missing node")

	child = node.FindChild([]string{"child"})
	assert.NotNil(t, child, "Child found for dependent node")
}

func TestDiffNode(t *testing.T) {
	node := Node{
		Name:  "node1",
		Type:  "txt",
		IsTag: true,
		Multi: true,
		Path:  "foo",
	}

	diffs := node.diffNode(
		&Node{
			Name:  "node2",
			Type:  "int",
			IsTag: false,
			Multi: false,
			Path:  "bar",
		})

	assert.Len(t, diffs, 5, "Should have 5 differences, got %d", diffs)
}

func TestNodeConstraint_GetProperty(t *testing.T) {
	constraint := NodeConstraint{FailureReason: "failed"}

	assert.Equal(t, "failed", constraint.GetProperty(FailureReason))
	assert.Equal(t, UnsetMinBound, constraint.GetProperty(MinBound))
	assert.Equal(t, UnsetMaxBound, constraint.GetProperty(MaxBound))

	assert.Nil(t, constraint.GetProperty("UnknownProperty"))
}
