package config

import (
	"github.com/ammesonb/ubiquiti-config-generator/utils"
	"os"
	"path"
	"strings"
)

// EnumerateConfigFiles will return a full list of all configuration files in a given path that the device requires
func EnumerateConfigFiles(device *DeviceConfig, pathRoot string) ([]string, []error) {
	files := make([]string, 0)
	errors := make([]error, 0)

	// Start with getting all entries in this path
	// Since directories can be specified in configuration too, need to consider all of them
	entries, err := os.ReadDir(pathRoot)
	if err != nil {
		errors = append(errors, utils.ErrWithCtxParent("failed to read directory %s", pathRoot, err))
		return files, errors
	}

	for _, entry := range entries {
		fullPath := path.Join(pathRoot, entry.Name())
		// If file or directory not in config, skip it
		if !entryInConfig(device.ConfigFiles, fullPath) {
			continue
		}

		info, err := os.Stat(fullPath)
		if err != nil {
			errors = append(errors, utils.ErrWithCtxParent("failed to stat entry %s", fullPath, err))
			continue
		}

		if info.IsDir() {
			// For directory, recurse and get any nested files
			children, errs := EnumerateConfigFiles(device, fullPath)
			files = append(files, children...)
			errors = append(errors, errs...)
		} else {
			// Otherwise simply append this file
			files = append(files, fullPath)
		}
	}

	return files, errors

}

func entryInConfig(paths []string, path string) bool {
	for _, p := range paths {
		if strings.HasPrefix(p, path) {
			return true
		}
	}

	return false
}

func DeviceFilesChanged(device *DeviceConfig, changedFiles []string) bool {
	// For each changed file, check if it is a dependency
	for _, file := range changedFiles {
		for _, dep := range device.ConfigFiles {
			// It is a dependency if the file is an exact match to one the device requires, or
			// if it is inside a required directory
			if strings.HasPrefix(file, dep) {
				return true
			}
		}
	}

	return false
}
