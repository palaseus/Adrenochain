//go:build go1.20

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/palaseus/adrenochain/pkg/api"
	"github.com/palaseus/adrenochain/pkg/block"
	"github.com/palaseus/adrenochain/pkg/chain"
	"github.com/palaseus/adrenochain/pkg/consensus"
	"github.com/palaseus/adrenochain/pkg/logger"
	"github.com/palaseus/adrenochain/pkg/mempool"
	"github.com/palaseus/adrenochain/pkg/miner"
	"github.com/palaseus/adrenochain/pkg/monitoring"
	netpkg "github.com/palaseus/adrenochain/pkg/net"
	proto_net "github.com/palaseus/adrenochain/pkg/proto/net"
	"github.com/palaseus/adrenochain/pkg/storage"
	"github.com/palaseus/adrenochain/pkg/utxo"
	"github.com/palaseus/adrenochain/pkg/wallet"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/proto"
)

var (
	configFile string
	port       int
	mining     bool
	network    string
	walletFile string // New global flag
	passphrase string // New global flag
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "adrenochain",
		Short: "adrenochain - A modular blockchain implementation in Go",
		Long: `adrenochain is a modular blockchain implementation written in Go.
It features proof-of-work consensus, P2P networking, transaction mempool,
and wallet functionality.`,
		RunE: runNode,
	}

	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().IntVar(&port, "port", 0, "network port (0 for random)")
	rootCmd.PersistentFlags().BoolVar(&mining, "mining", false, "enable mining")
	rootCmd.PersistentFlags().StringVar(&network, "network", "mainnet", "network type (mainnet, testnet, devnet)")
	rootCmd.PersistentFlags().StringVar(&walletFile, "wallet-file", "wallet.dat", "path to wallet file")   // New flag
	rootCmd.PersistentFlags().StringVar(&passphrase, "passphrase", "", "passphrase for wallet encryption") // New flag

	// Add subcommands
	rootCmd.AddCommand(createWalletCmd())
	rootCmd.AddCommand(createTransactionCmd())
	rootCmd.AddCommand(getBalanceCmd())
	rootCmd.AddCommand(getBlockchainInfoCmd())
	rootCmd.AddCommand(getSafeInfoCmd()) // Add new safe command

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runNode(cmd *cobra.Command, args []string) error {
	// Load configuration
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Starting adrenochain node...")
	fmt.Printf("Network: %s\n", network)
	fmt.Printf("Port: %d\n", port)
	fmt.Printf("Mining: %t\n", mining)

	// Create blockchain components
	storageFactory := storage.NewStorageFactory()

	// Determine storage type from config or use default
	storageType := storage.StorageTypeFile // Default to file storage
	configDBType := viper.GetString("storage.db_type")

	if configDBType == "leveldb" {
		storageType = storage.StorageTypeLevelDB
	}

	dataDir := viper.GetString("storage.data_dir")
	if dataDir == "" {
		dataDir = "./data"
	}

	nodeStorage, err := storageFactory.CreateStorage(storageType, dataDir)
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}
	defer nodeStorage.Close()

	chainConfig := chain.DefaultChainConfig()
	consensusConfig := consensus.DefaultConsensusConfig()
	chain, err := chain.NewChain(chainConfig, consensusConfig, nodeStorage)
	if err != nil {
		return fmt.Errorf("failed to create chain: %w", err)
	}

	mempoolConfig := mempool.DefaultMempoolConfig()
	mempool := mempool.NewMempool(mempoolConfig)

	minerConfig := miner.DefaultMinerConfig()
	minerConfig.MiningEnabled = mining
	minerConfig.CoinbaseAddress = "miner_reward"
	miner := miner.NewMiner(chain, mempool, minerConfig, consensusConfig)

	networkConfig := netpkg.DefaultNetworkConfig()
	networkConfig.ListenPort = port
	networkConfig.EnableMDNS = true
	networkConfig.MaxPeers = 50

	net, err := netpkg.NewNetwork(networkConfig, chain, mempool)
	if err != nil {
		return fmt.Errorf("failed to create network: %w", err)
	}

	// Set up logging
	logger := setupLogger()

	// Set up monitoring service
	var monitoringService *monitoring.Service
	if viper.GetBool("monitoring.enabled") {
		monitoringConfig := createMonitoringConfig()
		monitoringService = monitoring.NewService(monitoringConfig, chain, mempool, net)

		// Start monitoring service
		if err := monitoringService.Start(); err != nil {
			logger.Error("Failed to start monitoring service: %v", err)
		} else {
			logger.Info("Monitoring service started")
			logger.Info("Metrics endpoint: %s", monitoringService.GetMetricsEndpoint())
			logger.Info("Health endpoint: %s", monitoringService.GetHealthEndpoint())
			if prometheusEndpoint := monitoringService.GetPrometheusEndpoint(); prometheusEndpoint != "" {
				logger.Info("Prometheus endpoint: %s", prometheusEndpoint)
			}
		}

		// Set up mining callback for monitoring
		miner.SetOnBlockMined(func(minedBlock *block.Block) {
			// Update mining metrics when a block is successfully mined
			monitoringService.GetMetrics().UpdateBlocksMined(1)
			monitoringService.GetMetrics().UpdateBlockHeight(int64(minedBlock.Header.Height))
			monitoringService.GetMetrics().UpdateLastBlockTime(minedBlock.Header.Timestamp)

			// Update difficulty metrics
			monitoringService.GetMetrics().UpdateChainDifficulty(float64(minedBlock.Header.Difficulty))

			// Update transaction metrics
			txnCount := len(minedBlock.Transactions)
			if txnCount > 0 {
				monitoringService.GetMetrics().UpdateTotalTxns(int64(txnCount))
				monitoringService.GetMetrics().UpdateAvgTxnPerBlock(float64(txnCount))
			}

			// Update block size metrics
			blockSize := int64(len(minedBlock.Transactions) * 256) // Rough estimate
			monitoringService.GetMetrics().UpdateAvgBlockSize(blockSize)

			// Log the successful mining
			logger.Info("Block successfully mined and added to chain: Height=%d, Hash=%x, Transactions=%d",
				minedBlock.Header.Height, minedBlock.CalculateHash(), txnCount)
		})
	}

	// Set up network message handlers
	blockSub, err := net.SubscribeToBlocks()
	if err != nil {
		logger.Error("Failed to subscribe to blocks: %v", err)
		return fmt.Errorf("failed to subscribe to blocks: %w", err)
	}
	defer blockSub.Cancel() // Ensure subscription is cancelled on shutdown

	// Create context for goroutines
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Enhanced block processing with better monitoring integration
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg, err := blockSub.Next(net.GetContext())
				if err != nil {
					if err == context.Canceled {
						return
					}
					logger.Error("Error receiving block: %v", err)
					if monitoringService != nil {
						monitoringService.GetMetrics().IncrementNetworkErrors()
					}
					continue
				}

				var networkMsg proto_net.Message
				if err := proto.Unmarshal(msg.Data, &networkMsg); err != nil {
					logger.Error("Failed to unmarshal network message for block: %v", err)
					if monitoringService != nil {
						monitoringService.GetMetrics().IncrementValidationErrors()
					}
					continue
				}

				// Verify message signature
				pubKey, err := peer.ID(networkMsg.FromPeerId).ExtractPublicKey()
				if err != nil {
					logger.Error("Error extracting public key for block message: %v", err)
					if monitoringService != nil {
						monitoringService.GetMetrics().IncrementValidationErrors()
					}
					continue
				}
				tempMsg := proto.Clone(&networkMsg).(*proto_net.Message)
				tempMsg.Signature = nil // Clear the signature for verification
				dataToVerify, err := proto.Marshal(tempMsg)
				if err != nil {
					logger.Error("Error marshaling block message for verification: %v", err)
					if monitoringService != nil {
						monitoringService.GetMetrics().IncrementValidationErrors()
					}
					continue
				}
				verified, err := pubKey.Verify(dataToVerify, networkMsg.Signature)
				if err != nil {
					logger.Error("Error verifying block message signature: %v", err)
					if monitoringService != nil {
						monitoringService.GetMetrics().IncrementValidationErrors()
					}
					continue
				}
				if !verified {
					logger.Error("Invalid block message signature from %s", peer.ID(networkMsg.FromPeerId).String())
					if monitoringService != nil {
						monitoringService.GetMetrics().IncrementValidationErrors()
					}
					continue
				}

				// Handle block message content
				switch content := networkMsg.Content.(type) {
				case *proto_net.Message_BlockMessage:
					var block block.Block
					if err := json.Unmarshal(content.BlockMessage.BlockData, &block); err != nil {
						logger.Error("Failed to unmarshal block from payload: %v", err)
						if monitoringService != nil {
							monitoringService.GetMetrics().IncrementValidationErrors()
						}
						continue
					}

					// Record block processing start time for metrics
					startTime := time.Now()

					logger.Info("Received block from network: %s", block.String())
					if err := chain.AddBlock(&block); err != nil {
						logger.Error("Failed to add received block: %v", err)
						if monitoringService != nil {
							monitoringService.GetMetrics().IncrementRejectedBlocks()
							monitoringService.GetMetrics().IncrementErrors()
						}
					} else {
						if monitoringService != nil {
							monitoringService.GetMetrics().UpdateTotalBlocks(int64(chain.GetHeight() + 1))
							monitoringService.GetMetrics().UpdateBlockHeight(int64(chain.GetHeight()))
							monitoringService.GetMetrics().UpdateLastBlockTime(block.Header.Timestamp)

							// Update block processing time
							processingTime := time.Since(startTime)
							monitoringService.GetMetrics().UpdateBlockProcessingTime(processingTime)

							// Update transaction metrics
							txnCount := len(block.Transactions)
							if txnCount > 0 {
								monitoringService.GetMetrics().UpdateTotalTxns(int64(txnCount))
								monitoringService.GetMetrics().UpdateAvgTxnPerBlock(float64(txnCount))
							}

							// Update block size metrics (rough estimate)
							blockSize := int64(len(block.Transactions) * 256) // Rough estimate
							monitoringService.GetMetrics().UpdateAvgBlockSize(blockSize)
						}
					}
				default:
					logger.Error("Received unknown message type for block subscription: %T", content)
					if monitoringService != nil {
						monitoringService.GetMetrics().IncrementValidationErrors()
					}
					continue
				}
			}
		}
	}()

	txSub, err := net.SubscribeToTransactions()
	if err != nil {
		logger.Error("Failed to subscribe to transactions: %v", err)
		return fmt.Errorf("failed to subscribe to transactions: %w", err)
	}
	defer txSub.Cancel() // Ensure subscription is cancelled on shutdown

	// Enhanced transaction processing with better monitoring integration
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg, err := txSub.Next(net.GetContext())
				if err != nil {
					if err == context.Canceled {
						return
					}
					logger.Error("Error receiving transaction: %v", err)
					if monitoringService != nil {
						monitoringService.GetMetrics().IncrementNetworkErrors()
					}
					continue
				}

				var networkMsg proto_net.Message
				if err := proto.Unmarshal(msg.Data, &networkMsg); err != nil {
					logger.Error("Failed to unmarshal network message for transaction: %v", err)
					if monitoringService != nil {
						monitoringService.GetMetrics().IncrementValidationErrors()
					}
					continue
				}

				// Verify message signature
				pubKey, err := peer.ID(networkMsg.FromPeerId).ExtractPublicKey()
				if err != nil {
					logger.Error("Error extracting public key for transaction message: %v", err)
					if monitoringService != nil {
						monitoringService.GetMetrics().IncrementValidationErrors()
					}
					continue
				}
				tempMsg := proto.Clone(&networkMsg).(*proto_net.Message)
				tempMsg.Signature = nil // Clear the signature for verification
				dataToVerify, err := proto.Marshal(tempMsg)
				if err != nil {
					logger.Error("Error marshaling transaction message for verification: %v", err)
					if monitoringService != nil {
						monitoringService.GetMetrics().IncrementValidationErrors()
					}
					continue
				}
				verified, err := pubKey.Verify(dataToVerify, networkMsg.Signature)
				if err != nil {
					logger.Error("Error verifying transaction message signature: %v", err)
					if monitoringService != nil {
						monitoringService.GetMetrics().IncrementValidationErrors()
					}
					continue
				}
				if !verified {
					senderPeerID, err := peer.IDFromBytes(networkMsg.FromPeerId)
					if err != nil {
						logger.Error("Failed to get peer ID from bytes: %v", err)
						if monitoringService != nil {
							monitoringService.GetMetrics().IncrementValidationErrors()
						}
						continue
					}
					logger.Error("Invalid transaction message signature from %s", senderPeerID.String())
					if monitoringService != nil {
						monitoringService.GetMetrics().IncrementValidationErrors()
					}
					continue
				}

				// Handle transaction message content
				switch content := networkMsg.Content.(type) {
				case *proto_net.Message_TransactionMessage:
					var tx block.Transaction
					if err := json.Unmarshal(content.TransactionMessage.TransactionData, &tx); err != nil {
						logger.Error("Failed to unmarshal transaction from payload: %v", err)
						if monitoringService != nil {
							monitoringService.GetMetrics().IncrementValidationErrors()
						}
						continue
					}

					// Record transaction processing start time for metrics
					startTime := time.Now()

					logger.Info("Received transaction from network: %s", tx.String())
					if err := mempool.AddTransaction(&tx); err != nil {
						logger.Error("Failed to add received transaction: %v", err)
						if monitoringService != nil {
							monitoringService.GetMetrics().IncrementRejectedTxns()
							monitoringService.GetMetrics().IncrementErrors()
						}
					} else {
						if monitoringService != nil {
							// Update transaction metrics
							monitoringService.GetMetrics().UpdateTotalTxns(int64(mempool.GetTransactionCount()))
							monitoringService.GetMetrics().UpdatePendingTxns(int64(mempool.GetTransactionCount()))

							// Update transaction processing time
							processingTime := time.Since(startTime)
							monitoringService.GetMetrics().UpdateTxnProcessingTime(processingTime)
						}
					}
				default:
					logger.Error("Received unknown message type for transaction subscription: %T", content)
					if monitoringService != nil {
						monitoringService.GetMetrics().IncrementValidationErrors()
					}
					continue
				}
			}
		}
	}()

	// Start mining if enabled
	if mining {
		if err := miner.StartMining(); err != nil {
			return fmt.Errorf("failed to start mining: %w", err)
		}
		logger.Info("Mining started")

		// Update mining metrics
		if monitoringService != nil {
			monitoringService.GetMetrics().SetMiningEnabled(true)
		}
	}

	// Start API server if enabled
	var apiServer *api.Server
	if viper.GetBool("api.enabled") {
		apiAddr := viper.GetString("api.listen_addr")
		apiPort := 8080 // Default port

		// Parse port from address string (e.g., "127.0.0.1:8080" -> 8080)
		if apiAddr != "" && apiAddr != "127.0.0.1:8080" {
			// Extract port from address
			if len(apiAddr) > 0 {
				// Simple port extraction - in production you'd want more robust parsing
				for i := len(apiAddr) - 1; i >= 0; i-- {
					if apiAddr[i] == ':' {
						if portStr := apiAddr[i+1:]; portStr != "" {
							if port, err := strconv.Atoi(portStr); err == nil {
								apiPort = port
							}
						}
						break
					}
				}
			}
		}

		// Create a dummy wallet for API (in a real implementation, this would be the actual wallet)
		dummyWallet := &wallet.Wallet{}

		apiConfig := &api.ServerConfig{
			Port:   apiPort,
			Chain:  chain,
			Wallet: dummyWallet,
		}

		apiServer = api.NewServer(apiConfig)

		// Start API server in background
		go func() {
			if err := apiServer.Start(); err != nil {
				logger.Error("API server error: %v", err)
			}
		}()

		logger.Info("API server started on port %d", apiPort)
	}

	// Start periodic status updates with enhanced monitoring
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				bestBlock := chain.GetBestBlock()
				peerCount := len(net.GetPeers())
				mempoolCount := mempool.GetTransactionCount()

				if bestBlock != nil && bestBlock.Header != nil {
					logger.Info("Status: Height=%d, Hash=%x, Peers=%d, Mempool=%d",
						chain.GetHeight(),
						bestBlock.CalculateHash(),
						peerCount,
						mempoolCount)
				} else {
					logger.Info("Status: Height=%d, Peers=%d, Mempool=%d",
						chain.GetHeight(),
						peerCount,
						mempoolCount)
				}

				// Update network metrics
				if monitoringService != nil {
					monitoringService.GetMetrics().UpdateConnectedPeers(int64(peerCount))
					monitoringService.GetMetrics().UpdatePendingTxns(int64(mempoolCount))

					// Update chain size metrics
					if bestBlock != nil {
						monitoringService.GetMetrics().UpdateChainSize(int64(chain.GetHeight() * 1024)) // Rough estimate
					}
				}
			}
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down adrenochain node...")

	// Cancel context to stop all goroutines
	cancel()

	// Cleanup
	if mining {
		miner.StopMining()
		// Update mining metrics on shutdown
		if monitoringService != nil {
			monitoringService.GetMetrics().SetMiningEnabled(false)
		}
	}
	miner.Close()
	net.Close()

	// Stop monitoring service if it was started
	if monitoringService != nil {
		logger.Info("Stopping monitoring service...")
		if err := monitoringService.Stop(); err != nil {
			logger.Error("Failed to stop monitoring service: %v", err)
		} else {
			logger.Info("Monitoring service stopped successfully")
		}
	}

	// Close API server if it was started
	if apiServer != nil {
		// Note: The API server doesn't have a Close method yet, but we could add one if needed
		logger.Info("API server stopped")
	}

	logger.Info("adrenochain node stopped")
	return nil
}

