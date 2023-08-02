package abstraction

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"regexp"

	"github.com/ammesonb/ubiquiti-config-generator/logger"
	"gopkg.in/yaml.v3"
)

// PortGroup contains a list of ports for a firewall group
type PortGroup struct {
	Name        string
	Description string `yaml:"description"`
	Ports       []int  `yaml:"ports"`
}

// Equals returns true if another port group is identical to this one
func (group *PortGroup) Equals(other *PortGroup) bool {
	return group.Name == other.Name &&
		group.Description == other.Description &&
		reflect.DeepEqual(group.Ports, other.Ports)
}

// LoadPortGroups will look at all yaml files in the given path and return a list of port groups
func LoadPortGroups(portGroupsPath string) ([]PortGroup, error) {
	ports := []PortGroup{}
	entries, err := os.ReadDir(portGroupsPath)
	if err != nil {
		return ports, fmt.Errorf(
			"Failed to read port group path '%s': %v", portGroupsPath, err,
		)
	}

	fileNameRegex := regexp.MustCompile(`^(.*)\.ya?ml`)
	for _, entry := range entries {
		// Only want yaml files
		if entry.IsDir() || !entry.Type().IsRegular() {
			continue
		}

		if fileNameRegex.MatchString(entry.Name()) {
			groupData, err := os.ReadFile(path.Join(portGroupsPath, entry.Name()))
			if err != nil {
				return []PortGroup{}, fmt.Errorf(
					"Failed to read port group path '%s': %v", path.Join(portGroupsPath, entry.Name()), err,
				)
			}
			groupName := fileNameRegex.FindStringSubmatch(entry.Name())[1]
			group := PortGroup{Name: groupName}
			if err = yaml.Unmarshal(groupData, &group); err != nil {
				return []PortGroup{}, fmt.Errorf(
					"Failed to parse port group in '%s': %v", path.Join(portGroupsPath, entry.Name()), err,
				)
			}

			if len(group.Ports) > 0 {
				ports = append(ports, group)
			} else {
				logger.DefaultLogger().Warnf("No ports detected for port group '%s'", groupName)
			}
		}
	}

	return ports, nil
}
