package storage

import (
	"encoding/gob"
	"os"
	"path/filepath"

	"github.com/getsavvyinc/savvy-cli/client"
	"github.com/getsavvyinc/savvy-cli/config"
)

const defaultDBFilename = "savvy.local"

var defaultLocalDBPath = filepath.Join(config.DefaultConfigDir, defaultDBFilename)

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
	return os.OpenFile(defaultLocalDBPath, os.O_RDWR|os.O_CREATE, 0666)
}
