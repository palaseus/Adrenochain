//go:build go1.20

package main

import (
        "context"
        "encoding/json"
        "fmt"
        "os"
        "os/signal"
        "strconv"
        "syscall"
        "time"

        "github.com/gochain/gochain/pkg/api"
        "github.com/gochain/gochain/pkg/block"
        "github.com/gochain/gochain/pkg/chain"
        "github.com/gochain/gochain/pkg/consensus"
        "github.com/gochain/gochain/pkg/mempool"
        "github.com/gochain/gochain/pkg/miner"
        netpkg "github.com/gochain/gochain/pkg/net"
        proto_net "github.com/gochain/gochain/pkg/proto/net"
        "github.com/gochain/gochain/pkg/storage"
        "github.com/gochain/gochain/pkg/utxo"
        "github.com/gochain/gochain/pkg/wallet"
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
		Use:   "gochain",
		Short: "GoChain - A modular blockchain implementation in Go",
		Long: `GoChain is a modular blockchain implementation written in Go.
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

	fmt.Println("Starting GoChain node...")
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

	// Set up network message handlers
	blockSub, err := net.SubscribeToBlocks()
	if err != nil {
		return fmt.Errorf("failed to subscribe to blocks: %w", err)
	}
	defer blockSub.Cancel() // Ensure subscription is cancelled on shutdown

	// Create context for goroutines
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
					fmt.Printf("Error receiving block: %v\n", err)
					return
				}

				var networkMsg proto_net.Message
				if err := proto.Unmarshal(msg.Data, &networkMsg); err != nil {
					fmt.Printf("Failed to unmarshal network message for block: %v\n", err)
					continue
				}

				// Verify message signature
				pubKey, err := peer.ID(networkMsg.FromPeerId).ExtractPublicKey()
				if err != nil {
					fmt.Printf("Error extracting public key for block message: %v\n", err)
					continue
				}
				tempMsg := proto.Clone(&networkMsg).(*proto_net.Message)
				tempMsg.Signature = nil // Clear the signature for verification
				dataToVerify, err := proto.Marshal(tempMsg)
				if err != nil {
					fmt.Printf("Error marshaling block message for verification: %v\n", err)
					continue
				}
				verified, err := pubKey.Verify(dataToVerify, networkMsg.Signature)
				if err != nil {
					fmt.Printf("Error verifying block message signature: %v\n", err)
					continue
				}
				if !verified {
					fmt.Printf("Invalid block message signature from %s\n", peer.ID(networkMsg.FromPeerId).String())
					continue
				}

				// Handle block message content
				switch content := networkMsg.Content.(type) {
				case *proto_net.Message_BlockMessage:
					var block block.Block
					if err := json.Unmarshal(content.BlockMessage.BlockData, &block); err != nil {
						fmt.Printf("Failed to unmarshal block from payload: %v\n", err)
						continue
					}
					fmt.Printf("Received block from network: %s\n", block.String())
					if err := chain.AddBlock(&block); err != nil {
						fmt.Printf("Failed to add received block: %v\n", err)
					}
				default:
					fmt.Printf("Received unknown message type for block subscription: %T\n", content)
					continue
				}
			}
		}
	}()

	txSub, err := net.SubscribeToTransactions()
	if err != nil {
		return fmt.Errorf("failed to subscribe to transactions: %w", err)
	}
	defer txSub.Cancel() // Ensure subscription is cancelled on shutdown

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
					fmt.Printf("Error receiving transaction: %v\n", err)
					return
				}

				var networkMsg proto_net.Message
				if err := proto.Unmarshal(msg.Data, &networkMsg); err != nil {
					fmt.Printf("Failed to unmarshal network message for transaction: %v\n", err)
					continue
				}

				// Verify message signature
				pubKey, err := peer.ID(networkMsg.FromPeerId).ExtractPublicKey()
				if err != nil {
					fmt.Printf("Error extracting public key for transaction message: %v\n", err)
					continue
				}
				tempMsg := proto.Clone(&networkMsg).(*proto_net.Message)
				tempMsg.Signature = nil // Clear the signature for verification
				dataToVerify, err := proto.Marshal(tempMsg)
				if err != nil {
					fmt.Printf("Error marshaling transaction message for verification: %v\n", err)
					continue
				}
				verified, err := pubKey.Verify(dataToVerify, networkMsg.Signature)
				if err != nil {
					fmt.Printf("Error verifying transaction message signature: %v\n", err)
					continue
				}
				if !verified {
					senderPeerID, err := peer.IDFromBytes(networkMsg.FromPeerId)
					if err != nil {
						fmt.Printf("Failed to get peer ID from bytes: %v\n", err)
						continue
					}
					fmt.Printf("Invalid transaction message signature from %s\n", senderPeerID.String())
					continue
				}

				// Handle transaction message content
				switch content := networkMsg.Content.(type) {
				case *proto_net.Message_TransactionMessage:
					var tx block.Transaction
					if err := json.Unmarshal(content.TransactionMessage.TransactionData, &tx); err != nil {
						fmt.Printf("Failed to unmarshal transaction from payload: %v\n", err)
						continue
					}
					fmt.Printf("Received transaction from network: %s\n", tx.String())
					if err := mempool.AddTransaction(&tx); err != nil {
						fmt.Printf("Failed to add received transaction: %v\n", err)
					}
				default:
					fmt.Printf("Received unknown message type for transaction subscription: %T\n", content)
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
		fmt.Println("Mining started")
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
                                fmt.Printf("API server error: %v\n", err)
                        }
                }()
                
                fmt.Printf("API server started on port %d\n", apiPort)
        }

        // Start periodic status updates
        go func() {
                ticker := time.NewTicker(30 * time.Second)
                defer ticker.Stop()

                for {
                        select {
                        case <-ctx.Done():
                                return
                        case <-ticker.C:
                                bestBlock := chain.GetBestBlock()
                                if bestBlock != nil {
                                        fmt.Printf("Status: Height=%d, Hash=%x, Peers=%d, Mempool=%d\n",
                                                chain.GetHeight(),
                                                bestBlock.CalculateHash(),
                                                len(net.GetPeers()),
                                                mempool.GetTransactionCount())
                                } else {
                                        fmt.Printf("Status: Height=%d, Peers=%d, Mempool=%d\n",
                                                chain.GetHeight(),
                                                len(net.GetPeers()),
                                                mempool.GetTransactionCount())
                                }
                        }
                }
        }()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down GoChain node...")

	// Cancel context to stop all goroutines
	cancel()

	        // Cleanup
        if mining {
                miner.StopMining()
        }
        miner.Close()
        net.Close()
        
        // Close API server if it was started
        if apiServer != nil {
                // Note: The API server doesn't have a Close method yet, but we could add one if needed
                fmt.Println("API server stopped")
        }

	fmt.Println("GoChain node stopped")
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

			bestBlock := chain.GetBestBlock()
			fmt.Printf("Blockchain Information:\n")
			fmt.Printf("Height: %d\n", chain.GetHeight())
			fmt.Printf("Best Block Hash: %x\n", bestBlock.CalculateHash())
			fmt.Printf("Genesis Block Hash: %x\n", chain.GetGenesisBlock().CalculateHash())
			fmt.Printf("Difficulty: %d\n", bestBlock.Header.Difficulty)
			fmt.Printf("Next Difficulty: %d\n", chain.CalculateNextDifficulty())

			return nil
		},
	}
}
