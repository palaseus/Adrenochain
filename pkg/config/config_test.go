package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultSystemConfig(t *testing.T) {
	cfg := DefaultSystemConfig()
	if cfg.Consensus.TargetBlockTime != 12*time.Second {
		t.Fatalf("target block time = %s", cfg.Consensus.TargetBlockTime)
	}
	if cfg.Consensus.DifficultyAdjustmentInterval != 2016 || cfg.Consensus.MaxDifficulty != 1<<30 {
		t.Fatalf("difficulty bounds = %+v", cfg.Consensus)
	}
	if cfg.Consensus.MinDifficulty != 1 || cfg.Consensus.DifficultyAdjustmentFactor != 1.1 {
		t.Fatalf("difficulty factor = %+v", cfg.Consensus)
	}
	if cfg.Consensus.FinalityDepth != 10 || cfg.Consensus.CheckpointInterval != 262974 {
		t.Fatalf("finality = %+v", cfg.Consensus)
	}
	if cfg.Network.MaxPeers != 50 || cfg.Network.ConnectionTimeout != 30*time.Second {
		t.Fatalf("network = %+v", cfg.Network)
	}
	if cfg.Blockchain.GenesisBlockReward != 1_000_000_000 || cfg.Blockchain.MaxBlockSize != 1_000_000 {
		t.Fatalf("chain = %+v", cfg.Blockchain)
	}
	if cfg.Storage.FilePermissions != 0o700 {
		t.Fatalf("permissions = %#o", cfg.Storage.FilePermissions)
	}
}

func TestLoadSystemConfigFillsZeros(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("network:\n  max_peers: 7\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadSystemConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Network.MaxPeers != 7 {
		t.Fatalf("max peers = %d", cfg.Network.MaxPeers)
	}
	if cfg.Consensus.FinalityDepth != 10 || cfg.Storage.FilePermissions != 0o700 {
		t.Fatalf("defaults were not preserved: %+v", cfg)
	}
}

func TestLoadSystemConfigRestoresExplicitZeroMaxDifficulty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	body := []byte("consensus:\n  max_difficulty: 0\n  min_difficulty: 4\n")
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadSystemConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Consensus.MaxDifficulty != 1<<30 {
		t.Fatalf("max difficulty = %d", cfg.Consensus.MaxDifficulty)
	}
	if cfg.Consensus.MinDifficulty != 4 {
		t.Fatalf("min difficulty = %d", cfg.Consensus.MinDifficulty)
	}
}
