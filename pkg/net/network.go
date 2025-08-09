package net

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gochain/gochain/pkg/chain"
	"github.com/gochain/gochain/pkg/mempool"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/libp2p/go-libp2p/p2p/transport/websocket"
	"github.com/multiformats/go-multiaddr"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-pubsub"
)

// Notifiee methods for network.Notifiee interface
func (n *Network) Connected(net network.Network, conn network.Conn) {
	fmt.Printf("Connected to: %s/p2p/%s\n", conn.RemoteMultiaddr(), conn.RemotePeer().String())
}

func (n *Network) Disconnected(net network.Network, conn network.Conn) {
	fmt.Printf("Disconnected from: %s/p2p/%s\n", conn.RemoteMultiaddr(), conn.RemotePeer().String())
}

func (n *Network) OpenedStream(net network.Network, s network.Stream) {
	// fmt.Printf("Opened stream from: %s\n", s.Conn().RemotePeer().String())
}

func (n *Network) ClosedStream(net network.Network, s network.Stream) {
	// fmt.Printf("Closed stream from: %s\n", s.Conn().RemotePeer().String())
}

func (n *Network) OpenedConn(net network.Network, conn network.Conn) {
	// fmt.Printf("Opened connection to: %s\n", conn.RemotePeer().String())
}

func (n *Network) ClosedConn(net network.Network, conn network.Conn) {
	// fmt.Printf("Closed connection to: %s\n", conn.RemotePeer().String())
}

func (n *Network) Listen(net network.Network, multiaddr multiaddr.Multiaddr) {
	// fmt.Printf("Network listening on: %s\n", multiaddr.String())
}

func (n *Network) ListenClose(net network.Network, multiaddr multiaddr.Multiaddr) {
	// fmt.Printf("Network stopped listening on: %s\n", multiaddr.String())
}

// HandlePeerFound is called when a new peer is found via mDNS
func (n *Network) HandlePeerFound(peerInfo peer.AddrInfo) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if _, found := n.peers[peerInfo.ID]; !found {
		fmt.Printf("Discovered new peer: %s\n", peerInfo.ID.String())
		n.peers[peerInfo.ID] = &PeerInfo{
			ID:        peerInfo.ID,
			Addrs:     peerInfo.Addrs,
			Connected: time.Now(),
			LastSeen:  time.Now(),
		}
		// Attempt to connect to the discovered peer
		go func() {
			if err := n.host.Connect(n.ctx, peerInfo); err != nil {
				fmt.Printf("Failed to connect to discovered peer %s: %v\n", peerInfo.ID.String(), err)
			}
		}()
	}
}


// Network represents the P2P network layer
type Network struct {
	mu            sync.RWMutex
	host          host.Host
	dht           *dht.IpfsDHT
	pubsub        *pubsub.PubSub
	peers         map[peer.ID]*PeerInfo
	bootstrapPeers []multiaddr.Multiaddr
	config        *NetworkConfig
	ctx           context.Context
	cancel        context.CancelFunc
	chain         *chain.Chain
	mempool       *mempool.Mempool
}

// PeerInfo holds information about a connected peer
type PeerInfo struct {
	ID        peer.ID
	Addrs     []multiaddr.Multiaddr
	Protocols []string
	Connected time.Time
	LastSeen  time.Time
}

// NetworkConfig holds configuration for the network
type NetworkConfig struct {
	ListenPort     int
	BootstrapPeers []string
	EnableMDNS     bool
	EnableRelay    bool
	MaxPeers       int
	ConnectionTimeout time.Duration
}

// DefaultNetworkConfig returns the default network configuration
func DefaultNetworkConfig() *NetworkConfig {
	return &NetworkConfig{
		ListenPort:     0, // Random port
		BootstrapPeers: []string{},
		EnableMDNS:     true,
		EnableRelay:    false,
		MaxPeers:       50,
		ConnectionTimeout: 30 * time.Second,
	}
}

// Message represents a network message
type Message struct {
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
	From      peer.ID         `json:"from"`
}

