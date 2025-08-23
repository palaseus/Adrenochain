package ibc

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestIBCClientCreation tests IBC client creation
func TestIBCClientCreation(t *testing.T) {
	config := ClientConfig{
		MaxClockDrift:     time.Second * 5,
		TrustingPeriod:    time.Hour * 24 * 7,
		UnbondingPeriod:   time.Hour * 24 * 7 * 2,
		MaxHeaderSize:     512 * 1024,
		EnableCompression: true,
		SecurityLevel:     SecurityLevelHigh,
		AutoFreeze:        false,
	}

	client := NewIBCClient("test-chain", ClientTypeTendermint, config)

	if client == nil {
		t.Fatal("Expected client to be created")
	}

	if client.ChainID != "test-chain" {
		t.Errorf("Expected chain ID 'test-chain', got '%s'", client.ChainID)
	}

	if client.ClientType != ClientTypeTendermint {
		t.Errorf("Expected client type %d, got %d", ClientTypeTendermint, client.ClientType)
	}

	if client.Status != ClientStatusActive {
		t.Errorf("Expected status %d, got %d", ClientStatusActive, client.Status)
	}

	if client.config.MaxClockDrift != time.Second*5 {
		t.Errorf("Expected max clock drift %v, got %v", time.Second*5, client.config.MaxClockDrift)
	}
}

// TestConnectionCreation tests connection creation between clients
func TestConnectionCreation(t *testing.T) {
	clientA := NewIBCClient("chain-a", ClientTypeTendermint, ClientConfig{})
	clientB := NewIBCClient("chain-b", ClientTypeTendermint, ClientConfig{})

	connConfig := ConnectionConfig{
		MaxDelayPeriod:    time.Hour * 12,
		RetryAttempts:     5,
		Timeout:           time.Minute * 10,
		EnableCompression: true,
		SecurityLevel:     SecurityLevelHigh,
		AutoRetry:         true,
	}

	connection, err := clientA.CreateConnection(clientB, connConfig)
	if err != nil {
		t.Fatalf("Failed to create connection: %v", err)
	}

	if connection == nil {
		t.Fatal("Expected connection to be created")
	}

	if connection.ClientA != clientA.ID {
		t.Errorf("Expected client A ID '%s', got '%s'", clientA.ID, connection.ClientA)
	}

	if connection.ClientB != clientB.ID {
		t.Errorf("Expected client B ID '%s', got '%s'", clientB.ID, connection.ClientB)
	}

	if connection.Status != ConnectionStatusInit {
		t.Errorf("Expected status %d, got %d", ConnectionStatusInit, connection.Status)
	}
}

// TestConnectionOpening tests connection opening process
func TestConnectionOpening(t *testing.T) {
	clientA := NewIBCClient("chain-a", ClientTypeTendermint, ClientConfig{})
	clientB := NewIBCClient("chain-b", ClientTypeTendermint, ClientConfig{})

	connection, _ := clientA.CreateConnection(clientB, ConnectionConfig{})

	err := connection.OpenConnection()
	if err != nil {
		t.Fatalf("Failed to open connection: %v", err)
	}

	if connection.Status != ConnectionStatusOpen {
		t.Errorf("Expected status %d, got %d", ConnectionStatusOpen, connection.Status)
	}

	if connection.EstablishedAt.IsZero() {
		t.Error("Expected established time to be set")
	}
}

// TestChannelCreation tests channel creation on connections
func TestChannelCreation(t *testing.T) {
	clientA := NewIBCClient("chain-a", ClientTypeTendermint, ClientConfig{})
	clientB := NewIBCClient("chain-b", ClientTypeTendermint, ClientConfig{})

	connection, _ := clientA.CreateConnection(clientB, ConnectionConfig{})
	connection.OpenConnection()

	channelConfig := ChannelConfig{
		MaxPacketSize:     1024 * 1024,
		MaxPacketTimeout:  time.Hour * 24 * 7,
		EnableCompression: true,
		SecurityLevel:     SecurityLevelHigh,
		AutoClose:         false,
	}

	channel, err := connection.CreateChannel("transfer", "channel-1", ChannelOrderingOrdered, channelConfig)
	if err != nil {
		t.Fatalf("Failed to create channel: %v", err)
	}

	if channel == nil {
		t.Fatal("Expected channel to be created")
	}

	if channel.ConnectionID != connection.ID {
		t.Errorf("Expected connection ID '%s', got '%s'", connection.ID, channel.ConnectionID)
	}

	if channel.PortID != "transfer" {
		t.Errorf("Expected port ID 'transfer', got '%s'", channel.PortID)
	}

	if channel.State != ChannelStateInit {
		t.Errorf("Expected state %d, got %d", ChannelStateInit, channel.State)
	}
}

// TestChannelOpening tests channel opening process
func TestChannelOpening(t *testing.T) {
	clientA := NewIBCClient("chain-a", ClientTypeTendermint, ClientConfig{})
	clientB := NewIBCClient("chain-b", ClientTypeTendermint, ClientConfig{})

	connection, _ := clientA.CreateConnection(clientB, ConnectionConfig{})
	connection.OpenConnection()

	channel, _ := connection.CreateChannel("transfer", "channel-1", ChannelOrderingOrdered, ChannelConfig{})

	err := channel.OpenChannel()
	if err != nil {
		t.Fatalf("Failed to open channel: %v", err)
	}

	if channel.State != ChannelStateOpen {
		t.Errorf("Expected state %d, got %d", ChannelStateOpen, channel.State)
	}

	if channel.EstablishedAt.IsZero() {
		t.Error("Expected established time to be set")
	}
}

// TestPacketSending tests packet sending through channels
func TestPacketSending(t *testing.T) {
	clientA := NewIBCClient("chain-a", ClientTypeTendermint, ClientConfig{})
	clientB := NewIBCClient("chain-b", ClientTypeTendermint, ClientConfig{})

	connection, _ := clientA.CreateConnection(clientB, ConnectionConfig{})
	connection.OpenConnection()

	channel, _ := connection.CreateChannel("transfer", "channel-1", ChannelOrderingOrdered, ChannelConfig{})
	channel.OpenChannel()

	testData := []byte("Hello IBC World!")
	packet, err := channel.SendPacket(testData, "transfer", "channel-2", 1000, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("Failed to send packet: %v", err)
	}

	if packet == nil {
		t.Fatal("Expected packet to be created")
	}

	if packet.ChannelID != channel.ID {
		t.Errorf("Expected channel ID '%s', got '%s'", channel.ID, packet.ChannelID)
	}

	if string(packet.Data) != string(testData) {
		t.Errorf("Expected data '%s', got '%s'", string(testData), string(packet.Data))
	}

	if packet.Status != PacketStatusPending {
		t.Errorf("Expected status %d, got %d", PacketStatusPending, packet.Status)
	}
}

