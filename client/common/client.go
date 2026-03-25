package common

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	config          ClientConfig
	connection      *Connection
	shutdown_event  chan os.Signal
	Log						  *logging.Logger
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM)
	client := &Client{
		config:         config,
		shutdown_event: signalChan,
		Log:            log,
	}
	return client
}

func (c *Client) Start() error {
	c.connection = NewConnection(c.config.ID, c.config.ServerAddress, c.Log)

	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) Run() {
		if err := c.Start(); err != nil {
		return
	}
	defer c.closeResources()

	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {

		if c.isShutdownRequested() {
			return
		}

		if err := c.connection.Connect(); err != nil {
			c.Log.Errorf("action: connect | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		// TODO: Modify the send to avoid short-write
		if err := c.connection.SendMessage(fmt.Sprintf(
			"[CLIENT %v] Message N°%v\n",
			c.config.ID,
			msgID,
		)); err != nil {
			c.Log.Errorf("action: send_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			c.connection.Close()
			return
		}

		msg, err := c.connection.ReadLine()
		c.connection.Close()

		if err != nil {
			c.Log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		c.Log.Infof("action: receive_message | result: success | client_id: %v | msg: %v",
			c.config.ID,
			msg,
		)

		// Wait a time between sending one message and the next one
		time.Sleep(c.config.LoopPeriod)

	}
	c.Log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}

func (c *Client) isShutdownRequested() bool {
	select {
	case <-c.shutdown_event:
		c.Log.Infof("action: graceful_shutdown | result: in_progress | client_id: %v", c.config.ID)
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
