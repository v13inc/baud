package connection

import (
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Websocket struct {
	h         *Hub
	ws        *websocket.Conn
	broadcast chan []byte
	Message   chan []byte
}

func NewWebsocket(h *Hub, ws *websocket.Conn) Websocket {
	return Websocket{h: h, ws: ws, broadcast: make(chan []byte, bufferSize), Message: make(chan []byte, bufferSize)}
}

func (c Websocket) Run() {
	c.h.Register(&c)
	go c.writePump()
	c.readPump()
}

func (c Websocket) Send(message []byte) {
	c.broadcast <- message
}

func (c Websocket) Close() {
	close(c.broadcast)
	close(c.Message)
}

func (c Websocket) write(messageType int, message []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(messageType, message)
}

func (c Websocket) readPump() {
	defer func() {
		c.h.Unregister(&c)
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		c.Message <- message
		//_, messageReader, err := c.ws.NextReader()
		/*
			message, err := newMessageFromJson(messageReader)
			if err != nil {
				log.Println("Error creating new message:", err)
				break
			}
			messageString, err := message.save()
			if err != nil {
				log.Println("Error saving new message:", err)
				break
			}
		*/
	}
}

func (c Websocket) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.broadcast:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
