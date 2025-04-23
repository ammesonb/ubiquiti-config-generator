package vyos

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ammesonb/ubiquiti-config-generator/internal/errors"
	"github.com/ammesonb/ubiquiti-config-generator/services/filesystem"
	"github.com/ammesonb/ubiquiti-config-generator/services/filesystem/filesystemfakes"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	defaultFs := &filesystem.OSService{}

	t.Run("sample nodes parse successfully", func(t *testing.T) {
		nodes, err := Parse("./test-files/node", defaultFs)
		assert.NoError(t, err)
		assert.NotNil(t, nodes, "Node definitions should parse successfully")
		assert.NotNil(t, nodes.FindChild([]string{"firewall"}), "Firewall node parsed")

		nodes, err = Parse("./test-files/xml", defaultFs)
		assert.Nil(t, nodes, "No XML nodes generated")
		assert.Error(t, err)
		assert.ErrorIs(t, err, errors.ErrWithCtx(errUnsupportedType, "./test-files/xml"))
	})

	t.Run("directory does not exist", func(t *testing.T) {
		fakeFs := filesystemfakes.FakeFileSystemService{}
		fakeFs.StatReturns(nil, os.ErrNotExist)

		nodes, err := Parse("./test-files/invalid-node-dir", &fakeFs)
		assert.Nil(t, nodes, "No nodes generated")
		assert.Error(t, err)
		assert.ErrorIs(t, err, errors.ErrWithCtx(errUnsupportedType, "./test-files/invalid-node-dir"))
	})

	t.Run("failed stat", func(t *testing.T) {
		fakeFs := filesystemfakes.FakeFileSystemService{}
		fakeFs.StatReturns(nil, errors.Err(errFailedStat))

		nodes, err := Parse("./test-files/node", &fakeFs)
		assert.Nil(t, nodes, "No nodes generated")
		assert.Error(t, err)
		assert.ErrorIs(t, err, errors.ErrWithCtx(errFailedStat, filepath.Join("./test-files/node", "firewall", "node.def")))
	})
}

func TestParseErrors(t *testing.T) {
	fakeFs := filesystemfakes.FakeFileSystemService{}
	fakeFs.StatReturns(nil, errors.Err(errFailedStat))

	isNode, err := isNodeDef("/failure", &fakeFs)
	assert.False(t, isNode, "Not nodes if function errors")
	assert.Error(t, err, "Error thrown on failure")
	assert.ErrorIs(t, err, errors.ErrWithCtx(errFailedStat, filepath.Join("/failure", "firewall", "node.def")))
}
