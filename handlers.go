package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"
)

type Context struct {
	Messages  string
	WsAddress string
}

var indexTemplate *template.Template

func initHandlers() error {
	var err error
	indexTemplate, err = template.ParseFiles("index.html")
	if err != nil {
		return err
	}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/api/txt/", handleMessagesText)
	http.HandleFunc("/api/messages/", handleMessages)
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/ws", handleWs)

	return nil
}

func handleError(c http.ResponseWriter, req *http.Request, message string, err error) {
	log.Println(message, err)
	fmt.Fprintf(c, "{}")
}

func handleIndex(c http.ResponseWriter, req *http.Request) {
	messages, err := getMessagesJson(settings.PerPage)
	if err != nil {
		log.Println("Error getting JSON message list (index handler):", err)
		return
	}
	indexTemplate.Execute(c, &Context{Messages: messages, WsAddress: settings.WsAddress})
}

func handleMessages(c http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		handleMessagesGet(c, req)
	case "POST":
		handleMessagesPost(c, req)
	}
}

func handleMessagesGet(c http.ResponseWriter, req *http.Request) {
	var err error
	amount := settings.PerPage
	amountParam := req.URL.Query()["amount"]
	if amountParam != nil && len(amountParam) > 0 {
		amount, err = strconv.ParseInt(amountParam[0], 10, 64)
		if err != nil {
			log.Println("Error parsing amount parameter:", err)
			amount = settings.PerPage
		}
	}

	messages, err := getMessagesJson(amount)
	if err != nil {
		handleError(c, req, "Error getting JSON message list (messagesGet handler):", err)
		return
	}
	fmt.Fprint(c, messages)
}

func handleMessagesPost(c http.ResponseWriter, req *http.Request) {
	message, err := newMessageFromJson(req.Body)
	if err != nil {
		handleError(c, req, "Error creating new message from JSON (messagesPost handler):", err)
		return
	}

	messageJson, err := message.save()
	if err != nil {
		handleError(c, req, "Error saving message (messagesPost handler):", err)
		return
	}

	fmt.Fprintf(c, messageJson)
}

func handleMessagesText(c http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		handleError(c, req, "Message empty:", nil)
		return
	}

	name := "Anonymous"
	nameParam := req.URL.Query()["name"]
	if nameParam != nil && len(nameParam) > 0 {
		name = nameParam[0]
	}

	message, err := newMessageFromText(req.Body, name)
	if err != nil {
		handleError(c, req, "Error creating new message from text (messagesText handler):", err)
		return
	}

	messageJson, err := message.save()
	if err != nil {
		handleError(c, req, "Error saving message (messagesText handler):", err)
		return
	}

	h.broadcast <- []byte(messageJson)

	fmt.Fprintf(c, messageJson)
}

func handleWs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	c := &connection{send: make(chan []byte, 256), ws: ws}
	h.register <- c
	go c.writePump()
	c.readPump()
}
