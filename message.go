package main

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"io"
	"io/ioutil"
	"strings"
	"time"
)

type Message struct {
	Name    string `json:"name"`
	Date    int64  `json:"date"`
	Message string `json:"message"`
}

func (message *Message) json() (string, error) {
	str, err := json.Marshal(message)
	return string(str), err
}

func (message *Message) save() (string, error) {
	if message.Date == 0 {
		message.Date = time.Now().UTC().Unix()
	}

	str, err := message.json()
	if err != nil {
		return "", err
	}
	_, err = r.Do("zadd", settings.MessagesKey, message.Date, str)
	if err != nil {
		return "", err
	}
	return str, nil
}

func (message *Message) fromJson(reader io.Reader) error {
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&message); err != nil {
		return err
	}

	return nil
}

func (message *Message) fromText(reader io.Reader) error {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	text := string(bytes)

	body := ""
	lines := strings.Split(text, "\n")

	if len(lines) > settings.MessageHeight {
		lines = lines[0:settings.MessageHeight]
	}

	for _, line := range lines {
		body += pad(line, ' ', settings.MessageWidth) + "\n"
	}

	if len(lines) < settings.MessageHeight {
		diff := settings.MessageHeight - len(lines)
		body += strings.Repeat(pad("", ' ', settings.MessageWidth)+"\n", diff)
	}

	message.Message = body

	return nil
}

func newMessageFromText(reader io.Reader, name string) (Message, error) {
	var message Message
	if err := message.fromText(reader); err != nil {
		return message, err
	}

	message.Name = name

	return message, nil
}

func newMessageFromJson(reader io.Reader) (Message, error) {
	var message Message
	if err := message.fromJson(reader); err != nil {
		return message, err
	}

	return message, nil
}

func getMessagesRaw(amount int64) ([]string, error) {
	return redis.Strings(r.Do("zrevrange", settings.MessagesKey, 0, amount-1))
}

func getMessagesJson(amount int64) (string, error) {
	rawMessages, err := getMessagesRaw(amount)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("[%s]", strings.Join(rawMessages, ",")), nil
}
