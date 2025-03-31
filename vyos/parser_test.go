package vyos

import (
	"fmt"
	"github.com/ammesonb/ubiquiti-config-generator/internal/errors"
	"github.com/ammesonb/ubiquiti-config-generator/mocks"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	wrapper := mocks.GetFsWrapper()
	nodes, err := Parse("./test-files/node", wrapper)
	assert.NoError(t, err)
	assert.NotNil(t, nodes, "Node definitions should parse successfully")
	assert.NotNil(t, nodes.FindChild([]string{"firewall"}), "Firewall node parsed")

	nodes, err = Parse("./test-files/xml", wrapper)
	assert.Nil(t, nodes, "No XML nodes generated")
	assert.Error(t, err)
	assert.ErrorIs(t, err, errors.ErrWithCtx(errUnsupportedType, "./test-files/xml"))

	nodes, err = Parse("./test-files/invalid-node-dir", wrapper)
	assert.Nil(t, nodes, "No nodes generated")
	assert.Error(t, err)
	assert.ErrorIs(t, err, errors.ErrWithCtx(errUnsupportedType, "./test-files/invalid-node-dir"))

	wrapper.Stat = func(filename string) (os.FileInfo, error) {
		return nil, errors.Err(errFailedStat)
	}

	nodes, err = Parse("./test-files/node", wrapper)
	assert.Nil(t, nodes, "No nodes generated")
	assert.Error(t, err)
	assert.ErrorIs(t, err, errors.ErrWithCtx(errFailedStat, filepath.Join("./test-files/node", "firewall", "node.def")))
}

func TestParseErrors(t *testing.T) {
	fmt.Println("Testing stat")
	wrapper := &mocks.FsWrapper{
		Stat: func(_ string) (os.FileInfo, error) { return nil, errors.Err(errFailedStat) },
	}
	isNode, err := isNodeDef("/failure", wrapper)
	assert.False(t, isNode, "Not nodes if function errors")
	assert.Error(t, err, "Error thrown on failure")
	assert.ErrorIs(t, err, errors.ErrWithCtx(errFailedStat, filepath.Join("/failure", "firewall", "node.def")))
}
