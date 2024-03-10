package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMakeVyosPath(t *testing.T) {
	v := MakeVyosPath()
	assert.NotNil(t, v.Path)
	assert.Empty(t, v.Path)
	assert.NotNil(t, v.NodePath)
	assert.Empty(t, v.NodePath)
}

func TestMakeVyosPC(t *testing.T) {
	v := MakeVyosPC("test")
	assert.Equal(t, "test", v.Name)
	assert.False(t, v.IsDynamic)
}

func TestMakeVyosDynamicPC(t *testing.T) {
	v := MakeVyosDynamicPC("test")
	assert.Equal(t, "test", v.Name)
	assert.True(t, v.IsDynamic)
}

func TestVyosPath_Append(t *testing.T) {
	v := MakeVyosPath()
	v.Append(MakeVyosPC("test"), MakeVyosDynamicPC("test2"))
	assert.Equal(t, []string{"test", "test2"}, v.Path)
	assert.Equal(t, []string{"test", DYNAMIC_NODE}, v.NodePath)
}

func TestVyosPath_Extend(t *testing.T) {
	v := MakeVyosPath()
	v.Append(MakeVyosPC("foo"))
	v2 := v.Extend(MakeVyosPC("bar"), MakeVyosDynamicPC("ipsum"))

	assert.Equal(t, []string{"foo"}, v.Path)
	assert.Equal(t, []string{"foo"}, v.NodePath)
	assert.Equal(t, []string{"foo", "bar", "ipsum"}, v2.Path)
	assert.Equal(t, []string{"foo", "bar", DYNAMIC_NODE}, v2.NodePath)
}

func TestVyosPath_DivergeFrom(t *testing.T) {
	v := MakeVyosPath()
	v.Append(MakeVyosPC("foo"))
	v2 := v.DivergeFrom(1, MakeVyosPC("bar"), MakeVyosDynamicPC("ipsum"))

	assert.Equal(t, []string{"foo"}, v.Path)
	assert.Equal(t, []string{"foo"}, v.NodePath)
	assert.Equal(t, []string{"bar", "ipsum"}, v2.Path)
	assert.Equal(t, []string{"bar", DYNAMIC_NODE}, v2.NodePath)
}
