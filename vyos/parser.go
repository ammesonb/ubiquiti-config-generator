package vyos

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func isNodeDef(templatesPath string) (bool, error) {
	// Uses arbitrary firewall node.def file to determine if running using nodes or XML
	info, err := os.Stat(filepath.Join(templatesPath, "firewall", "node.def"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}

		return false,
			fmt.Errorf("Failed to stat firewall node info: %s", err.Error())
	}

	return !info.IsDir(), nil
}

// Parse converts the provided templates path into an analyzable list of nodes
func Parse(templatesPath string) (*Node, error) {
	isNode, err := isNodeDef(templatesPath)
	if err != nil {
		return nil, err
	} else if isNode {
		return ParseNodeDef(templatesPath, "")
	}

	return nil, fmt.Errorf("Unsupported templates directory type")
}
