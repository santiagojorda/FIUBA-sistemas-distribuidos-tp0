package common

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/op/go-logging"
)

const TIME_RETRY_ASK_WINNERS = 2000 * time.Millisecond

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

	protocol, err := NewProtocol(c.config, c.connection)
	if err != nil {
		c.log.Errorf("action: config | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return err
	}
	c.protocol = protocol

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

	batchBuilder, err := NewBatchBuilder(c.config.ID, c.config.MaxBatchAmount, c.config.MaxBatchSize)
	if err != nil {
		c.log.Errorf("action: send_batch | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}
	defer batchBuilder.Close()

	if err := c.runBatchLoop(batchBuilder); err != nil {
		c.log.Errorf("action: send_batch | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}

	if err := c.protocol.SendFINMessage(); err != nil {
		return
	}

	c.log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)

	if err := c.askWinnersLoop(); err != nil {
		c.log.Errorf("action: consulta_ganadores | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}
}

func (c *Client) askWinnersLoop() error {
	for {
		if err := c.Start(); err != nil {
			return err
		}

		if err := c.protocol.SendAgencyIDMessage(); err != nil {
			c.closeResources()
			return err
		}

		count, ready, err := c.protocol.AskWinners()
		c.closeResources()
		if err != nil {
			return err
		}

		if ready {
			c.log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %v", count)
			return nil
		}

		time.Sleep(TIME_RETRY_ASK_WINNERS)
	}
}

func (c *Client) runBatchLoop(batchBuilder *BatchBuilder) error {
	for {
		bets, reachedEOF, err := batchBuilder.NextBatch()
		if err != nil {
			return err
		}

		if len(bets) > 0 {
			if err := c.protocol.SendBatch(bets); err != nil {
				return err
			}
		}

		if reachedEOF {
			return nil
		}
	}
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
