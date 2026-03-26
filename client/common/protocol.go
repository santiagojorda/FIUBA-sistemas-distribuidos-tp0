package common

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/op/go-logging"
)

const AGENCY_CSV_PATH_TEMPLATE = "/.data/agency-%s.csv"
const MESSAGE_FIN = "FIN"

// Protocol manages serialization and sending of bets batches
type Protocol struct {
	connection *Connection
	log        *logging.Logger
	clientID   string
	maxAmount  int
	maxSize    int
}

// NewProtocol creates a new Protocol instance
func NewProtocol(config ClientConfig, connection *Connection) (*Protocol, error) {

	if config.MaxBatchAmount <= 0 {
		config.Log.Warningf("Invalid max batch amount %v", config.MaxBatchAmount)
		return nil, fmt.Errorf("invalid max batch amount: %d", config.MaxBatchAmount)
	}

	if config.MaxBatchSize <= 0 {
		config.Log.Warningf("Invalid max batch size %v", config.MaxBatchSize)
		return nil, fmt.Errorf("invalid max batch size: %d", config.MaxBatchSize)
	}

	if connection == nil {
		return nil, fmt.Errorf("connection cannot be nil")
	}

	return &Protocol{
		connection: connection,
		log:        config.Log,
		clientID:   config.ID,
		maxAmount:  config.MaxBatchAmount,
		maxSize:    config.MaxBatchSize,
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

// SendBatchesFromCSV reads the agency CSV and sends batches to server.
func (p *Protocol) SendBatchesFromCSV() error {
	file, err := p.openAgencyCSVFile()
	if err != nil {
		return err
	}
	defer file.Close()

	return p.sendBatchesLoop(file)
}

func (p *Protocol) sendBatch(bets []Bet) error {
	payload := p.serializeBatchPayload(bets)
	if err := p.connection.Send(payload); err != nil {
		return err
	}

	confirmation, err := p.connection.Receive()
	if err != nil || confirmation == "" {
		p.log.Errorf("action: receive_confirmation | result: fail | error: %v", err)
		return fmt.Errorf("failed to receive confirmation: %v", err)
	}

	p.log.Infof("action: batch_enviado | result: success | cantidad: %v | size: %v", len(bets), len(payload))
	return nil
}

func (p *Protocol) serializeBatchPayload(bets []Bet) []byte {
	var buffer bytes.Buffer
	buffer.WriteString(strconv.Itoa(len(bets)))
	buffer.WriteString("\n")

	for _, bet := range bets {
		formatBet(&buffer, bet)
		buffer.WriteString("\n")
	}
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
