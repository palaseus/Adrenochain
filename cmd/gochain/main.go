//go:build go1.20

package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gochain/gochain/pkg/block"
    "github.com/gochain/gochain/pkg/chain"
    "github.com/gochain/gochain/pkg/mempool"
    "github.com/gochain/gochain/pkg/miner"
    netpkg "github.com/gochain/gochain/pkg/net"
    "github.com/gochain/gochain/pkg/wallet"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var (
	configFile string
	port       int
	mining     bool
	network    string
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
	chainConfig := chain.DefaultChainConfig()
	chain := chain.NewChain(chainConfig)

	mempoolConfig := mempool.DefaultMempoolConfig()
	mempool := mempool.NewMempool(mempoolConfig)

	minerConfig := miner.DefaultMinerConfig()
	minerConfig.MiningEnabled = mining
	minerConfig.CoinbaseAddress = "miner_reward"
	miner := miner.NewMiner(chain, mempool, minerConfig)

    networkConfig := netpkg.DefaultNetworkConfig()
	networkConfig.ListenPort = port
	networkConfig.EnableMDNS = true
	networkConfig.MaxPeers = 50

    net, err := netpkg.NewNetwork(networkConfig)
	if err != nil {
		return fmt.Errorf("failed to create network: %w", err)
	}

	walletConfig := wallet.DefaultWalletConfig()
	wallet, err := wallet.NewWallet(walletConfig)
	if err != nil {
		return fmt.Errorf("failed to create wallet: %w", err)
	}

	// Set up network message handlers
	if err := net.SubscribeToBlocks(func(block *block.Block) {
		fmt.Printf("Received block from network: %s\n", block.String())
		if err := chain.AddBlock(block); err != nil {
			fmt.Printf("Failed to add received block: %v\n", err)
		}
	}); err != nil {
		return fmt.Errorf("failed to subscribe to blocks: %w", err)
	}

	if err := net.SubscribeToTransactions(func(tx *block.Transaction) {
		fmt.Printf("Received transaction from network: %s\n", tx.String())
		if err := mempool.AddTransaction(tx); err != nil {
			fmt.Printf("Failed to add received transaction: %v\n", err)
		}
	}); err != nil {
		return fmt.Errorf("failed to subscribe to transactions: %w", err)
	}

	// Start mining if enabled
	if mining {
		if err := miner.StartMining(); err != nil {
			return fmt.Errorf("failed to start mining: %w", err)
		}
		fmt.Println("Mining started")
	}

	// Start periodic status updates
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				bestBlock := chain.GetBestBlock()
				fmt.Printf("Status: Height=%d, Hash=%x, Peers=%d, Mempool=%d\n",
					chain.GetHeight(),
					bestBlock.CalculateHash(),
					net.GetPeerCount(),
					mempool.GetTransactionCount())
			}
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down GoChain node...")

	// Cleanup
	if mining {
		miner.StopMining()
	}
	miner.Close()
	net.Close()

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
		Short: "Create a new wallet",
		RunE: func(cmd *cobra.Command, args []string) error {
			walletConfig := wallet.DefaultWalletConfig()
			wallet, err := wallet.NewWallet(walletConfig)
			if err != nil {
				return fmt.Errorf("failed to create wallet: %w", err)
			}

			account := wallet.GetDefaultAccount()
			fmt.Printf("Wallet created successfully!\n")
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
			walletConfig := wallet.DefaultWalletConfig()
			wallet, err := wallet.NewWallet(walletConfig)
			if err != nil {
				return fmt.Errorf("failed to create wallet: %w", err)
			}

			tx, err := wallet.CreateTransaction(from, to, amount, fee)
			if err != nil {
				return fmt.Errorf("failed to create transaction: %w", err)
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
			walletConfig := wallet.DefaultWalletConfig()
			wallet, err := wallet.NewWallet(walletConfig)
			if err != nil {
				return fmt.Errorf("failed to create wallet: %w", err)
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
			chainConfig := chain.DefaultChainConfig()
			chain := chain.NewChain(chainConfig)

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