package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/op/go-logging"
	"github.com/spf13/viper"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/common"
)

var log = logging.MustGetLogger("log")

// InitConfig Function that uses viper library to parse configuration parameters.
// Viper is configured to read variables from both environment variables and the
// config file ./config.yaml. Environment variables takes precedence over parameters
// defined in the configuration file. If some of the variables cannot be parsed,
// an error is returned
func InitConfig() (*viper.Viper, error) {
	v := viper.New()

	// Configure viper to read env variables with the CLI_ prefix
	v.AutomaticEnv()
	v.SetEnvPrefix("cli")
	// Use a replacer to replace env variables underscores with points. This let us
	// use nested configurations in the config file and at the same time define
	// env variables for the nested configurations
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Add env variables supported
	v.BindEnv("id")
	v.BindEnv("server.address")
	v.BindEnv("loop.period")

	// Add env variables for player information
	v.BindEnv("NOMBRE", "NOMBRE")
	v.BindEnv("APELLIDO", "APELLIDO")
	v.BindEnv("DOCUMENTO", "DOCUMENTO")
	v.BindEnv("NACIMIENTO", "NACIMIENTO")
	v.BindEnv("NUMERO", "NUMERO")

	// Try to read configuration from config file. If config file
	// does not exists then ReadInConfig will fail but configuration
	// can be loaded from the environment variables so we shouldn't
	// return an error in that case
	v.SetConfigFile("./config.yaml")
	if err := v.ReadInConfig(); err != nil {
		fmt.Printf("Configuration could not be read from config file. Using env variables instead")
	}

	return v, nil
}

// InitLogger Receives the log level to be set in go-logging as a string. This method
// parses the string and set the level to the logger. If the level string is not
// valid an error is returned
func InitLogger(logLevel string) error {
	baseBackend := logging.NewLogBackend(os.Stdout, "", 0)
	format := logging.MustStringFormatter(
		`%{time:2006-01-02 15:04:05} %{level:.5s}     %{message}`,
	)
	backendFormatter := logging.NewBackendFormatter(baseBackend, format)

	backendLeveled := logging.AddModuleLevel(backendFormatter)
	logLevelCode, err := logging.LogLevel(logLevel)
	if err != nil {
		return err
	}
	backendLeveled.SetLevel(logLevelCode, "")

	// Set the backends to be used.
	logging.SetBackend(backendLeveled)
	return nil
}

// PrintConfig Print all the configuration parameters of the program.
// For debugging purposes only
func PrintConfig(v *viper.Viper) {
	log.Infof("action: config | result: success | client_id: %s | server_address: %s  | log_level: %s",
		v.GetString("id"),
		v.GetString("server.address"),
		v.GetString("log.level"),
	)
}

func PrintPlayer(v *viper.Viper) {
	log.Infof("action: config | result: success | client_id: %s | player_name: %s | player_lastname: %s | player_dni: %s | player_birthdate: %s | player_number: %s",
		v.GetString("id"),
		v.GetString("NOMBRE"),
		v.GetString("APELLIDO"),
		v.GetString("DOCUMENTO"),
		v.GetString("NACIMIENTO"),
		v.GetString("NUMERO"),
	)
}

func main() {
	v, err := InitConfig()
	if err != nil {
		log.Criticalf("%s", err)
	}

	if err := InitLogger(v.GetString("log.level")); err != nil {
		log.Criticalf("%s", err)
	}

	// Print program config with debugging purposes
	PrintConfig(v)
	PrintPlayer(v)

	clientConfig := common.ClientConfig{
		ServerAddress: v.GetString("server.address"),
		ID:            v.GetString("id"),
	}

	player := common.Player{
		Name:      v.GetString("NOMBRE"),
		Lastname:  v.GetString("APELLIDO"),
		Dni:       v.GetString("DOCUMENTO"),
		Birthdate: v.GetString("NACIMIENTO"),
		Number:    v.GetString("NUMERO"),
	}

	client := common.NewClient(clientConfig, player)
	client.StartClientLoop()
}
