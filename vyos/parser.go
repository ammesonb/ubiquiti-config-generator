package vyos

import (
	errs "errors"
	"os"
	"path/filepath"

	"github.com/ammesonb/ubiquiti-config-generator/internal/errors"
	"github.com/ammesonb/ubiquiti-config-generator/services/filesystem"

	"github.com/ammesonb/ubiquiti-config-generator/console_logger"
)

var (
	errUnsupportedType = "unsupported type for templates directory: %s"
	errFailedStat      = "failed to stat file %s"
)

func isNodeDef(templatesPath string) (bool, error) {
	// Uses arbitrary firewall node.def file to determine if running using nodes or XML
	fsService := filesystem.GetService()
	firewallPath := filepath.Join(templatesPath, "firewall", "node.def")
	info, err := fsService.Stat(firewallPath)
	if err != nil {
		if errs.Is(err, os.ErrNotExist) {
			return false, nil
		}

		return false, errors.ErrWithCtx(errFailedStat, firewallPath)
	}

	return !info.IsDir(), nil
}

// Parse converts the provided templates path into an analyzable list of nodes
func Parse(templatesPath string) (*Node, error) {
	isNode, err := isNodeDef(templatesPath)
	if err != nil {
		return nil, err
	} else if isNode {
		console_logger.DefaultLogger().Info("Detected node templates definitions")
		return ParseNodeDef(templatesPath)
	}

	return nil, errors.ErrWithCtx(errUnsupportedType, templatesPath)
}
