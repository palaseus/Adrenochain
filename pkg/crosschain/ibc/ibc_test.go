package ibc

import (
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