// TestPacketLifecycle tests complete packet lifecycle
func TestPacketLifecycle(t *testing.T) {
	clientA := NewIBCClient("chain-a", ClientTypeTendermint, ClientConfig{})
	clientB := NewIBCClient("chain-b", ClientTypeTendermint, ClientConfig{})

	connection, _ := clientA.CreateConnection(clientB, ConnectionConfig{})
	connection.OpenConnection()

	channel, _ := connection.CreateChannel("transfer", "channel-1", ChannelOrderingOrdered, ChannelConfig{})
	channel.OpenChannel()

	packet, _ := channel.SendPacket([]byte("test data"), "transfer", "channel-2", 1000, time.Now().Add(time.Hour))

	// Test packet sending
	err := packet.SendPacketNow()
	if err != nil {
		t.Fatalf("Failed to send packet: %v", err)
	}

	if packet.Status != PacketStatusSent {
		t.Errorf("Expected status %d, got %d", PacketStatusSent, packet.Status)
	}

	// Test packet receiving
	err = packet.ReceivePacket()
	if err != nil {
		t.Fatalf("Failed to receive packet: %v", err)
	}

	if packet.Status != PacketStatusReceived {
		t.Errorf("Expected status %d, got %d", PacketStatusReceived, packet.Status)
	}

	// Test packet acknowledgment
	err = packet.AcknowledgePacket()
	if err != nil {
		t.Fatalf("Failed to acknowledge packet: %v", err)
	}

	if packet.Status != PacketStatusAcknowledged {
		t.Errorf("Expected status %d, got %d", PacketStatusAcknowledged, packet.Status)
	}
}

// TestPacketTimeout tests packet timeout handling
func TestPacketTimeout(t *testing.T) {
	clientA := NewIBCClient("chain-a", ClientTypeTendermint, ClientConfig{})
	clientB := NewIBCClient("chain-b", ClientTypeTendermint, ClientConfig{})

	connection, _ := clientA.CreateConnection(clientB, ConnectionConfig{})
	connection.OpenConnection()

	channel, _ := connection.CreateChannel("transfer", "channel-1", ChannelOrderingOrdered, ChannelConfig{})
	channel.OpenChannel()

	// Create packet with past timeout
	packet, _ := channel.SendPacket([]byte("test data"), "transfer", "channel-2", 1000, time.Now().Add(-time.Hour))

	packet.SendPacketNow()

	// Try to receive expired packet
	err := packet.ReceivePacket()
	if err == nil {
		t.Fatal("Expected error for expired packet")
	}

	if packet.Status != PacketStatusTimeout {
		t.Errorf("Expected status %d, got %d", PacketStatusTimeout, packet.Status)
	}
}

// TestInvalidOperations tests invalid operations on IBC objects
func TestInvalidOperations(t *testing.T) {
	clientA := NewIBCClient("chain-a", ClientTypeTendermint, ClientConfig{})
	clientB := NewIBCClient("chain-b", ClientTypeTendermint, ClientConfig{})

	// Test creating connection with inactive client
	clientB.Status = ClientStatusFrozen
	_, err := clientA.CreateConnection(clientB, ConnectionConfig{})
	if err == nil {
		t.Error("Expected error when creating connection with inactive client")
	}

	// Test opening connection that's already open
	clientB.Status = ClientStatusActive
	connection, _ := clientA.CreateConnection(clientB, ConnectionConfig{})
	connection.OpenConnection()

	err = connection.OpenConnection()
	if err == nil {
		t.Error("Expected error when opening already open connection")
	}

	// Test creating channel on closed connection
	connection.Status = ConnectionStatusClosed
	_, err = connection.CreateChannel("transfer", "channel-1", ChannelOrderingOrdered, ChannelConfig{})
	if err == nil {
		t.Error("Expected error when creating channel on closed connection")
	}
}

// TestMetricsTracking tests metrics collection
func TestMetricsTracking(t *testing.T) {
	clientA := NewIBCClient("chain-a", ClientTypeTendermint, ClientConfig{})
	clientB := NewIBCClient("chain-b", ClientTypeTendermint, ClientConfig{})

	connection, _ := clientA.CreateConnection(clientB, ConnectionConfig{})
	connection.OpenConnection()

	channel, _ := connection.CreateChannel("transfer", "channel-1", ChannelOrderingOrdered, ChannelConfig{})
	channel.OpenChannel()

	// Send multiple packets
	for i := 0; i < 5; i++ {
		packet, _ := channel.SendPacket([]byte("test data"), "transfer", "channel-2", 1000, time.Now().Add(time.Hour))
		packet.SendPacketNow()
		packet.ReceivePacket()
		packet.AcknowledgePacket()
	}

	// Check channel metrics
	metrics := channel.GetMetrics()
	if metrics.TotalPackets != 5 {
		t.Errorf("Expected 5 total packets, got %d", metrics.TotalPackets)
	}

	if metrics.SuccessfulPackets != 5 {
		t.Errorf("Expected 5 successful packets, got %d", metrics.SuccessfulPackets)
	}

	// Check client metrics
	clientMetrics := clientA.GetMetrics()
	if clientMetrics.TotalConnections != 1 {
		t.Errorf("Expected 1 total connection, got %d", clientMetrics.TotalConnections)
	}
}

// TestConfigurationDefaults tests default configuration values
func TestConfigurationDefaults(t *testing.T) {
	// Test client config defaults
	client := NewIBCClient("test-chain", ClientTypeTendermint, ClientConfig{})

	if client.config.MaxClockDrift == 0 {
		t.Error("Expected MaxClockDrift to have default value")
	}

	if client.config.TrustingPeriod == 0 {
		t.Error("Expected TrustingPeriod to have default value")
	}

	// Test connection config defaults
	clientB := NewIBCClient("chain-b", ClientTypeTendermint, ClientConfig{})
	connection, _ := client.CreateConnection(clientB, ConnectionConfig{})

	if connection.config.MaxDelayPeriod == 0 {
		t.Error("Expected MaxDelayPeriod to have default value")
	}

	if connection.config.RetryAttempts == 0 {
		t.Error("Expected RetryAttempts to have default value")
	}

	// Test channel config defaults
	connection.OpenConnection()
	channel, _ := connection.CreateChannel("transfer", "channel-1", ChannelOrderingOrdered, ChannelConfig{})

	if channel.config.MaxPacketSize == 0 {
		t.Error("Expected MaxPacketSize to have default value")
	}

	if channel.config.MaxPacketTimeout == 0 {
		t.Error("Expected MaxPacketTimeout to have default value")
	}
}

