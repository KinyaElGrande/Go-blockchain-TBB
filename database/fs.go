package database

import (
	"os"
	"path/filepath"
)

func getDatabaseDirPath(dataDir string) string {
	return filepath.Join(dataDir, "database")
}

func getGenesisJsonFilePath(dataDir string) string {
	return filepath.Join(getDatabaseDirPath(dataDir), "genesis.json")
}

func getBlocksDBFilePath(dataDir string) string {
	return filepath.Join(getDatabaseDirPath(dataDir), "blocks.db")
}

func initDataDirIfNotExists(dataDir string) error {
	if fileExists(getGenesisJsonFilePath(dataDir)) {
		return nil
	}

	dbDir := getDatabaseDirPath(dataDir)
	if err := os.MkdirAll(dbDir, os.ModePerm); err != nil {
		return err
	}

	gen := getGenesisJsonFilePath(dataDir)
	if err := writeGenesisToDisk(gen); err != nil {
		return err
	}

	blocks := getBlocksDBFilePath(dataDir)
	if err := writeEmptyBlocksDbToDisk(blocks); err != nil {
		return err
	}

	return nil
}

// fileExists Checks if a file exists
// if it doesn't it returns false
func fileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}

func writeEmptyBlocksDbToDisk(path string) error {
	return os.WriteFile(path, []byte(""), os.ModePerm)
}
