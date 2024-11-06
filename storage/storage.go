package storage

import (
	"encoding/binary"
	"encoding/json"
	"os"
	"os/user"
	"path/filepath"

	"github.com/getsavvyinc/savvy-cli/client"
)

const defaultLocalDBDir = "savvy"
const defaultDBFilename = "savvy.local"

var defaultLocalDBPath = filepath.Join(defaultLocalDBDir, defaultDBFilename)

func Write(store map[string]*client.Runbook) error {
	// Write the store to disk
	data, err := json.Marshal(store)
	if err != nil {
		return err
	}

	f, err := openStore()
	if err != nil {
		return err
	}
	defer f.Close()

	if err := f.Truncate(0); err != nil {
		return err
	}

	if err := binary.Write(f, binary.LittleEndian, data); err != nil {
		return err
	}
	return nil
}

func Read() (map[string]*client.Runbook, error) {
	// Read the store from disk
	f, err := openStore()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	data := make([]byte, stat.Size())

	if err := binary.Read(f, binary.LittleEndian, data); err != nil {
		return nil, err
	}

	store := make(map[string]*client.Runbook)
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}
	return store, nil
}

func openStore() (*os.File, error) {
	u, err := user.Current()
	if err != nil {
		return nil, err
	}
	homeDir := u.HomeDir

	storageFile := filepath.Join(homeDir, defaultLocalDBDir, defaultDBFilename)

	return os.OpenFile(storageFile, os.O_RDWR|os.O_CREATE, 0666)
}
