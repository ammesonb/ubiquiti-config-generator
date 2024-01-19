package abstraction

import (
	"fmt"
	"os"
	"path"
	"regexp"

	"gopkg.in/yaml.v3"
)

var (
	ErrPortGroupEmpty = "no ports specified for port group"
	ErrParsePortYAML  = "failed to parse port group in"
)

// LoadPortGroups will look at all yaml files in the given path and return a list of port groups
func LoadPortGroups(portGroupsPath string) ([]PortGroup, []error) {
	var portGroups []PortGroup
	errors := make([]error, 0)

	entries, err := os.ReadDir(portGroupsPath)
	if err != nil {
		return portGroups,
			[]error{fmt.Errorf(
				"failed to read port group path '%s': %v", portGroupsPath, err,
			)}
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
				errors = append(errors, err)
			} else if len(group.Ports) > 0 {
				portGroups = append(portGroups, *group)
			} else {
				errors = append(errors, fmt.Errorf("%s %s", ErrPortGroupEmpty, groupName))
			}
		}
	}

	return portGroups, errors
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
			"%s '%s': %v", ErrParsePortYAML, filepath, err,
		)
	}

	return &group, nil
}
