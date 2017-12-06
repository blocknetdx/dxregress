package util

import "os"

// FileExists returns true if the path exists.
func FileExists(fPath string) bool {
	if _, err := os.Stat(fPath); !os.IsNotExist(err) {
		return true
	}
	return false
}
