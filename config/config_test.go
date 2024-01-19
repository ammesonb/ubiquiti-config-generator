package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldGetEnv(t *testing.T) {
	assert.False(t, shouldGetEnv("some_value"))
	assert.False(t, shouldGetEnv("special$#!@characters^&*"))
	assert.True(t, shouldGetEnv("$SOME_ENV_VAR"))
}

func TestTrimYAMLEnv(t *testing.T) {
	assert.Equal(t, "value", trimYamlEnv("value"))
	assert.Equal(t, "value", trimYamlEnv("$value"))
	assert.Equal(t, "value$", trimYamlEnv("$value$"))
}
