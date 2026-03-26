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
			c.log.Infof("action: connect | result: success | client_id: %v | attempt: %v/%v",
				c.clientID,
				attempt,
				c.maxRetries,
			)
			return nil
		}

		if attempt < c.maxRetries {
			c.log.Infof("action: connect | result: in_progress | client_id: %v | attempt: %v/%v | error: %v",
				c.clientID,
				attempt,
				c.maxRetries,
				err,
			)
			time.Sleep(retryDelay)
		} else {
			c.log.Criticalf("action: connect | result: fail | client_id: %v | error: %v",
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

// Receive reads a newline-terminated message from the server.
func (c *Connection) Receive() (string, error) {
	if c.reader == nil {
		return "", fmt.Errorf("connection reader is not initialized")
	}

	return c.reader.ReadString('\n')
}

// Close closes the connection gracefully
func (c *Connection) Close() error {
	if c.conn != nil {
		c.log.Infof("action: graceful_shutdown | result: in_progress | client_id: %v | connection_closed: true", c.clientID)
		err := c.conn.Close()
		c.conn = nil
		c.reader = nil
		return err
	}
	return nil
}

// IsConnected checks if connection is active
func (c *Connection) IsConnected() bool {
	return c.conn != nil
}