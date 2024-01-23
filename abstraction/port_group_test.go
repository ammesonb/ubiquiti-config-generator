package abstraction

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadPortGroups(t *testing.T) {
	groups, errs := LoadPortGroups("./test-files/port-groups")
	assert.Len(t, errs, 2, "Should have two errors")
	assert.ErrorContains(t, errs[0], ErrPortGroupEmpty+" blank", "Blank port group should throw error")
	assert.ErrorContains(t, errs[1], ErrParsePortYAML, "Invalid yaml should throw error")

	assert.Len(t, groups, 2, "Expected 2 port groups loaded")

	serverPorts := PortGroup{
		Name:        "server-ports",
		Description: "Ports to forward to the server",
		Ports:       []int{80, 443, 22},
	}
	webPorts := PortGroup{
		Name:        "web",
		Description: "Website ports",
		Ports:       []int{80, 443},
	}

	if !reflect.DeepEqual(serverPorts, groups[0]) {
		t.Errorf(
			"Group 0 did not match server ports, got %#v, expected %#v",
			groups[0],
			serverPorts,
		)
	}

	if !reflect.DeepEqual(webPorts, groups[1]) {
		t.Errorf(
			"Group 1 did not match web ports, got %#v, expected %#v",
			groups[1],
			webPorts,
		)
	}
}

func TestNonexistentPath(t *testing.T) {
	groups, errs := LoadPortGroups("/nonexistent")

	assert.Len(t, groups, 0, "Nonexistent path should not find groups")
	assert.Len(t, errs, 1, "Single error should have been returned")

	group, err := makePortGroup("/nonexistent", "group")
	assert.Nil(t, group, "Group should be null if file does not exist")
	assert.ErrorContains(t, err, errFailedReadPortGroup)
}
