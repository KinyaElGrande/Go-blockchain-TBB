package node

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/KinyaElGrande/TBB/database"
)

const httpPort = 8000

type ErrorRes struct {
	Error string `json:"error"`
}

type BalancesRes struct {
	Hash     database.Hash             `json:"block_hash"`
	Balances map[database.Account]uint `json:"balances"`
}

type StatusRes struct {
	Hash       database.Hash       `json:"block_hash"`
	Height     uint64              `json:"block_height"`
	KnownPeers map[string]PeerNode `json:"known_peers"`
}

type TransactionAddReq struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Value uint   `json:"value"`
	Data  string `json:"data"`
}

type TxAddRes struct {
	Hash database.Hash `json:"block_hash"`
}

func (n *Node) Run() error {
	ctx := context.Background()
	fmt.Println(fmt.Sprintf("Listening on HTTP port: %d", httpPort))

	state, err := database.NewStateFromDisk(n.dataDir)
	if err != nil {
		return err
	}
	// defer state.Close()

	n.state = state

	go n.sync(ctx)

	http.HandleFunc("/balances/list", func(w http.ResponseWriter, r *http.Request) {
		listBalancesHandler(w, r, state)
	})

	http.HandleFunc(endpointStatus, func(w http.ResponseWriter, r *http.Request) {
		statusHandler(w, r, n)
	})

	http.HandleFunc("/transaction/add", func(w http.ResponseWriter, r *http.Request) {
		txAddHandler(w, r, state)
	})

	http.ListenAndServe(fmt.Sprintf(":%d", httpPort), nil)

	return nil
}

// listBalancesHandler lists state balances
func listBalancesHandler(w http.ResponseWriter, r *http.Request, state *database.State) {
	writeRes(w, BalancesRes{state.LatestBlockHash(), state.Balances})
}

func writeRes(w http.ResponseWriter, content interface{}) {
	contentJSON, err := json.Marshal(content)
	if err != nil {
		writeErrRes(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(contentJSON)
}

func writeErrRes(w http.ResponseWriter, err error) {
	jsonErrRes, _ := json.Marshal(ErrorRes{err.Error()})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(jsonErrRes)
}

// txAddHandler adds a new transaction to the blockchain
func txAddHandler(w http.ResponseWriter, r *http.Request, state *database.State) {
	req := TransactionAddReq{}

	// parse the POST request body
	err := readReq(r, &req)
	if err != nil {
		writeErrRes(w, err)
		return
	}

	tx := database.NewTransaction(database.NewAccount(req.From), database.NewAccount(req.To), req.Value, req.Data)

	err = state.AddTx(tx)
	if err != nil {
		writeErrRes(w, err)
		return
	}

	// Flush the mempool
	hash, err := state.Persist()
	if err != nil {
		writeErrRes(w, err)
		return
	}

	writeRes(w, TxAddRes{hash})
}

func readReq(r *http.Request, reqBody interface{}) error {
	reqBodyJSON, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("Unable to read request body: %s", err.Error())
	}
	defer r.Body.Close()

	err = json.Unmarshal(reqBodyJSON, reqBody)
	if err != nil {
		return fmt.Errorf("Unable to unmarshal request body: %s", err.Error())
	}

	return nil
}

func readRes(r *http.Response, reqBody interface{}) error {
	reqBodyJson, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("Unabale to read response body. %s", err.Error())
	}
	defer r.Body.Close()

	err = json.Unmarshal(reqBodyJson, reqBody)
	if err != nil {
		return fmt.Errorf("Unabale to unmarshal response body. %s", err.Error())
	}

	return nil
}

func statusHandler(w http.ResponseWriter, r *http.Request, node *Node) {
	res := StatusRes{
		Hash:       node.state.LatestBlockHash(),
		Height:     node.state.LatestBlock().Header.Height,
		KnownPeers: node.knownPeers,
	}

	writeRes(w, res)
}
