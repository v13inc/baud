package connection

import (
	"log"
)

type Hub struct {
	connections    map[Connection]bool
	broadcastChan  chan []byte
	registerChan   chan Connection
	unregisterChan chan Connection
}

func NewHub() Hub {
	return Hub{
		broadcastChan:  make(chan []byte),
		registerChan:   make(chan Connection),
		unregisterChan: make(chan Connection),
		connections:    make(map[Connection]bool),
	}
}

func (h Hub) Register(conn Connection) {
	log.Print("Registered connection")
	h.registerChan <- conn
}

func (h Hub) Unregister(conn Connection) {
	h.unregisterChan <- conn
}

func (h Hub) Broadcast(message []byte) {
	log.Print("Broadcasting message")
	h.broadcastChan <- message
}

func (h Hub) close(conn Connection) {
	delete(h.connections, conn)
	conn.Close()
}

func (h Hub) Run() {
	for {
		select {
		case c := <-h.registerChan:
			h.connections[c] = true
		case c := <-h.unregisterChan:
			if _, ok := h.connections[c]; ok {
				h.close(c)
			}
		case m := <-h.broadcastChan:
			for c := range h.connections {
				c.Send(m)
			}
		}
	}
}
