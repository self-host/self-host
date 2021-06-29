package configdir

import (
	"os"
	"path/filepath"
)

var defaultFileMode = os.FileMode(0755)

// SystemConfig returns the system-wide configuration paths
func SystemConfig(application string) []string {
	p := make([]string, 0)
	for _, path := range findSystemPaths() {
		p = append(p, filepath.Join(path, application))
	}
	return p
}

// LocalConfig returns the local user configuration path
func LocalConfig(application string) []string {
	p := make([]string, 0)
	for _, path := range findLocalPaths() {
		p = append(p, filepath.Join(path, application))
	}
	return p
}

// MakePath ensures that the full path you wanted exists,
// including application-specific or vendor components.
func MakePath(path string) error {
	err := os.MkdirAll(path, defaultFileMode)
	if err != nil {
		return err
	}

	return nil
}
