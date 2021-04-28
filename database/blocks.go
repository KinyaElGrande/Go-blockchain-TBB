package database

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type Hash [32]byte

func (h Hash) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString(h[:])), nil
}

func (h *Hash) UnmarshalText(data []byte) error {
	_, err := hex.Decode(h[:], data)
	return err
}

type Block struct {
	Header BlockHeader   `json:"header"`
	TXs    []Transaction `json:"payload"`
}

type BlockHeader struct {
	Parent Hash   `json:"parent"`
	Time   uint64 `json:"time"`
}

type BlockFs struct {
	Key   Hash  `json:"hash"`
	Value Block `json:"block"`
}

func (b Block) Hash() (Hash, error) {
	blockJson, err := json.Marshal(b)
	if err != nil {
		return Hash{}, nil
	}

	return sha256.Sum256(blockJson), nil
}

func NewBlock(parent Hash, time uint64, txs []Transaction) Block {
	return Block{BlockHeader{parent, time}, txs}
}

