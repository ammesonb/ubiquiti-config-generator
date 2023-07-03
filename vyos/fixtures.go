package vyos

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// GetGeneratedNodes reads and parses the YAML file containing an already-analyzed
// sample set of VyOS templates
func GetGeneratedNodes() (*Node, error) {
	node := &Node{}
	data, err := os.ReadFile("../vyos_test/generated-node-fixtures.yaml")
	if err != nil {
		return nil, fmt.Errorf("Could not read node fixtures: %s", err.Error())
	}

	if err = yaml.Unmarshal(data, node); err != nil {
		return nil, fmt.Errorf("Failed to parse node fixture data: %s", err.Error())
	}

	return node, nil
}