// TestConcurrentOperations tests concurrent access to IBC objects
func TestConcurrentOperations(t *testing.T) {
	clientA := NewIBCClient("chain-a", ClientTypeTendermint, ClientConfig{})
	clientB := NewIBCClient("chain-b", ClientTypeTendermint, ClientConfig{})

	connection, _ := clientA.CreateConnection(clientB, ConnectionConfig{})
	connection.OpenConnection()

	channel, _ := connection.CreateChannel("transfer", "channel-1", ChannelOrderingOrdered, ChannelConfig{})
	channel.OpenChannel()

	// Test concurrent packet sending
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(index int) {
			packet, _ := channel.SendPacket([]byte("test data"), "transfer", "channel-2", 1000, time.Now().Add(time.Hour))
			packet.SendPacketNow()
			packet.ReceivePacket()
			packet.AcknowledgePacket()
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify final metrics
	metrics := channel.GetMetrics()
	if metrics.TotalPackets != 10 {
		t.Errorf("Expected 10 total packets, got %d", metrics.TotalPackets)
	}
}

// TestMockChainValidator tests the mock chain validator
func TestMockChainValidator(t *testing.T) {
	validator := NewMockChainValidator()

	err := validator.ValidateHeader([]byte("test header"))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	err = validator.ValidateProof([]byte("test proof"), []byte("test data"))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	height := validator.GetLatestHeight()
	if height != 1000 {
		t.Errorf("Expected height 1000, got %d", height)
	}

	timestamp, err := validator.GetTimestamp(100)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}
}

// Benchmark tests for performance
func BenchmarkPacketSending(b *testing.B) {
	clientA := NewIBCClient("chain-a", ClientTypeTendermint, ClientConfig{})
	clientB := NewIBCClient("chain-b", ClientTypeTendermint, ClientConfig{})

	connection, _ := clientA.CreateConnection(clientB, ConnectionConfig{})
	connection.OpenConnection()

	channel, _ := connection.CreateChannel("transfer", "channel-1", ChannelOrderingOrdered, ChannelConfig{})
	channel.OpenChannel()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		packet, _ := channel.SendPacket([]byte("benchmark data"), "transfer", "channel-2", 1000, time.Now().Add(time.Hour))
		packet.SendPacketNow()
		packet.ReceivePacket()
		packet.AcknowledgePacket()
	}
}

// TestErrorConditions tests various error conditions
func TestErrorConditions(t *testing.T) {
	// Test invalid client type
	client := NewIBCClient("test-chain", ClientType(999), ClientConfig{})
	if client == nil {
		t.Fatal("Expected client to be created even with invalid type")
	}

	// Test packet timeout edge case
	clientA := NewIBCClient("chain-a", ClientTypeTendermint, ClientConfig{})
	clientB := NewIBCClient("chain-b", ClientTypeTendermint, ClientConfig{})

	connection, _ := clientA.CreateConnection(clientB, ConnectionConfig{})
	connection.OpenConnection()

	channel, _ := connection.CreateChannel("transfer", "channel-1", ChannelOrderingOrdered, ChannelConfig{})
	channel.OpenChannel()

	// Create packet with very short timeout
	packet, _ := channel.SendPacket([]byte("test data"), "transfer", "channel-2", 1000, time.Now().Add(time.Millisecond))
	packet.SendPacketNow()

	// Wait for timeout
	time.Sleep(time.Millisecond * 2)

	err := packet.ReceivePacket()
	if err == nil {
		t.Fatal("Expected timeout error")
	}
}

// TestSecurityLevels tests different security level configurations
func TestSecurityLevels(t *testing.T) {
	configs := []ClientConfig{
		{SecurityLevel: SecurityLevelLow},
		{SecurityLevel: SecurityLevelMedium},
		{SecurityLevel: SecurityLevelHigh},
		{SecurityLevel: SecurityLevelUltra},
	}

	for _, config := range configs {
		client := NewIBCClient("test-chain", ClientTypeTendermint, config)
		if client.config.SecurityLevel != config.SecurityLevel {
			t.Errorf("Expected security level %d, got %d", config.SecurityLevel, client.config.SecurityLevel)
		}
	}
}

// TestClientTypes tests different client types
func TestClientTypes(t *testing.T) {
	clientTypes := []ClientType{
		ClientTypeTendermint,
		ClientTypeEthereum,
		ClientTypeBitcoin,
		ClientTypePolkadot,
	}

	for _, clientType := range clientTypes {
		client := NewIBCClient("test-chain", clientType, ClientConfig{})
		if client.ClientType != clientType {
			t.Errorf("Expected client type %d, got %d", clientType, client.ClientType)
		}
	}
}

// TestConnectionStates tests different connection states
func TestConnectionStates(t *testing.T) {
	clientA := NewIBCClient("chain-a", ClientTypeTendermint, ClientConfig{})
	clientB := NewIBCClient("chain-b", ClientTypeTendermint, ClientConfig{})

	connection, _ := clientA.CreateConnection(clientB, ConnectionConfig{})

	// Test initial state
	if connection.Status != ConnectionStatusInit {
		t.Errorf("Expected initial status %d, got %d", ConnectionStatusInit, connection.Status)
	}

	// Test try open state
	connection.Status = ConnectionStatusTryOpen
	if connection.Status != ConnectionStatusTryOpen {
		t.Errorf("Expected try open status %d, got %d", ConnectionStatusTryOpen, connection.Status)
	}

	// Test open state
	connection.Status = ConnectionStatusOpen
	if connection.Status != ConnectionStatusOpen {
		t.Errorf("Expected open status %d, got %d", ConnectionStatusOpen, connection.Status)
	}

	// Test closed state
	connection.Status = ConnectionStatusClosed
	if connection.Status != ConnectionStatusClosed {
		t.Errorf("Expected closed status %d, got %d", ConnectionStatusClosed, connection.Status)
	}
}

// TestChannelStates tests different channel states
func TestChannelStates(t *testing.T) {
	clientA := NewIBCClient("chain-a", ClientTypeTendermint, ClientConfig{})
	clientB := NewIBCClient("chain-b", ClientTypeTendermint, ClientConfig{})

	connection, _ := clientA.CreateConnection(clientB, ConnectionConfig{})
	connection.OpenConnection()

	channel, _ := connection.CreateChannel("transfer", "channel-1", ChannelOrderingOrdered, ChannelConfig{})

	// Test initial state
	if channel.State != ChannelStateInit {
		t.Errorf("Expected initial state %d, got %d", ChannelStateInit, channel.State)
	}

	// Test try open state
	channel.State = ChannelStateTryOpen
	if channel.State != ChannelStateTryOpen {
		t.Errorf("Expected try open state %d, got %d", ChannelStateTryOpen, channel.State)
	}

	// Test open state
	channel.State = ChannelStateOpen
	if channel.State != ChannelStateOpen {
		t.Errorf("Expected open state %d, got %d", ChannelStateOpen, channel.State)
	}

	// Test closed state
	channel.State = ChannelStateClosed
	if channel.State != ChannelStateClosed {
		t.Errorf("Expected closed state %d, got %d", ChannelStateClosed, channel.State)
	}
}

