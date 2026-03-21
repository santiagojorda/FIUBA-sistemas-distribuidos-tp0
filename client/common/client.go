package common

import (
	"fmt"
	"net"
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
}

// Client Entity that encapsulates how
type Client struct {
	config          ClientConfig
	conn            net.Conn
	shutdown_event  chan os.Signal
	players         []Player
}

// NewClient Initializes a clientnew client receiving the configuration
// as a parameter
func NewClient(config ClientConfig, player Player) *Client {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM)
	client := &Client{
		config:         config,
		shutdown_event: signalChan,
	}
	client.players = append(client.players, player)
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	maxRetries := 30
	retryDelay := time.Second
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		conn, err := net.Dial("tcp", c.config.ServerAddress)
		if err == nil {
			c.conn = conn
			return nil
		}
		
		if attempt < maxRetries {
			log.Infof(
				"action: connect | result: in_progress | client_id: %v | attempt: %v/%v | error: %v",
				c.config.ID,
				attempt,
				maxRetries,
				err,
			)
			time.Sleep(retryDelay)
		} else {
			log.Criticalf(
				"action: connect | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return err
		}
	}
	return fmt.Errorf("failed to connect after %d attempts", maxRetries)
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {

	// creo el socket del cliente
	if err := c.createClientSocket(); err != nil {
		log.Errorf("action: create_client_socket | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}

	// Envio el mensaje de agencia al servidor
	fmt.Fprintf(
		c.conn,
		"%s\n",
		c.config.ID,
	)

	// envio cada jugador al servidor
	for _, player := range c.players {
		// verifico que no se haya recibido una señal de shutdown
		select {
		case <-c.shutdown_event:
			log.Infof("action: graceful_shutdown | result: in_progress | client_id: %v", c.config.ID)
			if c.conn != nil {
				c.conn.Close()
			}
			log.Infof("action: graceful_shutdown | result: success | client_id: %v", c.config.ID)
			return
		default:
		}

		// envio el mensaje de cada jugador al servidor
		jsonData, err := serializePlayer(player)
		if err != nil {
			log.Errorf("action: send_message | result: fail | client_id: %v | error: %v", c.config.ID, err)
			return
		}

		if err := writeAll(c.conn, jsonData); err != nil {
			log.Errorf("action: send_message | result: fail | client_id: %v | error: %v", c.config.ID, err)
			return
		}

		log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v", player.Dni, player.Number)


			// Escribir un salto de línea después del mensaje JSON
		if err := writeAll(c.conn, []byte("\n")); err != nil {
			log.Errorf("action: send_message | result: fail | client_id: %v | error: %v", c.config.ID, err)
			return
		}
		log.Infof("action: send_message | result: success | client_id: %v | player: %v", c.config.ID, player.Name)

	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}