func loadConfig() error {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	return nil
}

func createWalletCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "wallet",
		Short: "Create or load a wallet",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create storage for wallet
			walletStorageConfig := storage.DefaultStorageConfig().WithDataDir("./wallet_data")
			walletStorage, err := storage.NewStorage(walletStorageConfig)
			if err != nil {
				return fmt.Errorf("failed to create wallet storage: %w", err)
			}
			defer walletStorage.Close()

			walletConfig := wallet.DefaultWalletConfig()
			walletConfig.WalletFile = walletFile
			walletConfig.Passphrase = passphrase

			us := utxo.NewUTXOSet() // Still a dummy UTXOSet for CLI commands
			wallet, err := wallet.NewWallet(walletConfig, us, walletStorage)
			if err != nil {
				return fmt.Errorf("failed to create/load wallet: %w", err)
			}

			// Save the wallet after creation/loading (important for new wallets)
			if err := wallet.Save(); err != nil {
				return fmt.Errorf("failed to save wallet: %w", err)
			}

			account := wallet.GetDefaultAccount()
			fmt.Printf("Wallet created/loaded successfully!\n")
			fmt.Printf("Default account address: %s\n", account.Address)
			fmt.Printf("Public key: %x\n", account.PublicKey)

			return nil
		},
	}
}

