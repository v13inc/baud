package main

import (
	"encoding/json"
	"os"
)

type Settings struct {
	RedisAddress  string `json:"redis"`
	MessagesKey   string `json:"messagesKey"`
	PerPage       int64  `json:"messagesPerPage"`
	ServeAddress  string `json:"serve"`
	WsAddress     string `json:"websocket"`
	MessageWidth  int    `json:"messageWidth"`
	MessageHeight int    `json:"messageHeight"`
}

var settings = Settings{}

func initSettings() error {
	file, err := os.Open(*settingsPath)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&settings)
	if err != nil {
		return err
	}

	return nil
}
