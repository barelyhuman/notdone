package lib

import (
	"os"
	"path/filepath"
)

func GetBaseDirectory() string {
	userHome, err := os.UserHomeDir()
	Bail(err)
	basePath := filepath.Join(userHome, ".notdone")
	return basePath
}
