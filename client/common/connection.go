package common

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/op/go-logging"
)

const SOCKET_CONNECTION_MAX_RETRIES = 5

// Connection encapsulates socket and transport operations
type Connection struct {
	address    string
	conn       net.Conn
	reader     *bufio.Reader
	log				 *logging.Logger
	clientID   string
	maxRetries int
}

// NewConnection creates a new Connection instance
func NewConnection(clientID, serverAddress string, log *logging.Logger) *Connection {
	return &Connection{
		address:    serverAddress,
		clientID:   clientID,
		maxRetries: SOCKET_CONNECTION_MAX_RETRIES,
		log:				log,
	}
}

// Connect establishes connection to server with retry logic
func (c *Connection) Connect() error {
	retryDelay := time.Second

	for attempt := 1; attempt <= c.maxRetries; attempt++ {
		conn, err := net.Dial("tcp", c.address)
		if err == nil {
			c.conn = conn
			c.reader = bufio.NewReader(conn)
			c.log.Infof(
				"action: connect | result: success | client_id: %v | attempt: %v/%v",
				c.clientID,
				attempt,
				c.maxRetries,
			)
			return nil
		}

		if attempt < c.maxRetries {
			c.log.Infof(
				"action: connect | result: in_progress | client_id: %v | attempt: %v/%v | error: %v",
				c.clientID,
				attempt,
				c.maxRetries,
				err,
			)
			time.Sleep(retryDelay)
		} else {
			c.log.Criticalf(
				"action: connect | result: fail | client_id: %v | error: %v",
				c.clientID,
				err,
			)
			return err
		}
	}
	return fmt.Errorf("failed to connect after %d attempts", c.maxRetries)
}

// Send sends data to server, handling short writes
func (c *Connection) Send(data []byte) error {
	if c.conn == nil {
		return fmt.Errorf("connection not established")
	}

	totalWritten := 0
	for totalWritten < len(data) {
		n, err := c.conn.Write(data[totalWritten:])
		if err != nil {
			c.log.Errorf("action: send_batch | result: fail | error: %v", err)
			return err
		}
		totalWritten += n
	}
	return nil
}

// ReceiveLine reads a line from server
func (c *Connection) ReceiveLine() (string, error) {
	if c.reader == nil {
		return "", fmt.Errorf("reader not initialized")
	}

	line, err := c.reader.ReadString('\n')
	if err != nil {
		c.log.Errorf("action: receive_message | result: fail | error: %v", err)
		return "", err
	}

	return trimLine(line), nil
}

// SendAgency sends agency ID to server
func (c *Connection) SendAgency(agencyID string) error {
	message := fmt.Sprintf("%s\n", agencyID)
	return c.Send([]byte(message))
}

// ReceiveConfirmation waits for server confirmation
func (c *Connection) ReceiveConfirmation() (string, error) {
	return c.ReceiveLine()
}

// SendFinished sends the finalization message to server.
func (c *Connection) SendFinished() error {
	return c.Send([]byte("FIN\n"))
}

// ReceiveFinishedConfirmation waits for server confirmation once all clients finished.
func (c *Connection) ReceiveFinishedConfirmation() (string, error) {
	return c.ReceiveLine()
}

// Close closes the connection gracefully
func (c *Connection) Close() error {
	if c.conn != nil {
		c.log.Infof("action: graceful_shutdown | result: in_progress | client_id: %v | connection_closed: true", c.clientID)
		return c.conn.Close()
	}
	return nil
}

// IsConnected checks if connection is active
func (c *Connection) IsConnected() bool {
	return c.conn != nil
}

// trimLine removes whitespace from line
func trimLine(line string) string {
	for i := len(line) - 1; i >= 0; i-- {
		if line[i] != '\n' && line[i] != '\r' && line[i] != ' ' && line[i] != '\t' {
			return line[:i+1]
		}
	}
	return ""
}