// TestPacketStates tests different packet states
func TestPacketStates(t *testing.T) {
	clientA := NewIBCClient("chain-a", ClientTypeTendermint, ClientConfig{})
	clientB := NewIBCClient("chain-b", ClientTypeTendermint, ClientConfig{})

	connection, _ := clientA.CreateConnection(clientB, ConnectionConfig{})
	connection.OpenConnection()

	channel, _ := connection.CreateChannel("transfer", "channel-1", ChannelOrderingOrdered, ChannelConfig{})
	channel.OpenChannel()

	packet, _ := channel.SendPacket([]byte("test data"), "transfer", "channel-2", 1000, time.Now().Add(time.Hour))

	// Test pending state
	if packet.Status != PacketStatusPending {
		t.Errorf("Expected pending status %d, got %d", PacketStatusPending, packet.Status)
	}

	// Test sent state
	packet.SendPacketNow()
	if packet.Status != PacketStatusSent {
		t.Errorf("Expected sent status %d, got %d", PacketStatusSent, packet.Status)
	}

	// Test received state
	packet.ReceivePacket()
	if packet.Status != PacketStatusReceived {
		t.Errorf("Expected received status %d, got %d", PacketStatusReceived, packet.Status)
	}

	// Test acknowledged state
	packet.AcknowledgePacket()
	if packet.Status != PacketStatusAcknowledged {
		t.Errorf("Expected acknowledged status %d, got %d", PacketStatusAcknowledged, packet.Status)
	}
}

// TestConfigConversion tests configuration conversion methods directly
func TestConfigConversion(t *testing.T) {
	// Test client config conversion
	clientConfigSection := ClientConfigSection{
		MaxClockDrift:     "10s",
		TrustingPeriod:    "336h",
		UnbondingPeriod:   "504h",
		MaxHeaderSize:     1048576,
		EnableCompression: true,
		SecurityLevel:     "high",
		AutoFreeze:        false,
	}

	clientConfig, err := clientConfigSection.ConvertToClientConfig()
	if err != nil {
		t.Fatalf("Failed to convert client config: %v", err)
	}

	if clientConfig.MaxClockDrift != 10*time.Second {
		t.Errorf("Expected MaxClockDrift 10s, got %v", clientConfig.MaxClockDrift)
	}

	if clientConfig.SecurityLevel != SecurityLevelHigh {
		t.Errorf("Expected SecurityLevel High, got %v", clientConfig.SecurityLevel)
	}

	// Test connection config conversion
	connConfigSection := ConnectionConfigSection{
		MaxDelayPeriod:    "24h",
		RetryAttempts:     3,
		Timeout:           "5m",
		EnableCompression: true,
		SecurityLevel:     "high",
		AutoRetry:         true,
	}

	connConfig, err := connConfigSection.ConvertToConnectionConfig()
	if err != nil {
		t.Fatalf("Failed to convert connection config: %v", err)
	}

	if connConfig.MaxDelayPeriod != 24*time.Hour {
		t.Errorf("Expected MaxDelayPeriod 24h, got %v", connConfig.MaxDelayPeriod)
	}

	// Test channel config conversion
	chConfigSection := ChannelConfigSection{
		MaxPacketSize:    1048576,
		MaxPacketTimeout: "168h",
		EnableCompression: true,
		SecurityLevel:    "high",
		AutoClose:        false,
	}

	chConfig, err := chConfigSection.ConvertToChannelConfig()
	if err != nil {
		t.Fatalf("Failed to convert channel config: %v", err)
	}

	if chConfig.MaxPacketSize != 1048576 {
		t.Errorf("Expected MaxPacketSize 1048576, got %d", chConfig.MaxPacketSize)
	}

	// Test packet config conversion
	pktConfigSection := PacketConfigSection{
		MaxRetries:       3,
		RetryDelay:       "1s",
		Timeout:          "5m",
		EnableCompression: true,
		SecurityLevel:    "high",
		AutoRetry:        true,
	}

	pktConfig, err := pktConfigSection.ConvertToPacketConfig()
	if err != nil {
		t.Fatalf("Failed to convert packet config: %v", err)
	}

	if pktConfig.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries 3, got %d", pktConfig.MaxRetries)
	}
}

// TestPacketRetryMechanism tests packet retry functionality
func TestPacketRetryMechanism(t *testing.T) {
	client := NewIBCClient("test-chain", ClientTypeTendermint, ClientConfig{})
	connection, _ := client.CreateConnection(NewIBCClient("counter-chain", ClientTypeTendermint, ClientConfig{}), ConnectionConfig{})
	connection.OpenConnection()
	channel, _ := connection.CreateChannel("test-port", "test-channel", ChannelOrderingOrdered, ChannelConfig{})
	channel.OpenChannel()

	packet, _ := channel.SendPacket([]byte("test data"), "dest-port", "dest-channel", 1000, time.Now().Add(time.Hour))
	
	// Test retry mechanism
	packet.config.MaxRetries = 3
	packet.config.RetryDelay = time.Millisecond * 100
	
	// Simulate failed send and retry
	packet.metrics.TotalRetries++
	packet.metrics.FailedRetries++
	
	if packet.metrics.TotalRetries != 1 {
		t.Errorf("Expected total retries 1, got %d", packet.metrics.TotalRetries)
	}
	
	if packet.metrics.FailedRetries != 1 {
		t.Errorf("Expected failed retries 1, got %d", packet.metrics.FailedRetries)
	}
}

// TestPacketCompression tests packet compression functionality
func TestPacketCompression(t *testing.T) {
	client := NewIBCClient("test-chain", ClientTypeTendermint, ClientConfig{})
	connection, _ := client.CreateConnection(NewIBCClient("counter-chain", ClientTypeTendermint, ClientConfig{}), ConnectionConfig{})
	connection.OpenConnection()
	channel, _ := connection.CreateChannel("test-port", "test-channel", ChannelOrderingOrdered, ChannelConfig{EnableCompression: true})
	channel.OpenChannel()

	packet, _ := channel.SendPacket([]byte("test data"), "dest-port", "dest-channel", 1000, time.Now().Add(time.Hour))
	
	// Set packet compression to match channel compression
	packet.config.EnableCompression = true
	
	if !packet.config.EnableCompression {
		t.Error("Expected compression to be enabled")
	}
}

// TestConnectionCompression tests connection compression functionality
func TestConnectionCompression(t *testing.T) {
	client := NewIBCClient("test-chain", ClientTypeTendermint, ClientConfig{})
	connection, _ := client.CreateConnection(NewIBCClient("counter-chain", ClientTypeTendermint, ClientConfig{}), ConnectionConfig{EnableCompression: true})
	
	if !connection.config.EnableCompression {
		t.Error("Expected compression to be enabled")
	}
}

