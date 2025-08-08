//go:build p2p
// +build p2p

package net

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gochain/gochain/pkg/block"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/libp2p/go-libp2p/p2p/transport/websocket"
	"github.com/multiformats/go-multiaddr"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-pubsub"
)

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
func NewNetwork(config *NetworkConfig) (*Network, error) {
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
		libp2p.EnableAutoRelay(),
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
	// Set up MDNS discovery if enabled
	if n.config.EnableMDNS {
		mdns.NewMdnsService(n.host, "gochain", n)
	}
	
	// Set up DHT discovery
	routingDiscovery := routing.NewRoutingDiscovery(n.dht)
	
	// Advertise ourselves
	_, err := routingDiscovery.Advertise(n.ctx, "gochain")
	if err != nil {
		return fmt.Errorf("failed to advertise: %w", err)
	}
	
	// Start looking for peers
	go n.discoverPeers(routingDiscovery)
	
	return nil
}

// discoverPeers continuously looks for new peers
func (n *Network) discoverPeers(routingDiscovery *routing.RoutingDiscovery) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			// Look for peers
			peerChan, err := routingDiscovery.FindPeers(n.ctx, "gochain")
			if err != nil {
				continue
			}
			
			for peer := range peerChan {
				if peer.ID == n.host.ID() {
					continue // Skip ourselves
				}
				
				// Try to connect
				go n.connectToPeer(peer)
			}
		}
	}
}

// connectToPeer attempts to connect to a peer
func (n *Network) connectToPeer(peerInfo peer.AddrInfo) {
	ctx, cancel := context.WithTimeout(n.ctx, n.config.ConnectionTimeout)
	defer cancel()
	
	if err := n.host.Connect(ctx, peerInfo); err != nil {
		return
	}
	
	// Add to our peer list
	n.mu.Lock()
	n.peers[peerInfo.ID] = &PeerInfo{
		ID:        peerInfo.ID,
		Addrs:     peerInfo.Addrs,
		Connected: time.Now(),
		LastSeen:  time.Now(),
	}
	n.mu.Unlock()
}

// connectToBootstrapPeers connects to bootstrap peers
func (n *Network) connectToBootstrapPeers() {
	for _, addr := range n.bootstrapPeers {
		peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			continue
		}
		
		go n.connectToPeer(*peerInfo)
	}
}

// BroadcastBlock broadcasts a block to all peers
func (n *Network) BroadcastBlock(block *block.Block) error {
	// Get the block topic
	topic, err := n.pubsub.Join("blocks")
	if err != nil {
		return fmt.Errorf("failed to join blocks topic: %w", err)
	}
	
	// Create message
	payload, err := json.Marshal(block)
	if err != nil {
		return fmt.Errorf("failed to marshal block: %w", err)
	}
	
	message := &Message{
		Type:      "block",
		Payload:   payload,
		Timestamp: time.Now(),
		From:      n.host.ID(),
	}
	
	// Marshal message
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	
	// Publish message
	return topic.Publish(n.ctx, data)
}

// BroadcastTransaction broadcasts a transaction to all peers
func (n *Network) BroadcastTransaction(tx *block.Transaction) error {
	// Get the transaction topic
	topic, err := n.pubsub.Join("transactions")
	if err != nil {
		return fmt.Errorf("failed to join transactions topic: %w", err)
	}
	
	// Create message
	payload, err := json.Marshal(tx)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %w", err)
	}
	
	message := &Message{
		Type:      "transaction",
		Payload:   payload,
		Timestamp: time.Now(),
		From:      n.host.ID(),
	}
	
	// Marshal message
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	
	// Publish message
	return topic.Publish(n.ctx, data)
}

// SubscribeToBlocks subscribes to block messages
func (n *Network) SubscribeToBlocks(handler func(*block.Block)) error {
	topic, err := n.pubsub.Join("blocks")
	if err != nil {
		return fmt.Errorf("failed to join blocks topic: %w", err)
	}
	
	sub, err := topic.Subscribe()
	if err != nil {
		return fmt.Errorf("failed to subscribe to blocks: %w", err)
	}
	
	go n.handleBlockMessages(sub, handler)
	
	return nil
}

