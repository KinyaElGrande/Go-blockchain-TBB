package node

import (
	"fmt"

	"github.com/KinyaElGrande/TBB/database"
)

type Node struct {
	dataDir string
	port    uint64

	state *database.State

	knownPeers map[string]PeerNode
}

type PeerNode struct {
	IP          string `json:"ip"`
	Port        uint64 `json:"port"`
	IsBootstrap bool   `json:"is_bootstrap"`

	IsActive bool `json:"is_active"`
}

func New(dataDir string, port uint64, bootstrap PeerNode) *Node {
	// initialize a new map with only one known peer
	knownPeers := make(map[string]PeerNode)
	knownPeers[bootstrap.TcpAddress()] = bootstrap
	return &Node{
		dataDir:    dataDir,
		port:       port,
		knownPeers: knownPeers,
	}
}

func NewPeerNode(ip string, port uint64, isBootstrap bool, isActive bool) PeerNode {
	return PeerNode{ip, port, isBootstrap, isActive}
}

func (n *Node) RemovePeer(peer PeerNode) {
	delete(n.knownPeers, peer.TcpAddress())
}

func (pn PeerNode) TcpAddress() string {
	return fmt.Sprintf("%s:%d", pn.IP, pn.Port)
}
