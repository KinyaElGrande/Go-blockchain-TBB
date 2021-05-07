package node

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// sync searches for new network's peeers for every 45 seconds
func (n *Node) sync(ctx context.Context) error {
	ticker := time.NewTicker(45 * time.Second)

	for {
		select {
		case <-ticker.C:
			fmt.Println("Searching for new Peers and Blocks")

			n.fetchNewBlocksAndPeers()
		case <-ctx.Done():
			ticker.Stop()
		}
	}
}

func (n *Node) fetchNewBlocksAndPeers() {
	for _, knownPeer := range n.knownPeers {
		status, err := queryPeerStatus(knownPeer)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			continue
		}

		localBlockNumber := n.state.LatestBlock().Header.Height
		if localBlockNumber < status.Height {
			newBlockCount := status.Height - localBlockNumber

			fmt.Printf("Found %d new blocks from Peer %s\n", newBlockCount, knownPeer.IP)
		}

		for _, maybeNewPeer := range status.KnownPeers {
			_, isKnownPeer := n.knownPeers[maybeNewPeer.TcpAddress()]
			if !isKnownPeer {
				fmt.Printf("Found new Peer %s\n", knownPeer.TcpAddress())

				n.knownPeers[maybeNewPeer.TcpAddress()] = maybeNewPeer
			}
		}
	}
}

func queryPeerStatus(peer PeerNode) (StatusRes, error) {
	url := fmt.Sprintf("http://%s%s", peer.TcpAddress(), endpointStatus)
	res, err := http.Get(url)
	if err != nil {
		return StatusRes{}, nil
	}

	statusRes := StatusRes{}
	err = readRes(res, &statusRes)
	if err != nil {
		return StatusRes{}, nil
	}

	return statusRes, nil
}