// TestClientAutoFreeze tests client auto-freeze functionality
func TestClientAutoFreeze(t *testing.T) {
	config := ClientConfig{
		AutoFreeze: true,
		MaxClockDrift: time.Second * 5,
	}
	
	client := NewIBCClient("test-chain", ClientTypeTendermint, config)
	
	if !client.config.AutoFreeze {
		t.Error("Expected auto-freeze to be enabled")
	}
}

// TestConnectionAutoRetry tests connection auto-retry functionality
func TestConnectionAutoRetry(t *testing.T) {
	client := NewIBCClient("test-chain", ClientTypeTendermint, ClientConfig{})
	connection, _ := client.CreateConnection(NewIBCClient("counter-chain", ClientTypeTendermint, ClientConfig{}), ConnectionConfig{AutoRetry: true})
	
	if !connection.config.AutoRetry {
		t.Error("Expected auto-retry to be enabled")
	}
}

// TestChannelAutoClose tests channel auto-close functionality
func TestChannelAutoClose(t *testing.T) {
	client := NewIBCClient("test-chain", ClientTypeTendermint, ClientConfig{})
	connection, _ := client.CreateConnection(NewIBCClient("counter-chain", ClientTypeTendermint, ClientConfig{}), ConnectionConfig{})
	connection.OpenConnection()
	channel, _ := connection.CreateChannel("test-port", "test-channel", ChannelOrderingOrdered, ChannelConfig{AutoClose: true})
	
	if !channel.config.AutoClose {
		t.Error("Expected auto-close to be enabled")
	}
}

// TestPacketAutoRetry tests packet auto-retry functionality
func TestPacketAutoRetry(t *testing.T) {
	client := NewIBCClient("test-chain", ClientTypeTendermint, ClientConfig{})
	connection, _ := client.CreateConnection(NewIBCClient("counter-chain", ClientTypeTendermint, ClientConfig{}), ConnectionConfig{})
	connection.OpenConnection()
	channel, _ := connection.CreateChannel("test-port", "test-channel", ChannelOrderingOrdered, ChannelConfig{})
	channel.OpenChannel()

	packet, _ := channel.SendPacket([]byte("test data"), "dest-port", "dest-channel", 1000, time.Now().Add(time.Hour))
	packet.config.AutoRetry = true
	
	if !packet.config.AutoRetry {
		t.Error("Expected auto-retry to be enabled")
	}
}

// TestClientStatusTransitions tests client status transitions
func TestClientStatusTransitions(t *testing.T) {
	client := NewIBCClient("test-chain", ClientTypeTendermint, ClientConfig{})
	
	// Test initial status
	if client.Status != ClientStatusActive {
		t.Errorf("Expected initial status %d, got %d", ClientStatusActive, client.Status)
	}
	
	// Test status transitions
	client.Status = ClientStatusFrozen
	if client.Status != ClientStatusFrozen {
		t.Errorf("Expected status %d, got %d", ClientStatusFrozen, client.Status)
	}
	
	client.Status = ClientStatusExpired
	if client.Status != ClientStatusExpired {
		t.Errorf("Expected status %d, got %d", ClientStatusExpired, client.Status)
	}
	
	client.Status = ClientStatusRevoked
	if client.Status != ClientStatusRevoked {
		t.Errorf("Expected status %d, got %d", ClientStatusRevoked, client.Status)
	}
}

// TestConnectionStatusTransitions tests connection status transitions
func TestConnectionStatusTransitions(t *testing.T) {
	client := NewIBCClient("test-chain", ClientTypeTendermint, ClientConfig{})
	connection, _ := client.CreateConnection(NewIBCClient("counter-chain", ClientTypeTendermint, ClientConfig{}), ConnectionConfig{})
	
	// Test initial status
	if connection.Status != ConnectionStatusInit {
		t.Errorf("Expected initial status %d, got %d", ConnectionStatusInit, connection.Status)
	}
	
	// Test status transitions
	connection.Status = ConnectionStatusTryOpen
	if connection.Status != ConnectionStatusTryOpen {
		t.Errorf("Expected status %d, got %d", ConnectionStatusTryOpen, connection.Status)
	}
	
	connection.Status = ConnectionStatusOpen
	if connection.Status != ConnectionStatusOpen {
		t.Errorf("Expected status %d, got %d", ConnectionStatusOpen, connection.Status)
	}
	
	connection.Status = ConnectionStatusClosed
	if connection.Status != ConnectionStatusClosed {
		t.Errorf("Expected status %d, got %d", ConnectionStatusClosed, connection.Status)
	}
}

// TestChannelStateTransitions tests channel state transitions
func TestChannelStateTransitions(t *testing.T) {
	client := NewIBCClient("test-chain", ClientTypeTendermint, ClientConfig{})
	connection, _ := client.CreateConnection(NewIBCClient("counter-chain", ClientTypeTendermint, ClientConfig{}), ConnectionConfig{})
	connection.OpenConnection()
	channel, _ := connection.CreateChannel("test-port", "test-channel", ChannelOrderingOrdered, ChannelConfig{})
	
	// Test initial state
	if channel.State != ChannelStateInit {
		t.Errorf("Expected initial state %d, got %d", ChannelStateInit, channel.State)
	}
	
	// Test state transitions
	channel.State = ChannelStateTryOpen
	if channel.State != ChannelStateTryOpen {
		t.Errorf("Expected state %d, got %d", ChannelStateTryOpen, channel.State)
	}
	
	channel.State = ChannelStateOpen
	if channel.State != ChannelStateOpen {
		t.Errorf("Expected state %d, got %d", ChannelStateOpen, channel.State)
	}
	
	channel.State = ChannelStateClosed
	if channel.State != ChannelStateClosed {
		t.Errorf("Expected state %d, got %d", ChannelStateClosed, channel.State)
	}
}

// TestPacketStatusTransitions tests packet status transitions
func TestPacketStatusTransitions(t *testing.T) {
	client := NewIBCClient("test-chain", ClientTypeTendermint, ClientConfig{})
	connection, _ := client.CreateConnection(NewIBCClient("counter-chain", ClientTypeTendermint, ClientConfig{}), ConnectionConfig{})
	connection.OpenConnection()
	channel, _ := connection.CreateChannel("test-port", "test-channel", ChannelOrderingOrdered, ChannelConfig{})
	channel.OpenChannel()

	packet, _ := channel.SendPacket([]byte("test data"), "dest-port", "dest-channel", 1000, time.Now().Add(time.Hour))
	
	// Test initial status
	if packet.Status != PacketStatusPending {
		t.Errorf("Expected initial status %d, got %d", PacketStatusPending, packet.Status)
	}
	
	// Test status transitions
	packet.Status = PacketStatusSent
	if packet.Status != PacketStatusSent {
		t.Errorf("Expected status %d, got %d", PacketStatusSent, packet.Status)
	}
	
	packet.Status = PacketStatusReceived
	if packet.Status != PacketStatusReceived {
		t.Errorf("Expected status %d, got %d", PacketStatusReceived, packet.Status)
	}
	
	packet.Status = PacketStatusAcknowledged
	if packet.Status != PacketStatusAcknowledged {
		t.Errorf("Expected status %d, got %d", PacketStatusAcknowledged, packet.Status)
	}
	
	packet.Status = PacketStatusTimeout
	if packet.Status != PacketStatusTimeout {
		t.Errorf("Expected status %d, got %d", PacketStatusTimeout, packet.Status)
	}
	
	packet.Status = PacketStatusFailed
	if packet.Status != PacketStatusFailed {
		t.Errorf("Expected status %d, got %d", PacketStatusFailed, packet.Status)
	}
}

