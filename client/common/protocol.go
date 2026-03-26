package common

import (
	"bytes"
)

const AMOUNTS_BETS_PER_MESSAGE = 1
const MESSAGE_FIN = "FIN"

// serialize formats the batch data according to protocol
// Format: <name|lastname|dni|birthdate|number>\n for each bet
func SerializeBets(bets []Bet) []byte {
	var buffer bytes.Buffer

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