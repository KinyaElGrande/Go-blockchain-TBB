package main

import (
	"fmt"
	"os"
	"time"

	"github.com/KinyaElGrande/TBB/database"
)

func main() {
	state, err := database.NewStateFromDisk()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer state.Close()

	block0 := database.NewBlock(
		database.Hash{},
		uint64(time.Now().Unix()),
		[]database.Transaction{
			database.NewTransaction("kinya", "elgrande", 300, ""),
			database.NewTransaction("kinya", "elgrande", 100, "reward"),
		},
	)

	state.AddBlock(block0)
	// block0Hash, _ := state.Persist()

	state.Persist()
}
