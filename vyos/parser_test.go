package vyos

import (
	"path/filepath"
	"testing"

	"github.com/ammesonb/ubiquiti-config-generator/internal/errors"
	mockFs "github.com/ammesonb/ubiquiti-config-generator/mocks/filesystem"
	"github.com/ammesonb/ubiquiti-config-generator/services/filesystem"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	assert.NoError(t, filesystem.RegisterService(t.Context()))

	nodes, err := Parse("./test-files/node")
	assert.NoError(t, err)
	assert.NotNil(t, nodes, "Node definitions should parse successfully")
	assert.NotNil(t, nodes.FindChild([]string{"firewall"}), "Firewall node parsed")

	nodes, err = Parse("./test-files/xml")
	assert.Nil(t, nodes, "No XML nodes generated")
	assert.Error(t, err)
	assert.ErrorIs(t, err, errors.ErrWithCtx(errUnsupportedType, "./test-files/xml"))

	nodes, err = Parse("./test-files/invalid-node-dir")
	assert.Nil(t, nodes, "No nodes generated")
	assert.Error(t, err)
	assert.ErrorIs(t, err, errors.ErrWithCtx(errUnsupportedType, "./test-files/invalid-node-dir"))

	assert.NoError(t, mockFs.MockFileSystem(t))
	mockedFs := filesystem.GetService().(*mockFs.MockedFileSystem)
	mockedFs.SetNextResult(mockFs.StatFn, []any{nil, errors.Err(errFailedStat)})

	nodes, err = Parse("./test-files/node")
	assert.Nil(t, nodes, "No nodes generated")
	assert.Error(t, err)
	assert.ErrorIs(t, err, errors.ErrWithCtx(errFailedStat, filepath.Join("./test-files/node", "firewall", "node.def")))
}

func TestParseErrors(t *testing.T) {
	assert.NoError(t, mockFs.MockFileSystem(t))
	mockedFs := filesystem.GetService().(*mockFs.MockedFileSystem)
	mockedFs.SetNextResult(mockFs.StatFn, []any{nil, errors.Err(errFailedStat)})
	isNode, err := isNodeDef("/failure")
	assert.False(t, isNode, "Not nodes if function errors")
	assert.Error(t, err, "Error thrown on failure")
	assert.ErrorIs(t, err, errors.ErrWithCtx(errFailedStat, filepath.Join("/failure", "firewall", "node.def")))
}
