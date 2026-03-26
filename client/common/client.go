package common

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/op/go-logging"
)

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID             string
	ServerAddress  string
	MaxBatchAmount int
	MaxBatchSize   int
	LoopPeriod     time.Duration
	Log            *logging.Logger
}

// Client entity that coordinates connection and protocol flow
type Client struct {
	config         ClientConfig
	connection     *Connection
	shutdown_event chan os.Signal
	protocol       *Protocol
	log            *logging.Logger
}

// NewClient initializes a new client receiving the configuration as a parameter
func NewClient(config ClientConfig) *Client {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM)

	return &Client{
		config:         config,
		shutdown_event: signalChan,
		log:            config.Log,
	}
}

func (c *Client) Start() error {
	c.connection = NewConnection(c.config.ID, c.config.ServerAddress, c.log)
	c.protocol = NewProtocol(c.connection, c.log, c.config.ID, c.config.MaxBatchAmount, c.config.MaxBatchSize)

	if err := c.connection.Connect(); err != nil {
		c.log.Errorf("action: connect | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return err
	}

	return nil
}

// Run executes the client protocol lifecycle
func (c *Client) Run() {
	if err := c.Start(); err != nil {
		return
	}
	defer c.closeResources()

	if c.isShutdownRequested() {
		return
	}

	if err := c.protocol.SendAgencyIDMessage(); err != nil {
		return
	}

	if err := c.protocol.SendBatchesFromCSV(); err != nil {
		c.log.Errorf("action: send_batch | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}

	if err := c.protocol.SendFINMessage(); err != nil {
		return
	}

	c.log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}

func (c *Client) isShutdownRequested() bool {
	select {
	case <-c.shutdown_event:
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
}
