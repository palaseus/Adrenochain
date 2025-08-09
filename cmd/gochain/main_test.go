package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"syscall"
	"testing"
	"time"

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
