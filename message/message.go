package message

import (
	"encoding/json"
	"github.com/v13inc/baud/settings"
	"github.com/v13inc/baud/utils"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type JsonMessage struct {
	Name    string  `json:"name"`
	Date    float64 `json:"date"`
	Message string  `json:"message"`
}

type Message struct {
	Name    string `json:"name"`
	Date    int64  `json:"date"`
	Message string `json:"message"`
}

func New() Message {
	message := Message{}
	message.Date = time.Now().UTC().Unix()
	message.Name = settings.S.MessageName

	return message
}

func NewFromText(text string) (Message, error) {
	message := New()
	if err := message.FromText(text); err != nil {
		return message, err
	}

	return message, nil
}

func NewFromTextReader(reader io.Reader) (Message, error) {
	rawMessage, err := ioutil.ReadAll(reader)
	if err != nil {
		return Message{}, err
	}
	return NewFromText(string(rawMessage))
}

func NewFromJson(rawMessage string) (Message, error) {
	message := New()
	if err := message.FromJson(rawMessage); err != nil {
		return message, err
	}

	return message, nil
}

func NewFromJsonReader(reader io.Reader) (Message, error) {
	rawMessage, err := ioutil.ReadAll(reader)
	if err != nil {
		return Message{}, err
	}
	return NewFromJson(string(rawMessage))
}

func NewFromForm(r *http.Request) Message {
	message := New()
	message.FromForm(r)

	return message
}

func NewFromRawList(rawList []string) ([]Message, error) {
	messages := make([]Message, 0)
	for _, rawMessage := range rawList {
		message, err := NewFromJson(rawMessage)
		if err != nil {
			return messages, err
		}
		messages = append(messages, message)
	}

	return messages, nil
}

func (m *Message) Json() (string, error) {
	str, err := json.Marshal(m)
	return string(str), err
}

func (m *Message) Format() string {
	lines := strings.Split(m.Message, "\n")
	diff := time.Now().UTC().Unix() - m.Date
	seconds := strconv.FormatInt(diff, 10) + " seconds ago"
	header := m.Name + " - " + seconds

	out := "┌" + strings.Repeat("─", settings.S.MessageWidth) + "┐\n"
	out += "│" + utils.Center(header, ' ', settings.S.MessageWidth) + "│\n"
	out += "├" + strings.Repeat("─", settings.S.MessageWidth) + "┤\n"
	for _, line := range lines {
		out += "│" + utils.Pad(line, ' ', settings.S.MessageWidth) + "│\n"
	}
	out += "└" + strings.Repeat("─", settings.S.MessageWidth) + "┘\n\n"

	return out
}

func (m *Message) FromJson(rawMessage string) error {
	jsonMessage := JsonMessage{}
	if err := json.Unmarshal([]byte(rawMessage), &jsonMessage); err != nil {
		return err
	}

	m.Name = jsonMessage.Name
	m.Message = jsonMessage.Message
	m.Date = int64(jsonMessage.Date)

	return nil
}

func (m *Message) FromText(text string) error {
	body := ""
	lines := strings.Split(text, "\n")

	if len(lines) > settings.S.MessageHeight {
		lines = lines[0:settings.S.MessageHeight]
	}

	for _, line := range lines {
		body += utils.Pad(line, ' ', settings.S.MessageWidth) + "\n"
	}

	if len(lines) < settings.S.MessageHeight {
		diff := settings.S.MessageHeight - len(lines)
		body += strings.Repeat(utils.Pad("", ' ', settings.S.MessageWidth)+"\n", diff)
	}

	m.Message = body

	return nil
}

func (m *Message) FromRequestArgs(r *http.Request) {
	m.Name = utils.ArgDef(r, "name", m.Name)
	m.Message = utils.ArgDef(r, "message", m.Message)
}

func (m *Message) FromForm(r *http.Request) {
	m.Name = utils.FormDef(r, "name", m.Name)
	m.FromText(utils.FormDef(r, "message", m.Message))
}
