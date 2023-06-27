package vyos

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"testing"
)

// TestAddOption uses basic configuration stubs to verify variables are set properly
func TestAddOption(t *testing.T) {
	n := &Node{
		Name:        "TestNode",
		IsTag:       false,
		Multi:       false,
		Children:    map[string]*Node{},
		Constraints: []NodeConstraint{},
	}
	helpValues := &[]string{}
	allowed := ""
	expression := ""

	t.Run("tag", func(t *testing.T) {
		addOption("tag", "tag:", helpValues, &allowed, &expression, n)
		if !n.IsTag {
			t.Errorf("Tag flag not set on node")
		}
	})
	t.Run("multi", func(t *testing.T) {
		addOption("multi", "multi:", helpValues, &allowed, &expression, n)
		if !n.Multi {
			t.Errorf("Multi flag not set on node")
		}
	})
	t.Run("type", func(t *testing.T) {
		addOption("type", "type: txt", helpValues, &allowed, &expression, n)
		if n.Type != "txt" {
			t.Errorf("Expected node type to be txt, got '%s'", n.Type)
		}
	})
	t.Run("help", func(t *testing.T) {
		addOption("help", "help: some description", helpValues, &allowed, &expression, n)
		if n.Help != "some description" {
			t.Errorf("Got unexpected node description: '%s'", n.Type)
		}
	})
	t.Run("val_help", func(t *testing.T) {
		addOption("val_help", "val_help: u32", helpValues, &allowed, &expression, n)
		if len(*helpValues) == 0 {
			t.Errorf("No value added to help descriptions")
		} else if len(*helpValues) > 1 {
			t.Errorf("Too many values added to help descriptions: %#v", *helpValues)
		} else if (*helpValues)[0] != "u32" {
			t.Errorf("Got unexpected help value: %s", (*helpValues)[0])
		}
	})
	t.Run("allowed", func(t *testing.T) {
		addOption("allowed", "allowed: cli-shell-api", helpValues, &allowed, &expression, n)
		if allowed != "cli-shell-api" {
			t.Errorf("Got unexpected allowed value: '%s'", allowed)
		}
	})
	t.Run("syntax", func(t *testing.T) {
		addOption("syntax", "syntax:expression: pattern $VAR(@)", helpValues, &allowed, &expression, n)
		if expression != "pattern $VAR(@)" {
			t.Errorf("Got unexpected expression value: '%s'", expression)
		}
	})
}

func getBlank() string {
	if rand.Intn(2) == 1 {
		return "\n"
	}

	return ""
}

func constructDefinition(
	ntype *string,
	help *string,
	isTag *bool,
	isMulti *bool,
	helpValues []string,
	allowed *string,
	syntax *string,
) string {
	value := ""
	if ntype != nil {
		value += fmt.Sprintf("type: %s\n%s", *ntype, getBlank())
	}
	if help != nil {
		value += fmt.Sprintf("help: %s\n%s", *help, getBlank())
	}
	if isTag != nil && *isTag {
		value += fmt.Sprintf("tag:\n%s", getBlank())
	}
	if isMulti != nil && *isMulti {
		value += fmt.Sprintf("multi:\n%s", getBlank())
	}
	if len(helpValues) > 0 {
		for _, v := range helpValues {
			value += fmt.Sprintf("val_help: %s\n", v)
		}
	}
	if allowed != nil {
		value += fmt.Sprintf("allowed: %s\n%s", *allowed, getBlank())
	}
	if syntax != nil {
		value += fmt.Sprintf("syntax:expression: %s\n%s", *syntax, getBlank())
	}

	return value
}

func createTestNode(isTag bool, isMulti bool, options []string) (*Node, error) {
	node := &Node{
		Name:        "TestNode",
		IsTag:       isTag,
		Multi:       isMulti,
		Children:    map[string]*Node{},
		Constraints: []NodeConstraint{},
	}

	var buffer bytes.Buffer
	for _, option := range options {
		buffer.Truncate(0)
		buffer.WriteString(option)

		err := parseDefinition(ioutil.NopCloser(&buffer), node)
		if err != nil {
			return nil, err
		}
	}

	return node, nil
}

func validateNode(node *Node, isTag bool, isMulti bool, ntype string, help string) []string {
	errors := []string{}

	if node.IsTag != isTag {
		errors = append(errors, fmt.Sprintf("Node tag should be %t, got %t", isTag, node.IsTag))
	}
	if node.Multi != isMulti {
		errors = append(errors, fmt.Sprintf("Node multi should be %t, got %t", isMulti, node.Multi))
	}
	if node.Type != ntype {
		errors = append(errors, fmt.Sprintf("Node should have type %s, got: '%s'", ntype, node.Type))
	}
	if node.Help != help {
		errors = append(errors, fmt.Sprintf("Node should have help %s, got: '%s'", help, node.Help))
	}

	return errors
}

// TestSimpleDefinition checks basic type/help attributes for a node
func TestSimpleDefinition(t *testing.T) {
	ntype := "u32"
	help := "Port numbers to include in the group"
	val := "u32:1-65535 ;\\\nA port number to include"
	expr := "syntax:expression: ($VAR(@) >= 1 && $VAR(@) <= 65535) ; \\\n    \"Must be a valid port number\""
	tr := true

	node, err := createTestNode(false, false,
		[]string{constructDefinition(
			&ntype,
			&help,
			&tr,
			&tr,
			[]string{val},
			nil,
			&expr,
		)},
	)

	if err != nil {
		t.Error(err.Error())
	}

	for _, err := range validateNode(node, tr, tr, ntype, help) {
		t.Error(err)
	}

	if len(node.Constraints) != 1 {
		t.Errorf("Single constraint should be added for port range, got %d", len(node.Constraints))
		t.FailNow()
	}
	if node.Constraints[0].MinBound != 1 {
		t.Errorf("Node constraint should have minimum bound of 1, got: %d", node.Constraints[0].MinBound)
	}
	if node.Constraints[0].MaxBound != 65535 {
		t.Errorf("Node constraint should have maximum bound of 65535, got: %d", node.Constraints[0].MaxBound)
	}
	if node.Constraints[0].FailureReason != "\"Must be a valid port number\"" {
		t.Errorf("Got incorrect node constraint failure reason, got: %s", node.Constraints[0].FailureReason)
	}
}

func TestPatternDefinition(t *testing.T) {
	ntype := "txt"
	help := "Zone name"
	pattern := "^[[:print:]]{1,18}$"
	reason := "\"Zone name must be 18 characters or less\""
	expr := fmt.Sprintf(`syntax:expression: pattern $VAR(@) "%s" ;
                %s`, pattern, reason)
	tr := true
	fa := false

	node, err := createTestNode(
		tr,
		fa,
		[]string{constructDefinition(
			&ntype,
			&help,
			&tr,
			&fa,
			[]string{},
			nil,
			&expr,
		)},
	)

	if err != nil {
		t.Error(err.Error())
	}

	for _, err := range validateNode(node, tr, fa, ntype, help) {
		t.Error(err)
	}

	if len(node.Constraints) != 1 {
		t.Errorf("Single constraint should be added for pattern, got %d", len(node.Constraints))
		t.FailNow()
	}

	if node.Constraints[0].Pattern != pattern {
		t.Errorf("Pattern was incorrect, got '%s'", node.Constraints[0].Pattern)
	}

	if node.Constraints[0].FailureReason != reason {
		t.Errorf("Failure reason was incorrect, got '%s'", node.Constraints[0].FailureReason)
	}

}
