// Package config holds process-wide defaults shared by the node packages.
// Values match the assertions in the consensus and network tests.
package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// SystemConfig is the top-level configuration shared by node subsystems.
type SystemConfig struct {
	Network    NetworkConfig    `yaml:"network"`
	Blockchain BlockchainConfig `yaml:"blockchain"`
	Consensus  ConsensusConfig  `yaml:"consensus"`
	Storage    StorageConfig    `yaml:"storage"`
}

// NetworkConfig controls peer limits for the libp2p host.
type NetworkConfig struct {
	ListenPort        int           `yaml:"listen_port"`
	BootstrapPeers    []string      `yaml:"bootstrap_peers"`
	EnableMDNS        bool          `yaml:"enable_mdns"`
	EnableRelay       bool          `yaml:"enable_relay"`
	MaxPeers          int           `yaml:"max_peers"`
	ConnectionTimeout time.Duration `yaml:"connection_timeout"`
}

// BlockchainConfig controls chain-level limits.
type BlockchainConfig struct {
	GenesisBlockReward uint64        `yaml:"genesis_block_reward"`
	BlockTime          time.Duration `yaml:"block_time"`
	MaxBlockSize       uint64        `yaml:"max_block_size"`
	MaxReorgDepth      uint64        `yaml:"max_reorg_depth"`
}

// ConsensusConfig controls proof-of-work difficulty and finality.
type ConsensusConfig struct {
	TargetBlockTime              time.Duration `yaml:"target_block_time"`
	DifficultyAdjustmentInterval uint64        `yaml:"difficulty_adjustment_interval"`
	MaxDifficulty                uint64        `yaml:"max_difficulty"`
	MinDifficulty                uint64        `yaml:"min_difficulty"`
	DifficultyAdjustmentFactor   float64       `yaml:"difficulty_adjustment_factor"`
	FinalityDepth                uint64        `yaml:"finality_depth"`
	CheckpointInterval           uint64        `yaml:"checkpoint_interval"`
}

// StorageConfig controls on-disk layout and directory permissions.
type StorageConfig struct {
	DataDir         string `yaml:"data_dir"`
	DBType          string `yaml:"db_type"`
	FilePermissions uint32 `yaml:"file_permissions"`
}

// DefaultSystemConfig returns the built-in defaults asserted by package tests.
func DefaultSystemConfig() *SystemConfig {
	return &SystemConfig{
		Network: NetworkConfig{
			ListenPort:        0,
			BootstrapPeers:    []string{},
			EnableMDNS:        true,
			EnableRelay:       false,
			MaxPeers:          50,
			ConnectionTimeout: 30 * time.Second,
		},
		Blockchain: BlockchainConfig{
			GenesisBlockReward: 1_000_000_000,
			BlockTime:          10 * time.Second,
			MaxBlockSize:       1_000_000,
			MaxReorgDepth:      100,
		},
		Consensus: ConsensusConfig{
			TargetBlockTime:              12 * time.Second,
			DifficultyAdjustmentInterval: 2016,
			MaxDifficulty:                1 << 30,
			MinDifficulty:                1,
			DifficultyAdjustmentFactor:   1.1,
			FinalityDepth:                10,
			CheckpointInterval:           262974,
		},
		Storage: StorageConfig{
			DataDir:         "./data",
			DBType:          "file",
			FilePermissions: 0o700,
		},
	}
}

// LoadSystemConfig reads a YAML file and fills zero fields from the defaults
// so a partial file cannot disable directory permissions or consensus bounds.
func LoadSystemConfig(path string) (*SystemConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	cfg := DefaultSystemConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	if cfg.Storage.FilePermissions == 0 {
		cfg.Storage.FilePermissions = 0o700
	}
	if cfg.Consensus.MinDifficulty == 0 {
		cfg.Consensus.MinDifficulty = 1
	}
	if cfg.Consensus.MaxDifficulty == 0 {
		cfg.Consensus.MaxDifficulty = 1 << 30
	}
	if cfg.Consensus.TargetBlockTime == 0 {
		cfg.Consensus.TargetBlockTime = 12 * time.Second
	}
	return cfg, nil
}
