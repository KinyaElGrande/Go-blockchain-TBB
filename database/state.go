package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"time"
)

// Holds all user balances and transaction events
type State struct {
	Balances  map[Account]uint
	txMempool []Transaction

	dbFile          *os.File
	latestBlockHash Hash
	latestBlock     Block
	hasGenesisBlock bool
}

func (s *State) LatestBlockHash() Hash {
	return s.latestBlockHash
}

func (s *State) LatestBlock() Block {
	return s.latestBlock
}

func (s *State) NextBlockNumber() uint64 {
	if !s.hasGenesisBlock {
		return uint64(0)
	}
	return s.latestBlock.Header.Height + 1
}

func (s *State) AddBlock(b Block) (Hash, error) {
	pendingState := s.copy()

	err := applyBlock(b, pendingState)
	if err != nil {
		return Hash{}, nil
	}

	blockHash, err := b.Hash()
	if err != nil {
		return Hash{}, nil
	}

	blockFs := BlockFs{blockHash, b}

	blockFsJson, err := json.Marshal(blockFs)
	if err != nil {
		return Hash{}, nil
	}

	fmt.Printf("Persisting new block to disk: \n")
	fmt.Printf("\t%s\n", blockFsJson)

	_, err = s.dbFile.Write(append(blockFsJson, '\n'))
	if err != nil {
		return Hash{}, nil
	}

	s.latestBlockHash = blockHash
	s.latestBlock = b
	// Reset the mempool
	s.txMempool = []Transaction{}

	return s.latestBlockHash, nil
}

func (s *State) copy() State {
	c := State{}
	c.latestBlock = s.latestBlock
	c.latestBlockHash = s.latestBlockHash
	c.txMempool = make([]Transaction, len(s.txMempool))
	c.Balances = make(map[Account]uint)

	for acc, balance := range s.Balances {
		c.Balances[acc] = balance
	}

	for _, tx := range s.txMempool {
		c.txMempool = append(c.txMempool, tx)
	}

	return c
}

func (s *State) AddTx(tx Transaction) error {
	if err := s.apply(tx); err != nil {
		return err
	}

	s.txMempool = append(s.txMempool, tx)

	return nil
}

func NewStateFromDisk(dataDir string) (*State, error) {
	err := initDataDirIfNotExists(dataDir)
	if err != nil {
		return nil, err
	}

	gen, err := loadGenesis(getGenesisJsonFilePath(dataDir))
	if err != nil {
		return nil, err
	}

	balances := make(map[Account]uint)
	for account, balance := range gen.Balances {
		balances[account] = balance
	}

	f, err := os.OpenFile(getBlocksDBFilePath(dataDir), os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(f)
	state := &State{balances, make([]Transaction, 0), f, Hash{}, Block{}, false}

	// iterate through the block DB file line by line
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		blockFsJson := scanner.Bytes()

		if len(blockFsJson) == 0 {
			break
		}

		var blockFs BlockFs
		err = json.Unmarshal(blockFsJson, &blockFs)
		if err != nil {
			return nil, err
		}

		err = state.applyBlock(blockFs.Value)
		if err != nil {
			return nil, err
		}

		state.latestBlockHash = blockFs.Key
		state.latestBlock = blockFs.Value
	}

	return state, nil
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
func (s *State) Persist() (Hash, error) {
	// Create a new Block with only the new Transactions
	block := NewBlock(
		s.latestBlockHash,
		s.latestBlock.Header.Height+1, // increases Block Height
		uint64(time.Now().Unix()),
		s.txMempool,
	)

	blockHash, err := block.Hash()
	if err != nil {
		return Hash{}, err
	}

	blockFs := BlockFs{blockHash, block}

	// Encode the block into a JSON string
	blockFsJson, err := json.Marshal(blockFs)
	if err != nil {
		return Hash{}, err
	}

	fmt.Printf("Persisting new Block to disk:\n")
	fmt.Printf("\t%s\n", blockFsJson)

	// Write the new Block into the DB file on a new line
	if _, err = s.dbFile.Write(append(blockFsJson, '\n')); err != nil {
		return Hash{}, err
	}

	s.latestBlockHash = blockHash
	s.latestBlock = block
	// Reset the mempool
	s.txMempool = []Transaction{}

	return s.latestBlockHash, nil
}

// Close Closes the DB file
func (s *State) Close() {
	s.dbFile.Close()
}

// applyBlock validates block meta + payload
func applyBlock(b Block, s State) error {
	nextExpectedBlockHeight := s.latestBlock.Header.Height + 1

	if s.hasGenesisBlock && b.Header.Height != nextExpectedBlockHeight {
		return fmt.Errorf("Next expected block must be '%d' and not '%d'", nextExpectedBlockHeight, b.Header.Height)
	}

	// validate the incoming block parent hash equals the latest known hash
	if s.hasGenesisBlock && s.latestBlock.Header.Height > 0 && !reflect.DeepEqual(b.Header.Parent, s.latestBlockHash) {
		return fmt.Errorf("Next block parent hash must be '%x' and not '%x' ", s.latestBlockHash, b.Header.Parent)
	}

	return applyTXs(b.TXs, &s)
}

func applyTXs(txs []Transaction, s *State) error {
	for _, tx := range txs {
		err := applyTx(tx, s)
		if err != nil {
			return err
		}
	}

	return nil
}

// applyTx replays transactions to verify balances
func applyTx(tx Transaction, s *State) error {
	if tx.IsReward() {
		s.Balances[tx.To] += tx.Value
		return nil
	}

	if tx.Value > s.Balances[tx.From] {
		return fmt.Errorf("wrong TX. Sender '%s' balance is %d TBB. Tx cost is %d",
			tx.From,
			s.Balances[tx.From],
			tx.Value,
		)
	}

	s.Balances[tx.From] -= tx.Value
	s.Balances[tx.To] += tx.Value

	return nil
}