func createTransactionCmd() *cobra.Command {
	var from, to string
	var amount, fee uint64

	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send a transaction",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create storage for wallet
			walletStorageConfig := storage.DefaultStorageConfig().WithDataDir("./wallet_data")
			walletStorage, err := storage.NewStorage(walletStorageConfig)
			if err != nil {
				return fmt.Errorf("failed to create wallet storage: %w", err)
			}
			defer walletStorage.Close()

			walletConfig := wallet.DefaultWalletConfig()
			walletConfig.WalletFile = walletFile
			walletConfig.Passphrase = passphrase

			us := utxo.NewUTXOSet() // Still a dummy UTXOSet for CLI commands
			wallet, err := wallet.NewWallet(walletConfig, us, walletStorage)
			if err != nil {
				return fmt.Errorf("failed to load wallet: %w", err)
			}

			tx, err := wallet.CreateTransaction(from, to, amount, fee)
			if err != nil {
				return fmt.Errorf("failed to create transaction: %w", err)
			}

			// Save the wallet after transaction (to update nonce)
			if err := wallet.Save(); err != nil {
				return fmt.Errorf("failed to save wallet: %w", err)
			}

			fmt.Printf("Transaction created successfully!\n")
			fmt.Printf("Hash: %x\n", tx.Hash)
			fmt.Printf("From: %s\n", from)
			fmt.Printf("To: %s\n", to)
			fmt.Printf("Amount: %d\n", amount)
			fmt.Printf("Fee: %d\n", fee)

			return nil
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "sender address")
	cmd.Flags().StringVar(&to, "to", "", "recipient address")
	cmd.Flags().Uint64Var(&amount, "amount", 0, "amount to send")
	cmd.Flags().Uint64Var(&fee, "fee", 0, "transaction fee")

	cmd.MarkFlagRequired("from")
	cmd.MarkFlagRequired("to")
	cmd.MarkFlagRequired("amount")

	return cmd
}

