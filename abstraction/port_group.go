package abstraction

import (
	"fmt"
	"os"
	"path"
	"regexp"

	"gopkg.in/yaml.v3"

	"github.com/ammesonb/ubiquiti-config-generator/logger"
)

// LoadPortGroups will look at all yaml files in the given path and return a list of port groups
func LoadPortGroups(portGroupsPath string) ([]PortGroup, error) {
	var portGroups []PortGroup

	entries, err := os.ReadDir(portGroupsPath)
	if err != nil {
		return portGroups, fmt.Errorf(
			"failed to read port group path '%s': %v", portGroupsPath, err,
		)
	}

	for _, entry := range entries {
		// Only want yaml files
		if entry.IsDir() || !entry.Type().IsRegular() {
			continue
		}

		fileNameRegex := regexp.MustCompile(`^(.*)\.ya?ml`)
		if fileNameRegex.MatchString(entry.Name()) {
			groupName := fileNameRegex.FindStringSubmatch(entry.Name())[1]

			group, err := makePortGroup(path.Join(portGroupsPath, entry.Name()), groupName)
			if err != nil {
				return []PortGroup{}, err
			} else if len(group.Ports) > 0 {
				portGroups = append(portGroups, *group)
			} else {
				logger.DefaultLogger().Warnf("no ports detected for port group '%s'", groupName)
			}
		}
	}

	return portGroups, nil
}

func makePortGroup(filepath string, groupName string) (*PortGroup, error) {
	groupData, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read port group path '%s': %v", filepath, err,
		)
	}
	group := PortGroup{Name: groupName}
	if err = yaml.Unmarshal(groupData, &group); err != nil {
		return nil, fmt.Errorf(
			"failed to parse port group in '%s': %v", filepath, err,
		)
	}

	return &group, nil
}
