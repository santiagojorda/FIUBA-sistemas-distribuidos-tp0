package common

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

// BatchBuilder reads bets from CSV and creates batches with amount/size limits.
type BatchBuilder struct {
	file      *os.File
	reader    *csv.Reader
	maxAmount int
	maxSize   int
	carry     Bet
}

func NewBatchBuilder(clientID string, maxAmount int, maxSize int) (*BatchBuilder, error) {
	if maxAmount <= 0 {
		return nil, fmt.Errorf("invalid max batch amount: %d", maxAmount)
	}
	if maxSize <= 0 {
		return nil, fmt.Errorf("invalid max batch size: %d", maxSize)
	}

	file, err := os.Open(fmt.Sprintf(AGENCY_CSV_PATH_TEMPLATE, clientID))
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %v", err)
	}

	return &BatchBuilder{
		file:      file,
		reader:    csv.NewReader(file),
		maxAmount: maxAmount,
		maxSize:   maxSize,
		carry:     Bet{},
	}, nil
}

func (b *BatchBuilder) Close() error {
	if b.file == nil {
		return nil
	}

	return b.file.Close()
}

// NextBatch returns the next batch, plus a flag indicating if EOF was reached.
func (b *BatchBuilder) NextBatch() ([]Bet, bool, error) {
	bets, packageSize, err := b.initializeBatchWithCarry()
	if err != nil {
		return nil, false, err
	}

	for {
		if len(bets) >= b.maxAmount {
			return bets, false, nil
		}

		bet, reachedEOF, err := readNextValidBet(b.reader)
		if err != nil {
			return nil, false, err
		}
		if reachedEOF {
			return bets, true, nil
		}

		betSize := serializedBetSize(bet)
		if betSize > b.maxSize {
			return nil, false, fmt.Errorf("single bet exceeds max packet size")
		}

		if packageSize+betSize > b.maxSize {
			b.carry = bet
			return bets, false, nil
		}

		bets = append(bets, bet)
		packageSize += betSize
	}
}

func (b *BatchBuilder) initializeBatchWithCarry() ([]Bet, int, error) {
	bets := []Bet{}
	packageSize := 0

	if b.carry == (Bet{}) {
		return bets, packageSize, nil
	}

	carrySize := serializedBetSize(b.carry)
	if carrySize > b.maxSize {
		return nil, 0, fmt.Errorf("single bet exceeds max packet size")
	}

	bets = append(bets, b.carry)
	b.carry = Bet{}
	packageSize = carrySize
	return bets, packageSize, nil
}

func readNextValidBet(reader *csv.Reader) (Bet, bool, error) {
	for {
		record, err := reader.Read()
		if err == io.EOF {
			return Bet{}, true, nil
		}
		if err != nil {
			return Bet{}, false, fmt.Errorf("failed to read CSV record: %v", err)
		}

		bet, err := parseBetRecord(record)
		if err != nil {
			continue
		}

		return bet, false, nil
	}
}

func parseBetRecord(record []string) (Bet, error) {
	if len(record) < 5 {
		return Bet{}, fmt.Errorf("invalid record: expected 5 fields, got %d", len(record))
	}

	return Bet{
		Name:      record[0],
		Lastname:  record[1],
		Dni:       record[2],
		Birthdate: record[3],
		Number:    record[4],
	}, nil
}

func serializedBetSize(bet Bet) int {
	// 4 pipes + trailing newline
	return len(bet.Name) + len(bet.Lastname) + len(bet.Dni) + len(bet.Birthdate) + len(bet.Number) + 5
}
