package abstraction

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadPortGroups(t *testing.T) {
	groups, err := LoadPortGroups("../sample_router_config/port-groups")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

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
	groups, err := LoadPortGroups("/nonexistent")

	assert.Len(t, groups, 0, "Nonexistent path should not find groups")
	assert.NotNil(t, err, "Error should have been returned")
}