func getBalanceCmd() *cobra.Command {
	var address string

	cmd := &cobra.Command{
		Use:   "balance",
		Short: "Get account balance",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create storage for wallet
			walletStorageConfig := storage.DefaultStorageConfig().WithDataDir("./wallet_data")
			walletStorage, err := storage.NewStorage(walletStorageConfig)
			if err != nil {
				return fmt.Errorf("failed to create wallet storage: %w", err)
			}
			defer walletStorage.Close()

			walletConfig := wallet.DefaultWalletConfig()
			walletConfig.WalletFile = walletFile
			walletConfig.Passphrase = passphrase

			us := utxo.NewUTXOSet() // Still a dummy UTXOSet for CLI commands
			wallet, err := wallet.NewWallet(walletConfig, us, walletStorage)
			if err != nil {
				return fmt.Errorf("failed to load wallet: %w", err)
			}

			balance := wallet.GetBalance(address)
			fmt.Printf("Balance for %s: %d\n", address, balance)

			return nil
		},
	}

	cmd.Flags().StringVar(&address, "address", "", "account address")
	cmd.MarkFlagRequired("address")

	return cmd
}

func getBlockchainInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Get blockchain information",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration to determine storage type
			if err := loadConfig(); err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create storage using factory
			storageFactory := storage.NewStorageFactory()

			// Determine storage type from config or use default
			storageType := storage.StorageTypeFile // Default to file storage
			configDBType := viper.GetString("storage.db_type")
			if configDBType == "leveldb" {
				storageType = storage.StorageTypeLevelDB
			}

			dataDir := viper.GetString("storage.data_dir")
			if dataDir == "" {
				dataDir = "./data"
			}

			// Create storage
			nodeStorage, err := storageFactory.CreateStorage(storageType, dataDir)
			if err != nil {
				return fmt.Errorf("failed to create storage: %w", err)
			}

			// MINIMAL APPROACH: Read chainstate directly without loading full chain
			fmt.Printf("Blockchain Information:\n")
			
			// Read chainstate directly
			chainState, err := nodeStorage.GetChainState()
			if err != nil {
				fmt.Printf("Height: 0 (No chain state found)\n")
				fmt.Printf("Best Block Hash: Not available\n")
			} else {
				fmt.Printf("Height: %d\n", chainState.Height)
				if len(chainState.BestBlockHash) > 0 {
					fmt.Printf("Best Block Hash: %x\n", chainState.BestBlockHash)
				} else {
					fmt.Printf("Best Block Hash: Not available\n")
				}
			}
			
			// Count block files
			blockCount := 0
			if entries, err := os.ReadDir(dataDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() && len(entry.Name()) == 64 { // Block files are 64 chars
						blockCount++
					}
				}
			}
			fmt.Printf("Block Files: %d\n", blockCount)
			
			// Storage information
			fmt.Printf("Storage Type: %s\n", storageType)
			fmt.Printf("Data Directory: %s\n", dataDir)

			return nil
		},
	}
}

func getSafeInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "safe-info",
		Short: "Get safe blockchain information",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration to determine storage type
			if err := loadConfig(); err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create storage using factory
			storageFactory := storage.NewStorageFactory()

			// Determine storage type from config or use default
			storageType := storage.StorageTypeFile // Default to file storage
			configDBType := viper.GetString("storage.db_type")
			if configDBType == "leveldb" {
				storageType = storage.StorageTypeLevelDB
			}

			dataDir := viper.GetString("storage.data_dir")
			if dataDir == "" {
				dataDir = "./data"
			}

			// Create storage
			nodeStorage, err := storageFactory.CreateStorage(storageType, dataDir)
			if err != nil {
				return fmt.Errorf("failed to create storage: %w", err)
			}

			// MINIMAL APPROACH: Read chainstate directly without loading full chain
			fmt.Printf("Safe Blockchain Information:\n")
			
			// Read chainstate directly
			chainState, err := nodeStorage.GetChainState()
			if err != nil {
				fmt.Printf("Height: 0 (No chain state found)\n")
				fmt.Printf("Best Block Hash: Not available\n")
			} else {
				fmt.Printf("Height: %d\n", chainState.Height)
				if len(chainState.BestBlockHash) > 0 {
					fmt.Printf("Best Block Hash: %x\n", chainState.BestBlockHash)
				} else {
					fmt.Printf("Best Block Hash: Not available\n")
				}
			}
			
			// Count block files
			blockCount := 0
			if entries, err := os.ReadDir(dataDir); err == nil {
				for _, entry := range entries {
					if !entry.IsDir() && len(entry.Name()) == 64 { // Block files are 64 chars
						blockCount++
					}
				}
			}
			fmt.Printf("Block Files: %d\n", blockCount)
			
			// Storage information
			fmt.Printf("Storage Type: %s\n", storageType)
			fmt.Printf("Data Directory: %s\n", dataDir)

			return nil
		},
	}
}

