package database

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Holds all user balances and transaction events
type State struct {
	Balances  map[Account]uint
	txMempool []Transaction

	dbFile   *os.File
	snapshot Snapshot
}

type Snapshot [32]byte

func (s *State) LatestSnapshot() Snapshot {
	return s.snapshot
}

func NewStateFromDisk() (*State, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	genFilePath := filepath.Join(cwd, "database", "genesis.json")
	gen, err := loadGenesis(genFilePath)
	if err != nil {
		return nil, err
	}

	balances := make(map[Account]uint)
	for account, balance := range gen.Balances {
		balances[account] = balance
	}

	txDbFilePath := filepath.Join(cwd, "database", "transaction.db")
	f, err := os.OpenFile(txDbFilePath, os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(f)
	state := &State{balances, make([]Transaction, 0), f, Snapshot{}}

	// iterate through the transaction DB file line by line
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		// convert JSON encoded Transaction into an object
		var tx Transaction
		json.Unmarshal(scanner.Bytes(), &tx)

		if err := state.apply(tx); err != nil {
			return nil, err
		}
	}

	err = state.doSnapShot()
	if err !=nil {
		return nil, err
	}

	return state, nil
}

// apply Changes and validates the state
func (s *State) apply(tx Transaction) error {
	if tx.IsReward() {
		s.Balances[tx.To] += tx.Value
		return nil
	}

	if tx.Value > s.Balances[tx.From] {
		return fmt.Errorf("Insufficient Balance!")
	}

	s.Balances[tx.From] -= tx.Value
	s.Balances[tx.To] += tx.Value

	return nil
}

// Adding new transactions to the mempool
func (s *State) Add(tx Transaction) error {
	if err := s.apply(tx); err != nil {
		return err
	}

	s.txMempool = append(s.txMempool, tx)

	return nil
}

// Persisting and hashing the transactions to disk 
func (s *State) Persist() (Snapshot, error) {
	mempool := make([]Transaction, len(s.txMempool))
	copy(mempool, s.txMempool)

	for i := 0; i < len(mempool); i++ {
		txJson, err := json.Marshal(mempool[i])
		if err != nil {
			return Snapshot{}, err
		}

		fmt.Printf("Persisting new Transaction to disk: \n")
		fmt.Printf("\t%s\n", txJson)
		if _, err = s.dbFile.Write(append(txJson, '\n')); err != nil {
			return Snapshot{}, err
		}

		err = s.doSnapShot()
		if err != nil {
			return Snapshot{}, err
		}
		fmt.Printf("New DB Snapshot: %x\n", s.snapshot)

		// Remove the Transaction written to a file from the mempool
		s.txMempool = s.txMempool[1:]
	}

	return s.snapshot, nil
}

// Close Closes the DB file
func (s *State) Close() {
	s.dbFile.Close()
}

// doSnapShot records the contents of the entire blockchain ledger
func (s *State) doSnapShot() error {
	// Re-read the whole file from the first byte
	_, err := s.dbFile.Seek(0, 0)
	if err != nil {
		return err
	}

	txsData, err := ioutil.ReadAll(s.dbFile)
	if err != nil {
		return err
	}
	s.snapshot = sha256.Sum256(txsData)

	return nil
}
