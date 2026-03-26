package common

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/op/go-logging"
)

const MESSAGE_FIN = "FIN"
const defaultMaxBatchAmount = 1
const defaultMaxBatchSize = 8 * 1024

// Protocol manages serialization and sending of bets batches
type Protocol struct {
	connection *Connection
	log        *logging.Logger
	clientID   string
	maxAmount  int
	maxSize    int
}

// NewProtocol creates a new Protocol instance
func NewProtocol(connection *Connection, log *logging.Logger, clientID string, maxAmount int, maxSize int) *Protocol {
	if maxAmount <= 0 {
		maxAmount = defaultMaxBatchAmount
	}
	if maxSize <= 0 {
		maxSize = defaultMaxBatchSize
	}

	return &Protocol{
		connection: connection,
		log:        log,
		clientID:   clientID,
		maxAmount:  maxAmount,
		maxSize:    maxSize,
	}
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
	file, err := os.Open(fmt.Sprintf("/.data/agency-%s.csv", p.clientID))
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	carry := Bet{}

	for {
		bets, nextCarry, reachedEOF, err := p.readNextBatch(reader, carry)
		if err != nil {
			return err
		}

		carry = nextCarry
		if len(bets) == 0 {
			if reachedEOF {
				return nil
			}
			continue
		}

		if err := p.sendBatch(bets); err != nil {
			return err
		}

		if reachedEOF && carry == (Bet{}) {
			return nil
		}
	}
}

func (p *Protocol) readNextBatch(reader *csv.Reader, carry Bet) ([]Bet, Bet, bool, error) {
	bets := []Bet{}
	packageSize := 0

	if carry != (Bet{}) {
		carrySize := serializedBetSize(carry)
		if carrySize <= p.maxSize {
			bets = append(bets, carry)
			packageSize = carrySize
		} else {
			return nil, Bet{}, false, fmt.Errorf("single bet exceeds max packet size")
		}
		carry = Bet{}
	}

	for len(bets) < p.maxAmount {
		record, err := reader.Read()
		if err == io.EOF {
			return bets, Bet{}, true, nil
		}
		if err != nil {
			return nil, Bet{}, false, fmt.Errorf("failed to read CSV record: %v", err)
		}
		if len(record) < 5 {
			continue
		}

		bet := Bet{
			Name:      record[0],
			Lastname:  record[1],
			Dni:       record[2],
			Birthdate: record[3],
			Number:    record[4],
		}

		betSize := serializedBetSize(bet)
		if betSize > p.maxSize {
			return nil, Bet{}, false, fmt.Errorf("single bet exceeds max packet size")
		}

		if packageSize+betSize > p.maxSize {
			return bets, bet, false, nil
		}

		bets = append(bets, bet)
		packageSize += betSize
	}

	return bets, carry, false, nil
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

	for _, bet := range bets {
		p.log.Infof("action: apuesta_enviada | result: success | client_id: %v | dni: %v | numero: %v", p.clientID, bet.Dni, bet.Number)
	}

	p.log.Infof("action: batch_enviado | result: success | cantidad: %v", len(bets))
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

func serializedBetSize(bet Bet) int {
	// 4 pipes + trailing newline
	return len(bet.Name) + len(bet.Lastname) + len(bet.Dni) + len(bet.Birthdate) + len(bet.Number) + 5
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
