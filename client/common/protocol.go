package common

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/op/go-logging"
)

const AGENCY_CSV_PATH_TEMPLATE = "/.data/agency-%s.csv"
const MESSAGE_FIN = "FIN"
const MESSAGE_END_BATCH = "END_BATCH"
const MESSAGE_OK = "ok"
const MESSAGE_ERROR = "error"

// Protocol manages serialization and sending of bets batches
type Protocol struct {
	connection *Connection
	log        *logging.Logger
	clientID   string
}

// NewProtocol creates a new Protocol instance
func NewProtocol(config ClientConfig, connection *Connection) (*Protocol, error) {
	if connection == nil {
		return nil, fmt.Errorf("connection cannot be nil")
	}

	return &Protocol{
		connection: connection,
		log:        config.Log,
		clientID:   config.ID,
	}, nil
}

func (p *Protocol) SendAgencyIDMessage() error {
	if err := p.connection.Send([]byte(fmt.Sprintf("%s\n", p.clientID))); err != nil {
		p.log.Errorf("action: send_message | result: fail | client_id: %v | error: %v", p.clientID, err)
		return err
	}

	p.log.Infof("action: send_agency | result: success | client_id: %v | agency_id: %v", p.clientID, p.clientID)
	return nil
}

func (p *Protocol) SendFINMessage() error {
	if err := p.connection.Send([]byte(MESSAGE_FIN + "\n")); err != nil {
		p.log.Errorf("action: send_message | result: fail | client_id: %v | error: %v", p.clientID, err)
		return err
	}

	p.log.Infof("action: send_fin_message | result: success | client_id: %v", p.clientID)
	return nil
}

// SendBatch sends one bets batch and waits for server confirmation.
func (p *Protocol) SendBatch(bets []Bet) error {
	payload := p.serializeBatchPayload(bets)
	if err := p.connection.Send(payload); err != nil {
		return err
	}

	confirmation, err := p.connection.Receive()
	if err != nil || confirmation == "" {
		p.log.Errorf("action: receive_confirmation | result: fail | error: %v", err)
		return fmt.Errorf("failed to receive confirmation: %v", err)
	}

	confirmation = strings.ToLower(strings.TrimSpace(confirmation))
	switch confirmation {
	case MESSAGE_OK:
		// Expected successful ACK.
	case MESSAGE_ERROR:
		p.log.Errorf("action: receive_confirmation | result: fail | client_id: %v | error: server rejected batch", p.clientID)
		return fmt.Errorf("server rejected batch")
	default:
		p.log.Errorf("action: receive_confirmation | result: fail | client_id: %v | error: unexpected confirmation %q", p.clientID, confirmation)
		return fmt.Errorf("unexpected batch confirmation: %q", confirmation)
	}

	p.log.Infof("action: batch_enviado | result: success | cantidad: %v | size: %v", len(bets), len(payload))
	return nil
}

func (p *Protocol) serializeBatchPayload(bets []Bet) []byte {
	var buffer bytes.Buffer

	for _, bet := range bets {
		formatBet(&buffer, bet)
		buffer.WriteString("\n")
	}

	buffer.WriteString(MESSAGE_END_BATCH)
	buffer.WriteString("\n")
	return buffer.Bytes()
}

// formatBet writes a single bet in pipe-delimited format
func formatBet(buffer *bytes.Buffer, bet Bet) {
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
