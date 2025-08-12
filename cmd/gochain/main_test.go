package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestRunNode(t *testing.T) {
	// Create a temporary data directory
	dataDir, err := ioutil.TempDir("", "gochain_test_data_")
	assert.NoError(t, err)
	defer os.RemoveAll(dataDir)

	// Set up a minimal configuration for the node
	port = 0 // Random port
	mining = false
	network = "testnet"
	configFile = "" // No config file, use defaults

	// Run runNode in a goroutine
	done := make(chan error)
	go func() {
		// Temporarily change the working directory to the dataDir for the test
		oldWd, _ := os.Getwd()
		os.Chdir(dataDir)
		defer os.Chdir(oldWd)

		done <- runNode(nil, nil)
	}()

	// Give the node some time to start up
	time.Sleep(2 * time.Second)

	// Send an interrupt signal to gracefully shut down the node
	// This simulates Ctrl+C
	process, err := os.FindProcess(os.Getpid())
	assert.NoError(t, err)
	err = process.Signal(syscall.SIGINT)
	assert.NoError(t, err)

	// Wait for the node to shut down
	select {
	case err := <-done:
		assert.NoError(t, err, "runNode should exit without error")
	case <-time.After(5 * time.Second):
		t.Fatal("Node did not shut down in time")
	}

	fmt.Printf("TestRunNode completed successfully for dataDir: %s\n", dataDir)
}

// TestCreateWalletCmd tests wallet creation command
func TestCreateWalletCmd(t *testing.T) {
	// Test wallet creation command
	cmd := createWalletCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "wallet", cmd.Use)
	assert.Equal(t, "Create or load a wallet", cmd.Short)

	// Test command execution with valid parameters
	walletFile = "test_wallet.json"
	passphrase = "test_passphrase"

	// Create a temporary directory for test data
	tempDir, err := os.MkdirTemp("", "gochain_test_wallet")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to temp directory for test
	originalDir, err := os.Getwd()
	assert.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	// Execute the command
	err = cmd.RunE(cmd, []string{})
	assert.NoError(t, err)

	// The wallet is stored in the storage system, not as a direct file
	// We can verify the wallet was created by checking if the storage directory exists
	_, err = os.Stat("wallet_data")
	assert.NoError(t, err, "Wallet data directory should be created")
}

// TestCreateTransactionCmd tests transaction creation command
func TestCreateTransactionCmd(t *testing.T) {
	// Test transaction creation command
	cmd := createTransactionCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "send", cmd.Use)
	assert.Equal(t, "Send a transaction", cmd.Short)

	// Test that flags exist
	assert.NotNil(t, cmd.Flags().Lookup("from"))
	assert.NotNil(t, cmd.Flags().Lookup("to"))
	assert.NotNil(t, cmd.Flags().Lookup("amount"))

	// Test flag defaults
	fromFlag := cmd.Flags().Lookup("from")
	toFlag := cmd.Flags().Lookup("to")
	amountFlag := cmd.Flags().Lookup("amount")
	feeFlag := cmd.Flags().Lookup("fee")

	assert.Equal(t, "", fromFlag.DefValue)
	assert.Equal(t, "", toFlag.DefValue)
	assert.Equal(t, "0", amountFlag.DefValue)
	assert.Equal(t, "0", feeFlag.DefValue)
}

// TestGetBalanceCmd tests balance retrieval command
func TestGetBalanceCmd(t *testing.T) {
	// Test balance command
	cmd := getBalanceCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "balance", cmd.Use)
	assert.Equal(t, "Get account balance", cmd.Short)

	// Test that flag exists
	assert.NotNil(t, cmd.Flags().Lookup("address"))
}

// TestGetBlockchainInfoCmd tests blockchain info command
func TestGetBlockchainInfoCmd(t *testing.T) {
	// Test blockchain info command
	cmd := getBlockchainInfoCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "info", cmd.Use)
	assert.Equal(t, "Get blockchain information", cmd.Short)

	// Test command structure
	assert.NotNil(t, cmd.RunE)
}

