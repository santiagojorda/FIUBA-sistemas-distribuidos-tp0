package common

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	"encoding/csv"
	"strings"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")
var builder strings.Builder

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
}

// Client Entity that encapsulates how
type Client struct {
	config          ClientConfig
	conn            net.Conn
	reader          *bufio.Reader
	shutdown_event  chan os.Signal	
}

// NewClient Initializes a clientnew client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM)
	client := &Client{
		config:         config,
		shutdown_event: signalChan,
	}
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
			c.reader = bufio.NewReader(conn)
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
	// Crear el socket del cliente
	if err := c.createClientSocket(); err != nil {
		log.Errorf("action: create_client_socket | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}

	// Enviar el mensaje de agencia al servidor
	fmt.Fprintf(c.conn, "%s\n", c.config.ID)

	file, err := os.Open(fmt.Sprintf(".data/agency-%s.csv", c.config.ID))
	if err != nil {
		log.Errorf("action: open_file | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	temp_bet := Bet{}

	for {
		// Verificar señal de shutdown ANTES de procesar el batch
		select {
		case <-c.shutdown_event:
			log.Infof("action: graceful_shutdown | result: in_progress | client_id: %v", c.config.ID)
			if c.conn != nil {
				c.conn.Close()
			}
			return
		default:
		}

		bets := []Bet{}
		packageSize := 0
		if temp_bet.Name != "" {
			bets = append(bets, temp_bet)
			packageSize += len(temp_bet.Name) + len(temp_bet.Lastname) + len(temp_bet.Dni) + len(temp_bet.Birthdate) + len(temp_bet.Number)
			temp_bet = Bet{}
			log.Infof("action: read_bets_from_csv | result: in_progress | bet: %v | package_size: %v", bets[0].Name+" "+bets[0].Lastname, packageSize)
		}

		// Leer líneas del CSV hasta llenar el buffer de 8KB o llegar al EOF
		for {
			record, err := reader.Read()
			if err != nil {
				if err.Error() == "EOF" {
					log.Infof("action: read_bets_from_csv | result: EOF_reached | bets_count: %v", len(bets))
					break
				}
				log.Errorf("action: read_bets_from_csv | result: fail | error: %v", err)
				return
			}

			if len(record) < 5 {
				log.Errorf("CSV file does not contain enough columns")
				continue
			}

			bet := Bet{
				Name:      record[0],
				Lastname:  record[1],
				Dni:       record[2],
				Birthdate: record[3],
				Number:    record[4],
			}

			chunkBSize := len(bet.Name) + len(bet.Lastname) + len(bet.Dni) + len(bet.Birthdate) + len(bet.Number)
			bufferSize := 8 * 1024

			if chunkBSize+packageSize >= bufferSize {
				temp_bet = bet
				log.Infof("action: read_bets_from_csv | result: package_full | package_size: %v", packageSize)
				break
			} else {
				packageSize += chunkBSize
				bets = append(bets, bet)
			}
		}

		// Si no hay bets y llegamos al EOF, terminar
		if len(bets) == 0 {
			log.Infof("action: read_bets_from_csv | result: success | no_more_bets")
			break
		}

		// Construir el batch: cantidad + datos delimitados por |
		var batchData bytes.Buffer
		batchData.WriteString(strconv.Itoa(len(bets)))
		batchData.WriteString("\n")

		for _, bet := range bets {
			batchData.WriteString(bet.Name)
			batchData.WriteString("|")
			batchData.WriteString(bet.Lastname)
			batchData.WriteString("|")
			batchData.WriteString(bet.Dni)
			batchData.WriteString("|")
			batchData.WriteString(bet.Birthdate)
			batchData.WriteString("|")
			batchData.WriteString(bet.Number)
			batchData.WriteString("\n")
		}

		// Enviar el batch
		payload := batchData.Bytes()
		escritos := 0

		for escritos < len(payload) {
				n, err := c.conn.Write(payload[escritos:])
				escritos += n 
				
				if err != nil {
						log.Errorf("action: send_batch | result: fail | error: %v", err)
						return
				}
		}

		// Esperar confirmación del servidor
		confirmation, err := readLine(c.reader)
		if err != nil || confirmation == "" {
			log.Errorf("action: receive_confirmation | result: fail | error: %v", err)
			return
		}

		log.Infof("action: batch_enviado | result: success | cantidad: %v", len(bets))
	}

	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}
