package common

import (
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
	LoopPeriod    time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	config          ClientConfig
	connection      *Connection
	shutdown_event  chan os.Signal
	Log						  *logging.Logger
  bets						[]Bet
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig, bets []Bet) *Client {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM)
	client := &Client{
		config:         config,
		shutdown_event: signalChan,
		Log:            log,
		bets:           bets,
	}
	return client
}

func (c *Client) Start() error {
	c.connection = NewConnection(c.config.ID, c.config.ServerAddress, c.Log)

	if err := c.connection.Connect(); err != nil {
		c.Log.Errorf("action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}

	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) Run() {
	if err := c.Start(); err != nil {
		return
	}
	defer c.closeResources()

	if c.isShutdownRequested() {
		return
	}

	if err := c.connection.SendAgency(c.config.ID); err != nil {
		c.Log.Errorf("action: send_message | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}

	c.Log.Infof("action: send_agency | result: success | client_id: %v | agency_id: %v",
		c.config.ID,
		c.config.ID,
	)

	for _, bet := range c.bets {
		if err := c.connection.Send(SerializeBets([]Bet{bet})); err != nil {
			c.Log.Errorf("action: send_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		c.Log.Infof("action: apuesta_enviada | result: success | client_id: %v | dni: %v | numero: %v",
			c.config.ID,
			bet.Dni,
			bet.Number,
		)

		msg, err := c.connection.ReadLine()
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

		if c.config.LoopPeriod > 0 {
			time.Sleep(c.config.LoopPeriod)
		}
	}

	if err := c.connection.SendMessage(MESSAGE_FIN + "\n"); err != nil {
		c.Log.Errorf("action: send_message | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
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