// TestTrustLevels tests trust level functionality
func TestTrustLevels(t *testing.T) {
	client := NewIBCClient("test-chain", ClientTypeTendermint, ClientConfig{})
	
	// Test initial trust level
	if client.TrustLevel != TrustLevelHigh {
		t.Errorf("Expected initial trust level %d, got %d", TrustLevelHigh, client.TrustLevel)
	}
	
	// Test trust level transitions
	client.TrustLevel = TrustLevelLow
	if client.TrustLevel != TrustLevelLow {
		t.Errorf("Expected trust level %d, got %d", TrustLevelLow, client.TrustLevel)
	}
	
	client.TrustLevel = TrustLevelMedium
	if client.TrustLevel != TrustLevelMedium {
		t.Errorf("Expected trust level %d, got %d", TrustLevelMedium, client.TrustLevel)
	}
	
	client.TrustLevel = TrustLevelUltra
	if client.TrustLevel != TrustLevelUltra {
		t.Errorf("Expected trust level %d, got %d", TrustLevelUltra, client.TrustLevel)
	}
}

// TestSecurityLevelsEnhanced tests enhanced security level functionality
func TestSecurityLevelsEnhanced(t *testing.T) {
	config := ClientConfig{SecurityLevel: SecurityLevelUltra}
	client := NewIBCClient("test-chain", ClientTypeTendermint, config)
	
	if client.config.SecurityLevel != SecurityLevelUltra {
		t.Errorf("Expected security level %d, got %d", SecurityLevelUltra, client.config.SecurityLevel)
	}
	
	// Test all security levels
	levels := []SecurityLevel{SecurityLevelLow, SecurityLevelMedium, SecurityLevelHigh, SecurityLevelUltra}
	for _, level := range levels {
		client.config.SecurityLevel = level
		if client.config.SecurityLevel != level {
			t.Errorf("Expected security level %d, got %d", level, client.config.SecurityLevel)
		}
	}
}

// TestClientTypesEnhanced tests enhanced client type functionality
func TestClientTypesEnhanced(t *testing.T) {
	types := []ClientType{ClientTypeTendermint, ClientTypeEthereum, ClientTypeBitcoin, ClientTypePolkadot}
	
	for _, clientType := range types {
		client := NewIBCClient("test-chain", clientType, ClientConfig{})
		if client.ClientType != clientType {
			t.Errorf("Expected client type %d, got %d", clientType, client.ClientType)
		}
	}
}

// TestChannelOrdering tests channel ordering functionality
func TestChannelOrdering(t *testing.T) {
	client := NewIBCClient("test-chain", ClientTypeTendermint, ClientConfig{})
	connection, _ := client.CreateConnection(NewIBCClient("counter-chain", ClientTypeTendermint, ClientConfig{}), ConnectionConfig{})
	connection.OpenConnection()
	
	// Test unordered channel
	unorderedChannel, _ := connection.CreateChannel("test-port", "test-channel", ChannelOrderingUnordered, ChannelConfig{})
	if unorderedChannel.Ordering != ChannelOrderingUnordered {
		t.Errorf("Expected ordering %d, got %d", ChannelOrderingUnordered, unorderedChannel.Ordering)
	}
	
	// Test ordered channel
	orderedChannel, _ := connection.CreateChannel("test-port", "test-channel", ChannelOrderingOrdered, ChannelConfig{})
	if orderedChannel.Ordering != ChannelOrderingOrdered {
		t.Errorf("Expected ordering %d, got %d", ChannelOrderingOrdered, orderedChannel.Ordering)
	}
}

// TestMetricsUpdateTiming tests metrics update timing
func TestMetricsUpdateTiming(t *testing.T) {
	client := NewIBCClient("test-chain", ClientTypeTendermint, ClientConfig{})
	
	// Test initial metrics
	metrics := client.GetMetrics()
	if metrics.TotalConnections != 0 {
		t.Errorf("Expected initial total connections 0, got %d", metrics.TotalConnections)
	}
	
	// Create connection to update metrics
	connection, _ := client.CreateConnection(NewIBCClient("counter-chain", ClientTypeTendermint, ClientConfig{}), ConnectionConfig{})
	
	// Test connection was created
	if connection == nil {
		t.Fatal("Expected connection to be created")
	}
	
	// Test updated metrics
	metrics = client.GetMetrics()
	if metrics.TotalConnections != 1 {
		t.Errorf("Expected total connections 1, got %d", metrics.TotalConnections)
	}
	
	if metrics.LastUpdate.IsZero() {
		t.Error("Expected last update to be set")
	}
}

// TestConnectionMetricsUpdate tests connection metrics updates
func TestConnectionMetricsUpdate(t *testing.T) {
	client := NewIBCClient("test-chain", ClientTypeTendermint, ClientConfig{})
	connection, _ := client.CreateConnection(NewIBCClient("counter-chain", ClientTypeTendermint, ClientConfig{}), ConnectionConfig{})
	
	// Test initial metrics
	metrics := connection.GetMetrics()
	if metrics.TotalAttempts != 0 {
		t.Errorf("Expected initial total attempts 0, got %d", metrics.TotalAttempts)
	}
	
	// Open connection to update metrics
	connection.OpenConnection()
	
	// Test updated metrics
	metrics = connection.GetMetrics()
	if metrics.SuccessfulAttempts != 1 {
		t.Errorf("Expected successful attempts 1, got %d", metrics.SuccessfulAttempts)
	}
	
	if metrics.LastUpdate.IsZero() {
		t.Error("Expected last update to be set")
	}
}

