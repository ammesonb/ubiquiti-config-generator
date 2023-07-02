package vyos

import (
	"testing"
)

func TestLineDetection(t *testing.T) {
	comment := "   /* This is a comment */"
	tagNodeOpen := "    name some-firewall {"
	tagNodeClose := "    }"
	normalNodeOpen := "services {"
	value := "    port 8080"
	quotedValue := `    description "Firewall rule {20}"`

	t.Run("Comment", func(t *testing.T) {
		verifyLine(t, comment, true, false, false, false)
	})

	t.Run("Open Tag Node", func(t *testing.T) {
		verifyLine(t, tagNodeOpen, false, true, false, false)

		description, name := splitNameFromValue(tagNodeOpen)
		if description != "name" {
			t.Errorf(
				"Incorrect description for tag node open: '%s'", description,
			)
		}
		if name != "some-firewall" {
			t.Errorf(
				"Incorrect node name for tag node open: '%s'", name,
			)
		}
	})

	t.Run("Close Tag Node", func(t *testing.T) {
		verifyLine(t, tagNodeClose, false, false, true, false)
	})

	t.Run("Open Normal Node", func(t *testing.T) {
		verifyLine(t, normalNodeOpen, false, true, false, false)
	})

	t.Run("Value", func(t *testing.T) {
		verifyLine(t, value, false, false, false, true)

		nodeName, nodeValue := splitNameFromValue(value)
		if nodeName != "port" {
			t.Errorf(
				"Incorrect name for node: '%s'", nodeName,
			)
		}
		if nodeValue != "8080" {
			t.Errorf(
				"Incorrect value for node: '%s'", nodeValue,
			)
		}
	})

	t.Run("Quoted value", func(t *testing.T) {
		verifyLine(t, quotedValue, false, false, false, true)

	})
}

func verifyLine(t *testing.T, line string, isComment bool, isNode bool, closesNode bool, hasValue bool) {
	if lineIsComment(line) != isComment {
		negate := ""
		if !isComment {
			negate = "not "
		}
		t.Errorf("Line should be %sdetected as comment", negate)
	}
	if lineCreatesNode(line) != isNode {
		negate := ""
		if !isNode {
			negate = "not "
		}
		t.Errorf("Line should %screate a node", negate)
	}
	if lineEndsNode(line) != closesNode {
		negate := ""
		if !closesNode {
			negate = "not "
		}
		t.Errorf("Line should %sclose a node", negate)
	}
	if lineHasValue(line) != hasValue {
		negate := ""
		if !closesNode {
			negate = "not "
		}
		t.Errorf("Line should %shave a value", negate)
	}
}

func TestSplitName(t *testing.T) {
	tagNodeOpen := "    name some-firewall {"
	quotedValue := `    description "Firewall rule {20}"`

	t.Run("Open Tag Node", func(t *testing.T) {
		description, name := splitNameFromValue(tagNodeOpen)
		if description != "name" {
			t.Errorf(
				"Incorrect description for tag node open: '%s'", description,
			)
		}
		if name != "some-firewall" {
			t.Errorf(
				"Incorrect node name for tag node open: '%s'", name,
			)
		}
	})

	t.Run("Quoted value", func(t *testing.T) {
		nodeName, nodeValue := splitNameFromValue(quotedValue)
		if nodeName != "description" {
			t.Errorf(
				"Incorrect name for quoted node: '%s'", nodeName,
			)
		}
		if nodeValue != "Firewall rule {20}" {
			t.Errorf(
				"Incorrect value for node: '%s'", nodeValue,
			)
		}
	})
}