// NewNetwork creates a new P2P network
func NewNetwork(config *NetworkConfig, chain *chain.Chain, mempool *mempool.Mempool) (*Network, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Generate a new key pair
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.Ed25519, 2048, rand.Reader)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}
	
	// Create libp2p host
	host, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", config.ListenPort)),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d/ws", config.ListenPort)),
		libp2p.Security(noise.ID, noise.New),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Transport(websocket.New),
		
		libp2p.EnableHolePunching(),
		libp2p.NATPortMap(),
	)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create host: %w", err)
	}
	
	// Create DHT
	dht, err := dht.New(ctx, host, dht.Mode(dht.ModeServer))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create DHT: %w", err)
	}
	
	// Create pubsub
	pubsub, err := pubsub.NewGossipSub(ctx, host, pubsub.WithMessageSigning(true))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create pubsub: %w", err)
	}
	
	// Parse bootstrap peers
	var bootstrapPeers []multiaddr.Multiaddr
	for _, addr := range config.BootstrapPeers {
		ma, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			continue
		}
		bootstrapPeers = append(bootstrapPeers, ma)
	}
	
	network := &Network{
		host:          host,
		dht:           dht,
		pubsub:        pubsub,
		peers:         make(map[peer.ID]*PeerInfo),
		bootstrapPeers: bootstrapPeers,
		config:        config,
		ctx:           ctx,
		cancel:        cancel,
		chain:         chain,
		mempool:       mempool,
	}
	
	// Set up event handlers
	host.Network().Notify(network)
	
	// Start peer discovery
	if err := network.startPeerDiscovery(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start peer discovery: %w", err)
	}
	
	// Connect to bootstrap peers
	go network.connectToBootstrapPeers()
	
	return network, nil
}

// startPeerDiscovery starts the peer discovery process
func (n *Network) startPeerDiscovery() error {
	// Start mDNS discovery if enabled
	if n.config.EnableMDNS {
		service := mdns.NewMdnsService(n.host, "gochain-discovery", n)
		if err := service.Start(); err != nil {
			return fmt.Errorf("failed to start mDNS service: %w", err)
		}
	}

	// Bootstrap the DHT
	if err := n.dht.Bootstrap(n.ctx); err != nil {
		return fmt.Errorf("failed to bootstrap DHT: %w", err)
	}

	// Setup a routing discovery service and attach it to the DHT
	discovery := routing.NewRoutingDiscovery(n.dht)
	discovery.Advertise(n.ctx, "gochain-discovery")

	return nil
}

// connectToBootstrapPeers connects to the bootstrap peers
func (n *Network) connectToBootstrapPeers() {
	var wg sync.WaitGroup
	for _, peerAddr := range n.bootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := n.host.Connect(n.ctx, *peerinfo); err != nil {
				fmt.Printf("Failed to connect to bootstrap peer %s: %v\n", peerinfo.ID.String(), err)
			} else {
				fmt.Printf("Connected to bootstrap peer: %s\n", peerinfo.ID.String())
			}
		}()
	}
	wg.Wait()
}

// Close closes the network host and DHT
func (n *Network) Close() error {
	n.cancel()
	if err := n.host.Close(); err != nil {
		return fmt.Errorf("failed to close host: %w", err)
	}
	if err := n.dht.Close(); err != nil {
		return fmt.Errorf("failed to close DHT: %w", err)
	}
	return nil
}

// GetHost returns the libp2p host
func (n *Network) GetHost() host.Host {
	return n.host
}

// GetPeers returns a list of connected peers
func (n *Network) GetPeers() []peer.ID {
	return n.host.Peerstore().Peers()
}

// GetContext returns the network's context
func (n *Network) GetContext() context.Context {
	return n.ctx
}

// SubscribeToBlocks subscribes to the blocks topic
func (n *Network) SubscribeToBlocks() (*pubsub.Subscription, error) {
	return n.pubsub.Subscribe("blocks")
}

// SubscribeToTransactions subscribes to the transactions topic
func (n *Network) SubscribeToTransactions() (*pubsub.Subscription, error) {
	return n.pubsub.Subscribe("transactions")
}

// PublishBlock publishes a block to the network
func (n *Network) PublishBlock(blockData []byte) error {
	return n.pubsub.Publish("blocks", blockData)
}

// PublishTransaction publishes a transaction to the network
func (n *Network) PublishTransaction(txData []byte) error {
	return n.pubsub.Publish("transactions", txData)
}