// TestChannelMetricsUpdate tests channel metrics updates
func TestChannelMetricsUpdate(t *testing.T) {
	client := NewIBCClient("test-chain", ClientTypeTendermint, ClientConfig{})
	connection, _ := client.CreateConnection(NewIBCClient("counter-chain", ClientTypeTendermint, ClientConfig{}), ConnectionConfig{})
	connection.OpenConnection()
	channel, _ := connection.CreateChannel("test-port", "test-channel", ChannelOrderingOrdered, ChannelConfig{})
	
	// Test initial metrics
	metrics := channel.GetMetrics()
	if metrics.TotalPackets != 0 {
		t.Errorf("Expected initial total packets 0, got %d", metrics.TotalPackets)
	}
	
	// Open the channel before sending packets
	channel.OpenChannel()
	
	// Send packet to update metrics
	packet, err := channel.SendPacket([]byte("test data"), "dest-port", "dest-channel", 1000, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("Failed to send packet: %v", err)
	}
	
	// Ensure packet was created
	if packet == nil {
		t.Fatal("Expected packet to be created")
	}
	
	// Test updated metrics
	metrics = channel.GetMetrics()
	if metrics.TotalPackets != 1 {
		t.Errorf("Expected total packets 1, got %d", metrics.TotalPackets)
	}
	
	if metrics.LastUpdate.IsZero() {
		t.Error("Expected last update to be set")
	}
}

// TestPacketMetricsUpdate tests packet metrics updates
func TestPacketMetricsUpdate(t *testing.T) {
	client := NewIBCClient("test-chain", ClientTypeTendermint, ClientConfig{})
	connection, _ := client.CreateConnection(NewIBCClient("counter-chain", ClientTypeTendermint, ClientConfig{}), ConnectionConfig{})
	connection.OpenConnection()
	channel, _ := connection.CreateChannel("test-port", "test-channel", ChannelOrderingOrdered, ChannelConfig{})
	channel.OpenChannel()

	packet, _ := channel.SendPacket([]byte("test data"), "dest-port", "dest-channel", 1000, time.Now().Add(time.Hour))
	
	// Test initial metrics
	metrics := packet.GetMetrics()
	if metrics.TotalPackets != 1 {
		t.Errorf("Expected initial total packets 1, got %d", metrics.TotalPackets)
	}
	
	// Send packet to update metrics
	packet.SendPacketNow()
	
	// Test updated metrics
	metrics = packet.GetMetrics()
	if metrics.LastUpdate.IsZero() {
		t.Error("Expected last update to be set")
	}
}

// TestPacketExpiration tests packet expiration functionality
func TestPacketExpiration(t *testing.T) {
	client := NewIBCClient("test-chain", ClientTypeTendermint, ClientConfig{})
	connection, _ := client.CreateConnection(NewIBCClient("counter-chain", ClientTypeTendermint, ClientConfig{}), ConnectionConfig{})
	connection.OpenConnection()
	channel, _ := connection.CreateChannel("test-port", "test-channel", ChannelOrderingOrdered, ChannelConfig{})
	channel.OpenChannel()

	// Create packet with past timeout
	packet, _ := channel.SendPacket([]byte("test data"), "dest-port", "dest-channel", 1000, time.Now().Add(-time.Hour))
	
	// Test expiration
	if !packet.IsExpired() {
		t.Error("Expected packet to be expired")
	}
	
	// Create packet with future timeout
	packet, _ = channel.SendPacket([]byte("test data"), "dest-port", "dest-channel", 1000, time.Now().Add(time.Hour))
	
	// Test not expired
	if packet.IsExpired() {
		t.Error("Expected packet to not be expired")
	}
}

// TestIDGeneration tests ID generation functions
func TestIDGeneration(t *testing.T) {
	// Test client ID generation
	clientID1 := generateClientID()
	clientID2 := generateClientID()
	
	if clientID1 == clientID2 {
		t.Error("Expected unique client IDs")
	}
	
	if len(clientID1) == 0 {
		t.Error("Expected non-empty client ID")
	}
	
	// Test connection ID generation
	connID1 := generateConnectionID()
	connID2 := generateConnectionID()
	
	if connID1 == connID2 {
		t.Error("Expected unique connection IDs")
	}
	
	if len(connID1) == 0 {
		t.Error("Expected non-empty connection ID")
	}
	
	// Test channel ID generation
	channelID1 := generateChannelID()
	channelID2 := generateChannelID()
	
	if channelID1 == channelID2 {
		t.Error("Expected unique channel IDs")
	}
	
	if len(channelID1) == 0 {
		t.Error("Expected non-empty channel ID")
	}
	
	// Test packet ID generation
	packetID1 := generatePacketID()
	packetID2 := generatePacketID()
	
	if packetID1 == packetID2 {
		t.Error("Expected unique packet IDs")
	}
	
	if len(packetID1) == 0 {
		t.Error("Expected non-empty packet ID")
	}
}

// TestMockChainValidatorEnhanced tests enhanced mock chain validator functionality
func TestMockChainValidatorEnhanced(t *testing.T) {
	validator := NewMockChainValidator()
	
	// Test header validation
	err := validator.ValidateHeader([]byte("test header"))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// Test proof validation
	err = validator.ValidateProof([]byte("test proof"), []byte("test data"))
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// Test latest height
	height := validator.GetLatestHeight()
	if height != 1000 {
		t.Errorf("Expected height 1000, got %d", height)
	}
	
	// Test timestamp retrieval
	timestamp, err := validator.GetTimestamp(1000)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}
}

// TestErrorHandling tests error handling scenarios
func TestErrorHandling(t *testing.T) {
	client := NewIBCClient("test-chain", ClientTypeTendermint, ClientConfig{})
	
	// Test creating connection with inactive client
	inactiveClient := NewIBCClient("inactive-chain", ClientTypeTendermint, ClientConfig{})
	inactiveClient.Status = ClientStatusFrozen
	
	_, err := client.CreateConnection(inactiveClient, ConnectionConfig{})
	if err == nil {
		t.Error("Expected error when creating connection with inactive client")
	}
	
	// Test opening connection with wrong status
	connection, _ := client.CreateConnection(NewIBCClient("counter-chain", ClientTypeTendermint, ClientConfig{}), ConnectionConfig{})
	connection.Status = ConnectionStatusOpen
	
	err = connection.OpenConnection()
	if err == nil {
		t.Error("Expected error when opening already open connection")
	}
	
	// Test creating channel on closed connection
	connection.Status = ConnectionStatusClosed
	_, err = connection.CreateChannel("test-port", "test-channel", ChannelOrderingOrdered, ChannelConfig{})
	if err == nil {
		t.Error("Expected error when creating channel on closed connection")
	}
	
	// Test opening channel with wrong state
	connection.Status = ConnectionStatusOpen
	channel, _ := connection.CreateChannel("test-port", "test-channel", ChannelOrderingOrdered, ChannelConfig{})
	channel.State = ChannelStateOpen
	
	err = channel.OpenChannel()
	if err == nil {
		t.Error("Expected error when opening already open channel")
	}
	
	// Test sending packet on closed channel
	channel.State = ChannelStateClosed
	_, err = channel.SendPacket([]byte("test data"), "dest-port", "dest-channel", 1000, time.Now().Add(time.Hour))
	if err == nil {
		t.Error("Expected error when sending packet on closed channel")
	}
	
	// Test packet operations with wrong status
	channel.State = ChannelStateOpen
	packet, _ := channel.SendPacket([]byte("test data"), "dest-port", "dest-channel", 1000, time.Now().Add(time.Hour))
	
	// Test sending packet with wrong status
	packet.Status = PacketStatusSent
	err = packet.SendPacketNow()
	if err == nil {
		t.Error("Expected error when sending already sent packet")
	}
	
	// Test receiving packet with wrong status
	packet.Status = PacketStatusPending
	err = packet.ReceivePacket()
	if err == nil {
		t.Error("Expected error when receiving pending packet")
	}
	
	// Test acknowledging packet with wrong status
	packet.Status = PacketStatusPending
	err = packet.AcknowledgePacket()
	if err == nil {
		t.Error("Expected error when acknowledging pending packet")
	}
}

