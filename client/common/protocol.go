package common

import "encoding/json"

func serializePlayer(player Player) ([]byte, error) {
	protocolMsg := ProtocolMessage{
		Name:      player.Name,
		Lastname:  player.Lastname,
		Dni:       player.Dni,
		Birthdate: player.Birthdate,
		Number:    player.Number,
	}

	jsonData, err := json.Marshal(protocolMsg)
	if err != nil {
		return nil, err
	}

	log.Infof("action: serialize_player | result: success | client_id: %v | player: %v", player.Name, player.Lastname)
	return jsonData, nil
}
