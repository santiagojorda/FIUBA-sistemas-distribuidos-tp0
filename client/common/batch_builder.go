package common

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

func (p *Protocol) openAgencyCSVFile() (*os.File, error) {
	file, err := os.Open(fmt.Sprintf(AGENCY_CSV_PATH_TEMPLATE, p.clientID))
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %v", err)
	}

	return file, nil
}

func (p *Protocol) sendBatchesLoop(file *os.File) error {
	reader := csv.NewReader(file)
	carry := Bet{}

	for {
		bets, pendingBet, reachedEOF, err := p.readNextBatch(reader, carry)
		if err != nil {
			return err
		}

		carry = pendingBet
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
	bets, packageSize, err := p.initializeBatchWithCarry(carry)
	if err != nil {
		return nil, Bet{}, false, err
	}

	for len(bets) < p.maxAmount {
		bet, reachedEOF, err := readNextValidBet(reader)
		if err != nil {
			return nil, Bet{}, false, err
		}
		if reachedEOF {
			return bets, Bet{}, true, nil
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

	return bets, Bet{}, false, nil
}

func (p *Protocol) initializeBatchWithCarry(carry Bet) ([]Bet, int, error) {
	bets := []Bet{}
	packageSize := 0

	if carry == (Bet{}) {
		return bets, packageSize, nil
	}

	carrySize := serializedBetSize(carry)
	if carrySize > p.maxSize {
		return nil, 0, fmt.Errorf("single bet exceeds max packet size")
	}

	bets = append(bets, carry)
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
