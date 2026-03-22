package common

type ProtocolMessage struct {
	Name      string `json:"name"`
	Lastname  string `json:"lastname"`
	Dni       string `json:"dni"`
	Birthdate string `json:"birthdate"`
	Number    string `json:"number"`
}

type Bet struct {
	Name      string
	Lastname  string
	Dni       string
	Birthdate string
	Number    string
}
