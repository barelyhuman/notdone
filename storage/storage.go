package storage

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/barelyhuman/notdone/lib"
)

const filename = "./storage.json"

type Storage struct {
	rootPath    string
	storagePath string
}

func New() (s *Storage) {
	s = &Storage{}
	userHome, err := os.UserHomeDir()
	lib.Bail(err)
	basePath := filepath.Join(userHome, ".notdone")
	lib.Bail(os.MkdirAll(basePath, os.ModePerm))
	s.rootPath = basePath
	s.storagePath = filepath.Join(basePath, "storage.json")

	_, err = os.Stat(s.storagePath)

	if os.IsNotExist(err) {
		os.WriteFile(s.storagePath, []byte("{}"), os.ModePerm)
	} else {
		lib.Bail(err)
	}

	return
}

func (s *Storage) Load(v interface{}) {
	file, err := os.ReadFile(s.storagePath)
	lib.Bail(err)
	err = json.Unmarshal(file, &v)
	lib.Bail(err)
}

func (s *Storage) Save(v interface{}) {
	bytes, err := json.Marshal(v)
	lib.Bail(err)
	os.WriteFile(s.storagePath, bytes, os.ModePerm)
}
