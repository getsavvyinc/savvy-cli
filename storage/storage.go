package storage

import (
	"encoding/gob"
	"os"
	"os/user"
	"path/filepath"

	"github.com/getsavvyinc/savvy-cli/client"
)

const defaultLocalDBDir = "savvy"
const defaultDBFilename = "savvy.local"

var defaultLocalDBPath = filepath.Join(defaultLocalDBDir, defaultDBFilename)

func Write(store map[string]*client.Runbook) error {
	f, err := openStore()
	if err != nil {
		return err
	}
	defer f.Close()

	if err := f.Truncate(0); err != nil {
		return err
	}

	encoder := gob.NewEncoder(f)
	if err := encoder.Encode(store); err != nil {
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

	if _, err := f.Stat(); err != nil {
		return nil, err
	}

	store := make(map[string]*client.Runbook)

	decoder := gob.NewDecoder(f)
	if err := decoder.Decode(&store); err != nil {
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