// SubscribeToTransactions subscribes to transaction messages
func (n *Network) SubscribeToTransactions(handler func(*block.Transaction)) error {
	topic, err := n.pubsub.Join("transactions")
	if err != nil {
		return fmt.Errorf("failed to join transactions topic: %w", err)
	}
	
	sub, err := topic.Subscribe()
	if err != nil {
		return fmt.Errorf("failed to subscribe to transactions: %w", err)
	}
	
	go n.handleTransactionMessages(sub, handler)
	
	return nil
}

// handleBlockMessages handles incoming block messages
func (n *Network) handleBlockMessages(sub *pubsub.Subscription, handler func(*block.Block)) {
	defer sub.Cancel()
	
	for {
		msg, err := sub.Next(n.ctx)
		if err != nil {
			return
		}
		
		// Parse message
		var networkMsg Message
		if err := json.Unmarshal(msg.Data, &networkMsg); err != nil {
			continue
		}
		
		// Skip our own messages
		if networkMsg.From == n.host.ID() {
			continue
		}
		
		// Parse block
		var block block.Block
		if err := json.Unmarshal(networkMsg.Payload, &block); err != nil {
			continue
		}
		
		// Call handler
		handler(&block)
	}
}

// handleTransactionMessages handles incoming transaction messages
func (n *Network) handleTransactionMessages(sub *pubsub.Subscription, handler func(*block.Transaction)) {
	defer sub.Cancel()
	
	for {
		msg, err := sub.Next(n.ctx)
		if err != nil {
			return
		}
		
		// Parse message
		var networkMsg Message
		if err := json.Unmarshal(msg.Data, &networkMsg); err != nil {
			continue
		}
		
		// Skip our own messages
		if networkMsg.From == n.host.ID() {
			continue
		}
		
		// Parse transaction
		var tx block.Transaction
		if err := json.Unmarshal(networkMsg.Payload, &tx); err != nil {
			continue
		}
		
		// Call handler
		handler(&tx)
	}
}

// GetPeers returns information about connected peers
func (n *Network) GetPeers() []*PeerInfo {
	n.mu.RLock()
	defer n.mu.RUnlock()
	
	peers := make([]*PeerInfo, 0, len(n.peers))
	for _, peer := range n.peers {
		peers = append(peers, peer)
	}
	
	return peers
}

// GetPeerCount returns the number of connected peers
func (n *Network) GetPeerCount() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	
	return len(n.peers)
}

// GetHost returns the libp2p host
func (n *Network) GetHost() host.Host {
	return n.host
}

// GetMultiaddrs returns the multiaddrs of this node
func (n *Network) GetMultiaddrs() []multiaddr.Multiaddr {
	return n.host.Addrs()
}

// Close closes the network
func (n *Network) Close() error {
	n.cancel()
	
	if err := n.dht.Close(); err != nil {
		return fmt.Errorf("failed to close DHT: %w", err)
	}
	
	if err := n.host.Close(); err != nil {
		return fmt.Errorf("failed to close host: %w", err)
	}
	
	return nil
}

// Network interface implementation for libp2p
func (n *Network) Listen(network.Network, multiaddr.Multiaddr)      {}
func (n *Network) ListenClose(network.Network, multiaddr.Multiaddr) {}

func (n *Network) Connected(net network.Network, conn network.Conn) {
	peerInfo := &PeerInfo{
		ID:        conn.RemotePeer(),
		Addrs:     conn.RemoteMultiaddr().Encapsulate(multiaddr.StringCast("/p2p/" + conn.RemotePeer().String())).Split(),
		Connected: time.Now(),
		LastSeen:  time.Now(),
	}
	
	n.mu.Lock()
	n.peers[conn.RemotePeer()] = peerInfo
	n.mu.Unlock()
}

func (n *Network) Disconnected(net network.Network, conn network.Conn) {
	n.mu.Lock()
	delete(n.peers, conn.RemotePeer())
	n.mu.Unlock()
}

func (n *Network) OpenedStream(network.Network, network.Stream) {}
func (n *Network) ClosedStream(network.Network, network.Stream) {}

// MDNS interface implementation
func (n *Network) HandlePeerFound(pi peer.AddrInfo) {
	go n.connectToPeer(pi)
}

// String returns a string representation of the network
func (n *Network) String() string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	
	return fmt.Sprintf("Network{Peers: %d, HostID: %s}", 
		len(n.peers), n.host.ID().String())
} 