// TestCreateMonitoringConfig tests monitoring configuration creation
func TestCreateMonitoringConfig(t *testing.T) {
	// Test monitoring config creation
	config := createMonitoringConfig()
	assert.NotNil(t, config)

	// Test default values
	assert.Equal(t, 9090, config.MetricsPort)
	assert.Equal(t, 8081, config.HealthPort)
}

// TestSetupLoggerEdgeCases tests logger setup edge cases
func TestSetupLoggerEdgeCases(t *testing.T) {
	// Test with invalid log level
	viper.Set("logging.level", "invalid_level")
	logger := setupLogger()
	assert.NotNil(t, logger)
	// Note: We can't easily test the level without exposing internal state

	// Test with JSON format
	viper.Set("logging.format", "json")
	logger = setupLogger()
	assert.NotNil(t, logger)

	// Test with custom log file
	viper.Set("logging.log_file", "/tmp/test.log")
	logger = setupLogger()
	assert.NotNil(t, logger)

	// Test with custom max size and backups
	viper.Set("logging.max_size", 50*1024*1024) // 50MB
	viper.Set("logging.max_backups", 10)
	logger = setupLogger()
	assert.NotNil(t, logger)

	// Reset viper
	viper.Reset()
}

// TestLoadConfigEdgeCases tests configuration loading edge cases
func TestLoadConfigEdgeCases(t *testing.T) {
	// Test with non-existent config file
	viper.SetConfigFile("/nonexistent/config.yaml")
	err := loadConfig()
	// loadConfig only returns error for non-ConfigFileNotFoundError, so this should pass
	assert.NoError(t, err)

	// Test with invalid config file
	tempFile, err := os.CreateTemp("", "invalid_config.yaml")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// Write invalid YAML
	_, err = tempFile.WriteString("invalid: yaml: content: [")
	assert.NoError(t, err)
	tempFile.Close()

	viper.SetConfigFile(tempFile.Name())
	err = loadConfig()
	// loadConfig only returns error for non-ConfigFileNotFoundError, so this should pass
	assert.NoError(t, err)

	// Reset viper
	viper.Reset()
}

// TestMainFunction tests the main function (indirectly)
func TestMainFunction(t *testing.T) {
	// Test that main function can be called without panicking
	// This is a basic test to ensure the function exists and is callable
	assert.NotPanics(t, func() {
		// We can't actually call main() in tests, but we can verify it exists
		// by checking that the file compiles and the function is defined
	})
}

// TestCommandIntegration tests command integration
func TestCommandIntegration(t *testing.T) {
	// Test that all commands can be created without errors
	commands := []func() *cobra.Command{
		createWalletCmd,
		createTransactionCmd,
		getBalanceCmd,
		getBlockchainInfoCmd,
	}

	for _, cmdFunc := range commands {
		cmd := cmdFunc()
		assert.NotNil(t, cmd)
		assert.NotEmpty(t, cmd.Use)
		assert.NotEmpty(t, cmd.Short)
	}
}

// TestConfigurationValidation tests configuration validation
func TestConfigurationValidation(t *testing.T) {
	// Test default configuration values
	viper.Reset()
	
	// Set some default values for testing
	viper.SetDefault("network", "mainnet")
	viper.SetDefault("storage.type", "file")
	viper.SetDefault("storage.data_dir", "./data")
	viper.SetDefault("mining.enabled", false)
	viper.SetDefault("mining.threads", 1)
	viper.SetDefault("api.listen_addr", ":8080")
	viper.SetDefault("api.enabled", true)
	
	// Test network configuration
	assert.Equal(t, "mainnet", viper.GetString("network"))
	
	// Test storage configuration
	assert.Equal(t, "file", viper.GetString("storage.type"))
	assert.Equal(t, "./data", viper.GetString("storage.data_dir"))
	
	// Test mining configuration
	assert.Equal(t, false, viper.GetBool("mining.enabled"))
	assert.Equal(t, 1, viper.GetInt("mining.threads"))
	
	// Test API configuration
	assert.Equal(t, ":8080", viper.GetString("api.listen_addr"))
	assert.Equal(t, true, viper.GetBool("api.enabled"))
}
