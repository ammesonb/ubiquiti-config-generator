package vyos

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ammesonb/ubiquiti-config-generator/console_logger"
)

var errUnsupportedType = "unsupported templates directory type"
var errFailedStat = "failed to stat firewall node info"

type tStatFunc func(string) (os.FileInfo, error)

func isNodeDef(templatesPath string, statFunc tStatFunc) (bool, error) {
	// Uses arbitrary firewall node.def file to determine if running using nodes or XML
	info, err := statFunc(filepath.Join(templatesPath, "firewall", "node.def"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}

		return false,
			fmt.Errorf("%s: %v", errFailedStat, err)
	}

	return !info.IsDir(), nil
}

// Parse converts the provided templates path into an analyzable list of nodes
func Parse(templatesPath string) (*Node, error) {
	isNode, err := isNodeDef(templatesPath, os.Stat)
	if err != nil {
		return nil, err
	} else if isNode {
		console_logger.DefaultLogger().Info("Detected node templates definitions")
		return ParseNodeDef(templatesPath)
	}

	return nil, fmt.Errorf(errUnsupportedType)
}
