package vyos

import (
	"errors"
	"fmt"
	"github.com/ammesonb/ubiquiti-config-generator/utils"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// GetGeneratedNodes reads and parses the YAML file containing an already-analyzed
// sample set of VyOS templates
func GetGeneratedNodes(t *testing.T) *Node {
	t.Helper()
	fixtureFile := "../vyos_test/generated-node-fixtures.yaml"

	if _, err := os.Stat(fixtureFile); errors.Is(err, os.ErrNotExist) {
		if err = generateNodeFixtures("./templates", fixtureFile); err != nil {
			t.Errorf("failed generating test nodes: %v", err)
			t.FailNow()
		}
	}

	node := &Node{}
	data, err := os.ReadFile(fixtureFile)
	if err != nil {
		t.Errorf("could not read node fixtures: %v", err)
		t.FailNow()
	}

	if err = yaml.Unmarshal(data, node); err != nil {
		t.Errorf("failed to parse node fixture data: %v", err)
		t.FailNow()
	}

	assert.NotNil(t, node)
	return node
}

type ErrParseTemplates struct {
	templatePath string
}

func (e ErrParseTemplates) Error() string {
	return fmt.Sprintf("failed to parse node definitions in template dir %s", e.templatePath)
}

type ErrConvertNode struct {
	templatePath string
}

func (e ErrConvertNode) Error() string {
	return fmt.Sprintf("failed to convert nodes in template dir %s to YAML", e.templatePath)
}

type ErrWriteNode struct {
	output string
}

func (e ErrWriteNode) Error() string {
	return fmt.Sprintf("failed to write generated fixtures to %s", e.output)
}

var errAbsPath = "failed to get absolute system path to %s"

func generateNodeFixtures(templateDir string, outputFile string) error {
	templatePath, err := filepath.Abs(templateDir)
	if err != nil {
		return utils.ErrWithCtxParent(errAbsPath, templateDir, err)
	}

	fmt.Printf("Parsing templates from %s\n", templatePath)
	node, err := Parse(templateDir)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrParseTemplates{templatePath: templatePath}, err)
	}

	// This is just to generate for testing
	res, err := yaml.Marshal(node)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrConvertNode{templatePath: templatePath}, err)
	}

	outputPath, err := filepath.Abs(outputFile)
	if err != nil {
		return utils.ErrWithCtxParent(errAbsPath, outputPath, err)
	}
	if err = os.WriteFile(outputPath, res, 0o644); err != nil {
		return fmt.Errorf("%w: %w", ErrWriteNode{output: outputPath}, err)
	}

	return nil
}