// setupLogger creates and configures the logger based on configuration
func setupLogger() *logger.Logger {
	logLevel := logger.INFO
	if levelStr := viper.GetString("logging.level"); levelStr != "" {
		switch strings.ToLower(levelStr) {
		case "debug":
			logLevel = logger.DEBUG
		case "info":
			logLevel = logger.INFO
		case "warn":
			logLevel = logger.WARN
		case "error":
			logLevel = logger.ERROR
		}
	}

	logFormat := viper.GetString("logging.format")
	useJSON := strings.ToLower(logFormat) == "json"

	logFile := viper.GetString("logging.log_file")
	maxSize := viper.GetInt64("logging.max_size")
	maxBackups := viper.GetInt("logging.max_backups")

	// Set defaults if not specified
	if maxSize == 0 {
		maxSize = 100 * 1024 * 1024 // 100MB
	}
	if maxBackups == 0 {
		maxBackups = 5
	}

	logConfig := &logger.Config{
		Level:      logLevel,
		Prefix:     "adrenochain",
		UseJSON:    useJSON,
		LogFile:    logFile,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
	}

	return logger.NewLogger(logConfig)
}

// createMonitoringConfig creates monitoring configuration from viper settings
func createMonitoringConfig() *monitoring.Config {
	// Parse metrics port from address
	metricsAddr := viper.GetString("monitoring.metrics.listen_addr")
	metricsPort := 9090
	if metricsAddr != "" {
		if parts := strings.Split(metricsAddr, ":"); len(parts) > 1 {
			if port, err := strconv.Atoi(parts[1]); err == nil {
				metricsPort = port
			}
		}
	}

	// Parse health port from address
	healthAddr := viper.GetString("monitoring.health.listen_addr")
	healthPort := 8081
	if healthAddr != "" {
		if parts := strings.Split(healthAddr, ":"); len(parts) > 1 {
			if port, err := strconv.Atoi(parts[1]); err == nil {
				healthPort = port
			}
		}
	}

	// Parse intervals
	collectInterval := viper.GetDuration("monitoring.metrics.collect_interval")
	if collectInterval == 0 {
		collectInterval = 15 * time.Second
	}

	healthCheckInterval := viper.GetDuration("monitoring.health.check_interval")
	if healthCheckInterval == 0 {
		healthCheckInterval = 15 * time.Second
	}

	// Parse log level for monitoring
	monitoringLogLevel := logger.INFO
	if levelStr := viper.GetString("monitoring.logging.level"); levelStr != "" {
		switch strings.ToLower(levelStr) {
		case "debug":
			monitoringLogLevel = logger.DEBUG
		case "info":
			monitoringLogLevel = logger.INFO
		case "warn":
			monitoringLogLevel = logger.WARN
		case "error":
			monitoringLogLevel = logger.ERROR
		}
	}

	monitoringLogFormat := viper.GetString("monitoring.logging.format")
	monitoringUseJSON := strings.ToLower(monitoringLogFormat) == "json"

	monitoringLogFile := viper.GetString("monitoring.logging.log_file")
	monitoringMaxSize := viper.GetInt64("monitoring.logging.max_size")
	monitoringMaxBackups := viper.GetInt("monitoring.logging.max_backups")

	// Set defaults if not specified
	if monitoringMaxSize == 0 {
		monitoringMaxSize = 50 * 1024 * 1024 // 50MB
	}
	if monitoringMaxBackups == 0 {
		monitoringMaxBackups = 3
	}

	return &monitoring.Config{
		MetricsPort:         metricsPort,
		HealthPort:          healthPort,
		LogLevel:            monitoringLogLevel,
		LogJSON:             monitoringUseJSON,
		LogFile:             monitoringLogFile,
		MetricsPath:         "/metrics",
		HealthPath:          "/health",
		PrometheusPath:      "/prometheus",
		CollectInterval:     collectInterval,
		HealthCheckInterval: healthCheckInterval,
		EnablePrometheus:    viper.GetBool("monitoring.metrics.prometheus_enabled"),
	}
}
