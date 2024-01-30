package vyos

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var errStatFailed = "failed to stat file"

func TestParse(t *testing.T) {
	nodes, err := Parse("./test-files/node")
	assert.NoError(t, err)
	assert.NotNil(t, nodes, "Node definitions should parse successfully")
	assert.NotNil(t, nodes.FindChild([]string{"firewall"}), "Firewall node parsed")

	nodes, err = Parse("./test-files/xml")
	assert.Nil(t, nodes, "No XML nodes generated")
	assert.Error(t, err)
	assert.ErrorContains(t, err, errUnsupportedType)

	nodes, err = Parse("./test-files/invalid-node-dir")
	assert.Nil(t, nodes, "No nodes generated")
	assert.Error(t, err)
	assert.ErrorContains(t, err, errUnsupportedType)
}

func TestParseErrors(t *testing.T) {
	fmt.Println("Testing stat")
	isNode, err := isNodeDef(
		"/failure",
		func(_ string) (os.FileInfo, error) { return nil, fmt.Errorf(errStatFailed) },
	)
	assert.False(t, isNode, "Not nodes if function errors")
	assert.Error(t, err, "Error thrown on failure")
	assert.ErrorContains(t, err, errStatFailed)
}
