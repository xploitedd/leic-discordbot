package misc

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

// Configuration stores a config file
type Configuration struct {
	OwnerID       *string   `json:"owner_id"`
	DiscordToken  *string   `json:"discord_token"`
	DialogFlowID  *string   `json:"dialogflow_id"`
	CommandPrefix *string   `json:"command_prefix"`
	PlayingWith   []*string `json:"playing_with"`
}

// Config is the main configuration file of the bot
var Config Configuration

// LoadConfig loads a file into the main configuration
func LoadConfig(filename string) error {
	// load configuration file
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		return errors.New("error reading the configuration file: " + err.Error())
	}

	// try to parse the json
	Config = Configuration{}
	err = json.Unmarshal(data, &Config)
	if err != nil {
		return errors.New("error while parsing the config file: " + err.Error())
	}

	// check if the required configuration fields are available
	if Config.DiscordToken == nil ||
		Config.CommandPrefix == nil ||
		Config.DialogFlowID == nil ||
		Config.OwnerID == nil {
		return errors.New("please specify the required fields in the configuration file")
	}

	return nil
}
