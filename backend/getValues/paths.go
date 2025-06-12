package getvalues

import (
	"os"
	"path/filepath"
)

func FindRootPath() string {
	exePath, err := os.Executable()
	if err != nil {
		return ""
	}
	dir := filepath.Dir(exePath)
	parentDir := filepath.Dir(dir)
	return parentDir
}
