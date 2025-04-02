package filesystem

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetService(t *testing.T) {
	assert.NoError(t, RegisterService())
	fsOne := GetService()
	fmt.Printf("%v\n", &fsOne)
	fsTwo := GetService()
	fmt.Printf("%v\n", &fsTwo)

	assert.NotNil(t, fsOne)
	assert.NotNil(t, fsTwo)
	assert.Equal(t, fsOne, fsTwo, "Same service interface returned")
	assert.False(t, &fsOne == &fsTwo, "New service instances created")
}

func TestRegisterService(t *testing.T) {
	assert.NoError(t, RegisterService())
}
