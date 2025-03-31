package vyos

import (
	"errors"
	errors2 "github.com/ammesonb/ubiquiti-config-generator/internal/errors"
	"github.com/ammesonb/ubiquiti-config-generator/mocks"
	"os"
	"path/filepath"

	"github.com/ammesonb/ubiquiti-config-generator/console_logger"
)

var (
	errUnsupportedType = "unsupported type for templates directory: %s"
	errFailedStat      = "failed to stat file %s"
)

func isNodeDef(templatesPath string, fsWrapper *mocks.FsWrapper) (bool, error) {
	// Uses arbitrary firewall node.def file to determine if running using nodes or XML
	firewallPath := filepath.Join(templatesPath, "firewall", "node.def")
	info, err := fsWrapper.Stat(firewallPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}

		return false, errors2.ErrWithCtx(errFailedStat, firewallPath)
	}

	return !info.IsDir(), nil
}

// Parse converts the provided templates path into an analyzable list of nodes
func Parse(templatesPath string, fsWrapper *mocks.FsWrapper) (*Node, error) {
	isNode, err := isNodeDef(templatesPath, fsWrapper)
	if err != nil {
		return nil, err
	} else if isNode {
		console_logger.DefaultLogger().Info("Detected node templates definitions")
		return ParseNodeDef(templatesPath, fsWrapper)
	}

	return nil, errors2.ErrWithCtx(errUnsupportedType, templatesPath)
}
