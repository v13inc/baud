package stream

import (
	"fmt"
	"github.com/v13inc/baud/connection"
	"github.com/v13inc/baud/message"
	"github.com/v13inc/baud/repo"
	"github.com/v13inc/baud/settings"
	"github.com/v13inc/baud/utils"
	"log"
	"net/http"
	"text/template"
)

type context struct {
	Messages string
}

type MessageCallback func(*message.Message, []byte)
type Stream struct {
	name     string
	hub      connection.Hub
	messages repo.Messages
	template *template.Template
}

func New(name string, messages repo.Messages, temp *template.Template) Stream {
	hub := connection.NewHub()
	return Stream{name: name, hub: hub, messages: messages, template: temp}
}

func (s *Stream) Run() {
	s.hub.Run()
}

func (s *Stream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.Get(w, r)
	case "POST":
		s.Post(w, r)
	default:
		s.MethodError(w, r)
	}
}

func (s *Stream) MethodError(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.ErrNotSupported.Error(), http.StatusMethodNotAllowed)
}

func (s *Stream) Get(w http.ResponseWriter, r *http.Request) {
	switch {
	case utils.IsWebsocketReq(r):
		s.Websocket(w, r)
	case utils.IsJsonReq(r):
		s.GetJson(w, r)
	case utils.IsHtmlReq(r):
		s.GetHtml(w, r)
	default:
		s.GetText(w, r)
	}
}

func (s *Stream) Websocket(w http.ResponseWriter, r *http.Request) {
	ws, err := connection.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.Error(w, r, "Error upgrading to websocket connection: ", err)
		return
	}
	c := connection.NewWebsocket(&s.hub, ws)
	// incoming messages from client
	go s.WatchChannel(c.Message, func(m *message.Message, raw []byte) {
		_, err := s.AddAndBroadcast(m)
		if err != nil {
			log.Print("Error saving or broadcasting: ", err)
		}
	})
	c.Run()
}

func (s *Stream) GetJson(w http.ResponseWriter, r *http.Request) {
	amount := utils.ArgIntDef(r, "n", settings.S.PerPage)
	messages, err := s.messages.JsonList(amount, false)
	if err != nil {
		utils.Error(w, r, "Error getting JSON message list: ", err)
		return
	}
	fmt.Fprint(w, messages)

	if utils.ArgBool(r, "f") {
		s.StreamJson(w, r)
	}
}

func (s *Stream) GetHtml(w http.ResponseWriter, r *http.Request) {
	amount := utils.ArgIntDef(r, "n", settings.S.PerPage)
	messages, err := s.messages.JsonList(amount, false)
	if err != nil {
		utils.Error(w, r, "Error getting JSON message list: ", err)
		return
	}
	s.template.Execute(w, &context{Messages: messages})
}

func (s *Stream) GetText(w http.ResponseWriter, r *http.Request) {
	amount := utils.ArgIntDef(r, "n", settings.S.PerPage)
	messages, err := s.messages.List(amount, false)
	if err != nil {
		utils.Error(w, r, "Error getting messages: ", err)
		return
	}
	for i := len(messages) - 1; i >= 0; i-- {
		fmt.Fprint(w, messages[i].Format())
	}

	if utils.ArgBool(r, "f") {
		utils.Flush(w)
		s.StreamText(w, r)
	}
}

func (s *Stream) StreamText(w http.ResponseWriter, r *http.Request) {
	c := connection.NewLongpoll(&s.hub, w)
	go s.WatchChannel(c.Out, func(m *message.Message, raw []byte) {
		log.Print("New long poll message", m.Message)
		num, err := w.Write([]byte(m.Format()))
		if err != nil {
			log.Print("Error writing message:", err)
			return
		}
		log.Print("Wrote bytes:", num)
		utils.Flush(w)
	})
	c.Run()
}

func (s *Stream) StreamJson(w http.ResponseWriter, r *http.Request) {
	c := connection.NewLongpoll(&s.hub, w)
	go s.WatchChannel(c.Out, func(m *message.Message, raw []byte) {
		log.Print("New long poll message", m.Message)
		num, err := w.Write(raw)
		if err != nil {
			log.Print("Error writing message:", err)
			return
		}
		log.Print("Wrote bytes:", num)
		utils.Flush(w)
	})
	c.Run()
}

func (s *Stream) Post(w http.ResponseWriter, r *http.Request) {
	switch {
	case utils.IsJsonReq(r):
		s.PostJson(w, r)
	case utils.HasFormData(r):
		s.PostForm(w, r)
	default:
		s.PostText(w, r)
	}
}

func (s *Stream) PostJson(w http.ResponseWriter, r *http.Request) {
	m, err := message.NewFromJsonReader(r.Body)
	s.AddMessage(w, r, &m, err)
}

func (s *Stream) PostForm(w http.ResponseWriter, r *http.Request) {
	m := message.NewFromForm(r)
	s.AddMessage(w, r, &m, nil)
}

func (s *Stream) PostText(w http.ResponseWriter, r *http.Request) {
	m, err := message.NewFromTextReader(r.Body)
	s.AddMessage(w, r, &m, err)
}

func (s *Stream) AddMessage(w http.ResponseWriter, r *http.Request, m *message.Message, err error) {
	if err != nil {
		utils.Error(w, r, "Error reading message: ", err)
		return
	}
	m.FromRequestArgs(r)
	messageJson, err := s.AddAndBroadcast(m)
	if err != nil {
		utils.Error(w, r, "Error adding message: ", err)
		return
	}
	s.ShowMessage(w, r, m, messageJson)
}

func (s *Stream) ShowMessage(w http.ResponseWriter, r *http.Request, m *message.Message, messageJson string) {
	switch {
	case utils.IsJsonReq(r):
		fmt.Fprint(w, messageJson)
	case utils.IsHtmlReq(r):
		s.GetHtml(w, r)
	default:
		fmt.Fprint(w, m.Format())
	}
}

func (s *Stream) WatchChannel(c chan []byte, callback MessageCallback) {
	for {
		select {
		case raw := <-c:
			m, err := message.NewFromJson(string(raw))
			if err != nil {
				log.Print("Error parsing message:", err)
				continue
			}
			callback(&m, raw)
		}
	}
}

func (s *Stream) AddAndBroadcast(m *message.Message) (string, error) {
	messageJson, err := s.messages.Add(m)
	if err != nil {
		return "", err
	}
	s.hub.Broadcast([]byte(messageJson))
	return messageJson, nil
}

func (s *Stream) Broadcast(m *message.Message) (string, error) {
	messageJson, err := m.Json()
	if err != nil {
		return "", err
	}
	s.hub.Broadcast([]byte(messageJson))
	return messageJson, nil
}
