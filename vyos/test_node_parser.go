package vyos

import "testing"

// TestAddOption uses basic configuration stubs to verify variables are set properly
func TestAddOption(t *testing.T) {
	n := &Node{
		Name:  "TestNode",
		IsTag: false,
		Multi: false,
	}
	helpValues := &[]string{}
	allowed := ""
	expression := ""

	t.Run("tag", func(t *testing.T) {
		addOption("tag", "", helpValues, &allowed, &expression, n)
		if !n.IsTag {
			t.Errorf("Tag flag not set on node")
		}
	})
	t.Run("multi", func(t *testing.T) {
		addOption("multi", "", helpValues, &allowed, &expression, n)
		if !n.Multi {
			t.Errorf("Multi flag not set on node")
		}
	})
	t.Run("type", func(t *testing.T) {
		addOption("type", "txt", helpValues, &allowed, &expression, n)
		if n.Type != "txt" {
			t.Errorf("Expected node type to be txt, got '%s'", n.Type)
		}
	})
	t.Run("help", func(t *testing.T) {
		addOption("help", "some description", helpValues, &allowed, &expression, n)
		if n.Help != "some description" {
			t.Errorf("Got unexpected node description: '%s'", n.Type)
		}
	})
	t.Run("val_help", func(t *testing.T) {
		addOption("val_help", "u32", helpValues, &allowed, &expression, n)
		if len(*helpValues) == 0 {
			t.Errorf("No value added to help descriptions")
		} else if len(*helpValues) > 0 {
			t.Errorf("Too many values added to help descriptions: %#v", *helpValues)
		} else if (*helpValues)[0] != "u32" {
			t.Errorf("Got unexpected help value: %s", (*helpValues)[0])
		}
	})
	t.Run("allowed", func(t *testing.T) {
		addOption("allowed", "cli-shell-api", helpValues, &allowed, &expression, n)
		if allowed != "cli-shell-api" {
			t.Errorf("Got unexpected allowed value: '%s'", allowed)
		}
	})
	t.Run("syntax", func(t *testing.T) {
		addOption("syntax:", "pattern $VAR(@)", helpValues, &allowed, &expression, n)
		if expression != "pattern $VAR(@)" {
			t.Errorf("Got unexpected expression value: '%s'", expression)
		}
	})
}
