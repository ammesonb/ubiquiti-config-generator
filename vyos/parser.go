package vyos

import (
	"errors"
	"github.com/ammesonb/ubiquiti-config-generator/mocks"
	"github.com/ammesonb/ubiquiti-config-generator/utils"
	"os"
	"path/filepath"

	"github.com/ammesonb/ubiquiti-config-generator/console_logger"
)

var (
	errUnsupportedType = "unsupported type for templates directory: %s"
	errFailedStat      = "failed to stat file %s"
)

type tStatFunc func(string) (os.FileInfo, error)

func isNodeDef(templatesPath string, statFunc tStatFunc) (bool, error) {
	// Uses arbitrary firewall node.def file to determine if running using nodes or XML
	firewallPath := filepath.Join(templatesPath, "firewall", "node.def")
	info, err := statFunc(firewallPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}

		return false, utils.ErrWithCtx(errFailedStat, firewallPath)
	}

	return !info.IsDir(), nil
}

// Parse converts the provided templates path into an analyzable list of nodes
func Parse(templatesPath string, fsWrapper *mocks.FsWrapper) (*Node, error) {
	isNode, err := isNodeDef(templatesPath, os.Stat)
	if err != nil {
		return nil, err
	} else if isNode {
		console_logger.DefaultLogger().Info("Detected node templates definitions")
		return ParseNodeDef(templatesPath, fsWrapper)
	}

	return nil, utils.ErrWithCtx(errUnsupportedType, templatesPath)
}
