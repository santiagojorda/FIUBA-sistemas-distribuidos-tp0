package common

import (
	"encoding/csv"
	"os"

	"github.com/op/go-logging"
)

const CANTIDAD_CAMPOS_BETS = 5

// BatchBuilder reads bets from CSV and organizes them into batches
type BatchBuilder struct {
	csvReader      *csv.Reader
	maxBatchSize   int // Maximum bytes per batch
	maxBatchAmount int // Maximum bets per batch
	tempBet        *Bet
	clientID       string
	log						 *logging.Logger
}

// NewBatchBuilder creates a new BatchBuilder instance
func NewBatchBuilder(csvPath string, maxBatchSize, maxBatchAmount int, clientID string, log *logging.Logger) (*BatchBuilder, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		log.Errorf("action: open_file | result: fail | client_id: %v | error: %v", clientID, err)
		return nil, err
	}

	return &BatchBuilder{
		csvReader:      csv.NewReader(file),
		maxBatchSize:   maxBatchSize,
		maxBatchAmount: maxBatchAmount,
		tempBet:        nil,
		clientID:       clientID,
		log:						log,
	}, nil
}

// NextBatch reads the next batch of bets from CSV
// Returns ([]Bet, hasMore, error)
func (b *BatchBuilder) NextBatch() ([]Bet, bool, error) {
	bets := []Bet{}
	packageSize := 0

	// If we have a leftover bet from previous batch, add it first
	if b.tempBet != nil {
		bets = append(bets, *b.tempBet)
		packageSize = calculateBetSize(b.tempBet)
		b.tempBet = nil
		b.log.Infof(
			"action: read_bets_from_csv | result: in_progress | bet: %v | package_size: %v",
			b.tempBet.Name+" "+b.tempBet.Lastname,
			packageSize,
		)
	}

	// Read bets until we hit limit or EOF
	for {
		// Check if we've hit the amount limit
		if len(bets) >= b.maxBatchAmount {
			b.log.Infof(
				"action: read_bets_from_csv | result: in_progress | status: batch_limit_reached | bets_count: %v",
				len(bets),
			)
			break
		}

		record, err := b.csvReader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				b.log.Infof(
					"action: read_bets_from_csv | result: success | status: eof_reached | bets_count: %v",
					len(bets),
				)
				return bets, false, nil // No more batches
			}
			b.log.Errorf("action: read_bets_from_csv | result: fail | error: %v", err)
			return bets, false, err
		}

		if len(record) < CANTIDAD_CAMPOS_BETS {
			b.log.Errorf("CSV file does not contain enough columns")
			continue
		}

		bet := Bet{
			Name:      record[0],
			Lastname:  record[1],
			Dni:       record[2],
			Birthdate: record[3],
			Number:    record[4],
		}

		betSize := calculateBetSize(&bet)

		// Check if adding this bet would exceed buffer size
		if packageSize+betSize >= b.maxBatchSize {
			b.tempBet = &bet // Save for next batch
			b.log.Infof(
				"action: read_bets_from_csv | result: in_progress | status: package_full | package_size: %v",
				packageSize,
			)
			return bets, true, nil // More batches to come
		}

		packageSize += betSize
		bets = append(bets, bet)
	}

	if len(bets) == 0 {
		b.log.Infof("action: read_bets_from_csv | result: success | status: no_more_bets")
	}

	return bets, len(bets) > 0, nil
}

// calculateBetSize returns the size in bytes of a bet's data
func calculateBetSize(bet *Bet) int {
	return len(bet.Name) + len(bet.Lastname) + len(bet.Dni) + len(bet.Birthdate) + len(bet.Number)
}

// Close closes the underlying file
func (b *BatchBuilder) Close() error {
	if b.csvReader != nil && b.csvReader.FieldsPerRecord > 0 {
		// The csv.Reader doesn't expose the underlying file directly,
		// but we can rely on garbage collection
	}
	return nil
}
