package common

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/op/go-logging"
)


// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID              string
	ServerAddress   string
	MaxBatchAmount  int
	MaxBatchSize    int
	Log 					  *logging.Logger
}

// Client orchestrates the communication with the server
type Client struct {
	config          ClientConfig
	connection      *Connection
	batchBuilder    *BatchBuilder
	protocolHandler *ProtocolHandler
	shutdownEvent   chan os.Signal
	log						 *logging.Logger
}

// NewClient initializes the client with dependencies
func NewClient(config ClientConfig) *Client {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM)

	return &Client{
		config:        config,
		shutdownEvent: signalChan,
		log:           config.Log,
	}
}

// Start initializes connection and starts the client loop
func (c *Client) Start() error {
	// Initialize connection component
	c.connection = NewConnection(
		c.config.ID,
		c.config.ServerAddress,
		c.log,
	)

	// Connect to server
	if err := c.connection.Connect(); err != nil {
		c.log.Errorf("action: create_client_socket | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}

	// Initialize batch builder
	csvPath := fmt.Sprintf(".data/agency-%s.csv", c.config.ID)
	batchBuilder, err := NewBatchBuilder(
		csvPath,
		c.config.MaxBatchSize,
		c.config.MaxBatchAmount,
		c.config.ID,
		c.log,
	)
	if err != nil {
		return err
	}
	c.batchBuilder = batchBuilder

	// Initialize protocol handler
	c.protocolHandler = NewProtocolHandler(c.connection, c.log)

	return nil
}

// Run executes the main client loop
func (c *Client) Run() {
	if err := c.Start(); err != nil {
		return
	}

	defer c.closeResources()

	// Send agency ID
	if err := c.connection.SendAgency(c.config.ID); err != nil {
		c.log.Errorf("action: send_agency | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}

	// Process batches
	c.processBatches()

	c.log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}

// processBatches reads batches from CSV and sends them to server
func (c *Client) processBatches() {
	for {
		// Check for shutdown signal
		if c.isShutdownRequested() {
			return
		}

		// Read next batch
		bets, hasMore, err := c.batchBuilder.NextBatch()
		if err != nil {
			c.log.Errorf("action: read_bets_from_csv | result: fail | error: %v", err)
			return
		}

		// If no bets, we're done
		if len(bets) == 0 {
			break
		}

		// Send batch to server
		if err := c.protocolHandler.SendBatch(bets); err != nil {
			c.log.Errorf("action: send_batch | result: fail | error: %v", err)
			return
		}

		// If no more batches, exit loop
		if !hasMore {
			break
		}
	}
}

// isShutdownRequested checks if shutdown signal was received
func (c *Client) isShutdownRequested() bool {
	select {
	case <-c.shutdownEvent:
		c.log.Infof("action: graceful_shutdown | result: in_progress | client_id: %v", c.config.ID)
		return true
	default:
		return false
	}
}

// closeResources closes all resources gracefully
func (c *Client) closeResources() {
	if c.connection != nil && c.connection.IsConnected() {
		c.connection.Close()
	}
	if c.batchBuilder != nil {
		c.batchBuilder.Close()
	}
}

