//go:build !go1.20
// +build !go1.20

package net

import (
	"fmt"
	"time"

	"github.com/gochain/gochain/pkg/block"
)

// Network is a no-op stub for Go versions < 1.20
// to avoid importing dependencies that require newer Go.
type Network struct{}

type PeerInfo struct{}

type NetworkConfig struct {
	ListenPort        int
	BootstrapPeers    []string
	EnableMDNS        bool
	EnableRelay       bool
	MaxPeers          int
	ConnectionTimeout time.Duration
}

func DefaultNetworkConfig() *NetworkConfig {
	return &NetworkConfig{
		ListenPort:        0,
		BootstrapPeers:    []string{},
		EnableMDNS:        true,
		EnableRelay:       false,
		MaxPeers:          50,
		ConnectionTimeout: 30 * time.Second,
	}
}

func NewNetwork(config *NetworkConfig) (*Network, error) { return &Network{}, nil }

func (n *Network) SubscribeToBlocks(handler func(*block.Block)) error { return nil }

func (n *Network) SubscribeToTransactions(handler func(*block.Transaction)) error { return nil }

func (n *Network) GetPeerCount() int { return 0 }

func (n *Network) Close() error { return nil }

func (n *Network) String() string { return fmt.Sprintf("Network{Peers: %d}", 0) }
