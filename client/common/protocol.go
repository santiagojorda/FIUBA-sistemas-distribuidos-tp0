package common

import "encoding/json"

func serializeBet(bet Bet) ([]byte, error) {
	protocolMsg := ProtocolMessage{
		Name:      bet.Name,
		Lastname:  bet.Lastname,
		Dni:       bet.Dni,
		Birthdate: bet.Birthdate,
		Number:    bet.Number,
	}

	jsonData, err := json.Marshal(protocolMsg)
	if err != nil {
		return nil, err
	}

	log.Infof("action: serialize_bet | result: success | client_id: %v | bet: %v", bet.Dni, bet.Name + " " + bet.Lastname)
	return jsonData, nil
}
