package common

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/op/go-logging"
)

const AMOUNTS_BETS_PER_MESSAGE = 1
const MESSAGE_FIN = "FIN"

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

// serialize formats the batch data according to protocol
// Format: <name|lastname|dni|birthdate|number>\n for each bet
// an the last line contains 'FIN' to indicate the end of the batch
func (p *ProtocolHandler) serialize(bet Bet) []byte {
	var buffer bytes.Buffer

	for _, bet := range []Bet{bet} {
		p.formatBet(&buffer, bet)
		buffer.WriteString("\n")
	}
	buffer.WriteString(MESSAGE_FIN + "\n")
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