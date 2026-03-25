package common

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/op/go-logging"
)

// ProtocolHandler manages serialization and sending of bets batches
type ProtocolHandler struct {
	connection *Connection
	log			   *logging.Logger
}

// NewProtocolHandler creates a new ProtocolHandler instance
func NewProtocolHandler(connection *Connection, log *logging.Logger) *ProtocolHandler {
	return &ProtocolHandler{
		connection: connection,
		log: log,
	}
}

// SendBatch serializes and sends a batch of bets to the server
// Format: <count>\n<bet1_data>\n<bet2_data>\n...
func (p *ProtocolHandler) SendBatch(bets []Bet) error {
	if len(bets) == 0 {
		return fmt.Errorf("cannot send empty batch")
	}

	payload := p.serializeBatch(bets)
	
	err := p.connection.Send(payload)
	if err != nil {
		return err
	}

	// Wait for confirmation
	confirmation, err := p.connection.ReceiveConfirmation()
	if err != nil || confirmation == "" {
		p.log.Errorf("action: receive_confirmation | result: fail | error: %v", err)
		return fmt.Errorf("failed to receive confirmation: %v", err)
	}

	p.log.Infof("action: batch_enviado | result: success | cantidad: %v", len(bets))
	return nil
}

// serializeBatch formats the batch data according to protocol
// Format: <count>\n + <name|lastname|dni|birthdate|number>\n for each bet
func (p *ProtocolHandler) serializeBatch(bets []Bet) []byte {
	var buffer bytes.Buffer

	// Write batch count
	buffer.WriteString(strconv.Itoa(len(bets)))
	buffer.WriteString("\n")

	// Write each bet
	for _, bet := range bets {
		p.formatBet(&buffer, bet)
		buffer.WriteString("\n")
	}

	return buffer.Bytes()
}

// formatBet writes a single bet in pipe-delimited format
func (p *ProtocolHandler) formatBet(buffer *bytes.Buffer, bet Bet) {
	buffer.WriteString(bet.Name)
	buffer.WriteString("|")
	buffer.WriteString(bet.Lastname)
	buffer.WriteString("|")
	buffer.WriteString(bet.Dni)
	buffer.WriteString("|")
	buffer.WriteString(bet.Birthdate)
	buffer.WriteString("|")
	buffer.WriteString(bet.Number)
}
