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

func TestNodeConstraint_GetProperty(t *testing.T) {
	constraint := NodeConstraint{FailureReason: "failed"}

	assert.Equal(t, "failed", constraint.GetProperty(FailureReason))
	assert.Equal(t, UnsetMinBound, constraint.GetProperty(MinBound))
	assert.Equal(t, UnsetMaxBound, constraint.GetProperty(MaxBound))

	assert.Nil(t, constraint.GetProperty("UnknownProperty"))
}
