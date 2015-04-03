package repo

import (
	"github.com/garyburd/redigo/redis"
	"github.com/v13inc/baud/message"
	"strings"
)

type Messages struct {
	redis redis.Conn
	key   string
}

func NewMessages(r redis.Conn, key string) Messages {
	return Messages{redis: r, key: key}
}

func (m *Messages) Add(mess *message.Message) (string, error) {
	messageJson, err := mess.Json()
	if err != nil {
		return "", err
	}
	_, err = m.redis.Do("zadd", m.key, mess.Date, messageJson)
	return messageJson, err
}

func (m *Messages) RawRange(start int, end int, reverse bool) ([]string, error) {
	var method string
	if reverse {
		method = "zrange"
	} else {
		method = "zrevrange"
	}

	return redis.Strings(m.redis.Do(method, m.key, start, end))
}

func (m *Messages) JsonRange(start int, end int, reverse bool) (string, error) {
	raw, err := m.RawRange(start, end, reverse)
	if err != nil {
		return "", err
	}
	json := strings.Join(raw, ",")
	return "[" + json + "]", nil
}

func (m *Messages) Range(start int, end int, reverse bool) ([]message.Message, error) {
	rawMessages, err := m.RawRange(start, end, reverse)
	if err != nil {
		return make([]message.Message, 0), err
	}

	return message.NewFromRawList(rawMessages)
}

func (m *Messages) RawList(number int, reverse bool) ([]string, error) {
	return m.RawRange(0, number-1, reverse)
}

func (m *Messages) JsonList(number int, reverse bool) (string, error) {
	return m.JsonRange(0, number-1, reverse)
}

func (m *Messages) List(number int, reverse bool) ([]message.Message, error) {
	return m.Range(0, number-1, reverse)
}
