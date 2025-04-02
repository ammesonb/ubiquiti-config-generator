package config

import (
	"path"
	"strings"

	"github.com/ammesonb/ubiquiti-config-generator/internal/errors"
	"github.com/ammesonb/ubiquiti-config-generator/mocks"
	"github.com/ammesonb/ubiquiti-config-generator/services/configuration"
)

var errReadDir = "failed to read directory %s"

// EnumerateConfigFiles will return a full list of all configuration files in a given path that the device requires
func EnumerateConfigFiles(fs mocks.FsWrapper, device *configuration.DeviceConfig, pathRoot string) ([]string, []error) {
	files := make([]string, 0)
	errs := make([]error, 0)

	// Start with getting all entries in this path
	// Since directories can be specified in configuration too, need to consider all of them
	entries, err := fs.ReadDir(pathRoot)
	if err != nil {
		errs = append(errs, errors.ErrWithCtxParent(errReadDir, pathRoot, err))
		return files, errs
	}

	for _, entry := range entries {
		fullPath := path.Join(pathRoot, entry.Name())
		// TODO: add bypass to `entryInConfig` if whole directory should be included
		// If file not in config, skip it
		if !entry.IsDir() && !entryInConfig(device.ConfigFiles, fullPath) {
			continue
		} else if entry.IsDir() && !dirInConfig(device.ConfigFiles, fullPath) {
			continue
		}

		if entry.IsDir() {
			// For directory, recurse and get any nested files
			children, errs := EnumerateConfigFiles(fs, device, fullPath)
			files = append(files, children...)
			errs = append(errs, errs...)
		} else {
			// Otherwise simply append this file
			files = append(files, fullPath)
		}
	}

	return files, errs
}

func entryInConfig(paths []string, path string) bool {
	for _, p := range paths {
		// Must either be a direct path match for a file/dir, or if it is a directory
		// then the file must be inside the configured directory
		if p == path || (strings.HasPrefix(path, p) && strings.HasSuffix(p, "/")) {
			return true
		}
	}

	return false
}

func dirInConfig(paths []string, path string) bool {
	for _, p := range paths {
		// If exact match or file path starts with the configured one
		if p == path || strings.HasPrefix(p, path) {
			return true
		}
	}

	return false
}

func DeviceFilesChanged(device *configuration.DeviceConfig, changedFiles []string) bool {
	// For each changed file, check if it is a dependency
	for _, file := range changedFiles {
		if entryInConfig(device.ConfigFiles, file) {
			return true
		}
	}

	return false
}
