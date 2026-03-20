package common

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
	"encoding/json"

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
	conn            net.Conn
	shutdown_event  chan os.Signal
	players         []Player
}

type ProtocolMessage struct {
	Name string `json:"name"`
	Lastname string `json:"lastname"`
	Dni string `json:"dni"`
	Birthdate string `json:"birthdate"`
	Number string `json:"number"`
}

type Player struct{
	Name string
  Lastname string
  Dni string
  Birthdate string
  Number string	
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
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}
	c.conn = conn
	return nil
}

func serializePlayer(player Player) ([]byte, error) {
	
	protocolMsg := ProtocolMessage{
		Name: player.Name,
		Lastname: player.Lastname,
		Dni: player.Dni,
		Birthdate: player.Birthdate,
		Number: player.Number,
	}
	
	json_data, err := json.Marshal(protocolMsg)
	if err != nil {
		return nil, err
	}
	log.Infof("action: serialize_player | result: success | client_id: %v | player: %v", player.Name, player.Lastname)

	return json_data, nil
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
		"[AGENCY] %s\n",
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
		json_data, err := serializePlayer(player)
		totalWritten := 0
		for totalWritten < len(json_data) {
			n, err := c.conn.Write(json_data[totalWritten:])
			if err != nil {
				log.Errorf("action: send_message | result: fail | client_id: %v | error: %v", c.config.ID, err)
				return
			}
			totalWritten += n
		}
			// Escribir un salto de línea después del mensaje JSON
		_, err = c.conn.Write([]byte("\n"))
		if err != nil {
			log.Errorf("action: send_message | result: fail | client_id: %v | error: %v", c.config.ID, err)
			return
		}
		log.Infof("action: send_message | result: success | client_id: %v | player: %v", c.config.ID, player.Name)

		time.Sleep(c.config.LoopPeriod)
	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}
