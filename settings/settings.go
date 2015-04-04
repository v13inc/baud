package settings

import (
	"encoding/json"
	"os"
)

type Settings struct {
	RedisAddress  string `json:"redis"`
	MessagesKey   string `json:"messagesKey"`
	PerPage       int    `json:"messagesPerPage"`
	ServeAddress  string `json:"serve"`
	WsAddress     string `json:"websocket"`
	MessageWidth  int    `json:"messageWidth"`
	MessageHeight int    `json:"messageHeight"`
	MessageName   string `json:"messageName"`
	StreamPrefix  string `json:"streamPrefix"`
}

var S Settings

func Init(settingsPath string) error {
	S = Settings{}
	file, err := os.Open(settingsPath)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&S)
	if err != nil {
		return err
	}

	return nil
}