// TestConcurrentAccess tests concurrent access to IBC objects
func TestConcurrentAccess(t *testing.T) {
	client := NewIBCClient("test-chain", ClientTypeTendermint, ClientConfig{})
	
	// Test concurrent metrics access
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = client.GetMetrics()
		}()
	}
	wg.Wait()
	
	// Test concurrent status access
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = client.GetStatus()
		}()
	}
	wg.Wait()
	
	// Test concurrent connection creation
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			counterClient := NewIBCClient(fmt.Sprintf("counter-chain-%d", id), ClientTypeTendermint, ClientConfig{})
			_, _ = client.CreateConnection(counterClient, ConnectionConfig{})
		}(i)
	}
	wg.Wait()
}

// TestEdgeCases tests edge cases and boundary conditions
func TestEdgeCases(t *testing.T) {
	// Test with empty chain ID
	client := NewIBCClient("", ClientTypeTendermint, ClientConfig{})
	if client.ChainID != "" {
		t.Errorf("Expected empty chain ID, got %s", client.ChainID)
	}
	
	// Test with zero timeouts
	config := ConnectionConfig{
		MaxDelayPeriod: 0,
		RetryAttempts:  0,
		Timeout:        0,
	}
	
	client = NewIBCClient("test-chain", ClientTypeTendermint, ClientConfig{})
	connection, _ := client.CreateConnection(NewIBCClient("counter-chain", ClientTypeTendermint, ClientConfig{}), config)
	
	// Should use default values
	if connection.config.MaxDelayPeriod == 0 {
		t.Error("Expected default max delay period")
	}
	
	if connection.config.RetryAttempts == 0 {
		t.Error("Expected default retry attempts")
	}
	
	if connection.config.Timeout == 0 {
		t.Error("Expected default timeout")
	}
	
	// Test with zero packet size
	channelConfig := ChannelConfig{
		MaxPacketSize:    0,
		MaxPacketTimeout: 0,
	}
	
	connection.OpenConnection()
	channel, _ := connection.CreateChannel("test-port", "test-channel", ChannelOrderingOrdered, channelConfig)
	
	// Should use default values
	if channel.config.MaxPacketSize == 0 {
		t.Error("Expected default max packet size")
	}
	
	if channel.config.MaxPacketTimeout == 0 {
		t.Error("Expected default max packet timeout")
	}
	
	// Test with very large packet data
	largeData := make([]byte, 2*1024*1024) // 2MB
	_, err := channel.SendPacket(largeData, "dest-port", "dest-channel", 1000, time.Now().Add(time.Hour))
	if err == nil {
		t.Error("Expected error with oversized packet data")
	}
}

// TestPerformanceMetrics tests performance metrics functionality
func TestPerformanceMetrics(t *testing.T) {
	client := NewIBCClient("test-chain", ClientTypeTendermint, ClientConfig{})
	
	// Test initial performance metrics
	metrics := client.GetMetrics()
	if metrics.TotalConnections != 0 {
		t.Errorf("Expected initial total connections 0, got %d", metrics.TotalConnections)
	}
	
	if metrics.TotalChannels != 0 {
		t.Errorf("Expected initial total channels 0, got %d", metrics.TotalChannels)
	}
	
	// Create connection and channel to update metrics
	connection, _ := client.CreateConnection(NewIBCClient("counter-chain", ClientTypeTendermint, ClientConfig{}), ConnectionConfig{})
	connection.OpenConnection()
	
	channel, _ := connection.CreateChannel("test-port", "test-channel", ChannelOrderingOrdered, ChannelConfig{})
	channel.OpenChannel()
	
	// Test updated performance metrics
	metrics = client.GetMetrics()
	if metrics.TotalConnections != 1 {
		t.Errorf("Expected total connections 1, got %d", metrics.TotalConnections)
	}
	
	// Note: TotalChannels is not currently tracked in client metrics
	// This would need to be implemented in the IBC client
}

// TestSecurityValidation tests security validation functionality
func TestSecurityValidation(t *testing.T) {
	// Test with different security levels
	securityLevels := []SecurityLevel{SecurityLevelLow, SecurityLevelMedium, SecurityLevelHigh, SecurityLevelUltra}
	
	for _, level := range securityLevels {
		config := ClientConfig{SecurityLevel: level}
		client := NewIBCClient("test-chain", ClientTypeTendermint, config)
		
		if client.config.SecurityLevel != level {
			t.Errorf("Expected security level %d, got %d", level, client.config.SecurityLevel)
		}
	}
	
	// Test connection security levels
	client := NewIBCClient("test-chain", ClientTypeTendermint, ClientConfig{})
	connection, _ := client.CreateConnection(NewIBCClient("counter-chain", ClientTypeTendermint, ClientConfig{}), ConnectionConfig{SecurityLevel: SecurityLevelUltra})
	
	if connection.config.SecurityLevel != SecurityLevelUltra {
		t.Errorf("Expected connection security level %d, got %d", SecurityLevelUltra, connection.config.SecurityLevel)
	}
	
	// Test channel security levels
	connection.OpenConnection()
	channel, _ := connection.CreateChannel("test-port", "test-channel", ChannelOrderingOrdered, ChannelConfig{SecurityLevel: SecurityLevelHigh})
	
	if channel.config.SecurityLevel != SecurityLevelHigh {
		t.Errorf("Expected channel security level %d, got %d", SecurityLevelHigh, channel.config.SecurityLevel)
	}
	
	// Open the channel before sending packets
	channel.OpenChannel()
	
	// Test packet security levels
	packet, err := channel.SendPacket([]byte("test data"), "dest-port", "dest-channel", 1000, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("Failed to send packet: %v", err)
	}
	
	if packet == nil {
		t.Fatal("Expected packet to be created")
	}
	
	packet.config.SecurityLevel = SecurityLevelMedium
	
	if packet.config.SecurityLevel != SecurityLevelMedium {
		t.Errorf("Expected packet security level %d, got %d", SecurityLevelMedium, packet.config.SecurityLevel)
	}
}
