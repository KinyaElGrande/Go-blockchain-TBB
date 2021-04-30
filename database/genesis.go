package database

import (
	"encoding/json"
	"os"
)

var genesisJson = `
{
    "genesis_time": "2021-03-18T00:00:00.000000000Z",
    "chain_id": "tbb-ledger",
    "balances": {
        "kinya": 10000000
    }
}
`

type Genesis struct {
	Balances map[Account]uint `json:"balances"`
}

// Opens genesis file path
func loadGenesis(path string) (Genesis, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Genesis{}, err
	}

	var loadedGenesis Genesis
	err = json.Unmarshal(content, &loadedGenesis)
	if err != nil {
		return Genesis{}, err
	}

	return loadedGenesis, nil
}

func writeGenesisToDisk(path string) error {
	return os.WriteFile(path, []byte(genesisJson), 0644)
}
