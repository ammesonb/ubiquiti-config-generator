package vyos

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// GetGeneratedNodes reads and parses the YAML file containing an already-analyzed
// sample set of VyOS templates
func GetGeneratedNodes() (*Node, error) {
	fixtureFile := "../vyos_test/generated-node-fixtures.yaml"

	if _, err := os.Stat(fixtureFile); errors.Is(err, os.ErrNotExist) {
		if err = generateNodeFixtures("./templates"); err != nil {
			return nil, err
		}
	}

	node := &Node{}
	data, err := os.ReadFile(fixtureFile)
	if err != nil {
		return nil, fmt.Errorf("could not read node fixtures: %s", err.Error())
	}

	if err = yaml.Unmarshal(data, node); err != nil {
		return nil, fmt.Errorf("failed to parse node fixture data: %s", err.Error())
	}

	return node, nil
}

func generateNodeFixtures(templateDir string) error {
	fmt.Printf("Parsing templates from %s\n", templateDir)
	node, err := Parse(templateDir)
	if err != nil {
		return fmt.Errorf("failed to parse node definitions in template dir %s: %v", templateDir, err)
	}

	// This is just to generate for testing
	res, err := yaml.Marshal(node)
	if err != nil {
		return fmt.Errorf("failed to convert node to YAML: %v", err)
	}

	if err = os.WriteFile("./generated-node-fixtures.yaml", res, 0o644); err != nil {
		return fmt.Errorf("failed to write generated fixtures: %v", err)
	}

	return nil
}
