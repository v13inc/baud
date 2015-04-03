package connection

import (
	"net/http"
)

type Longpoll struct {
	h         *Hub
	w         http.ResponseWriter
	broadcast chan []byte
	Out       chan []byte
}

func NewLongpoll(h *Hub, w http.ResponseWriter) Longpoll {
	return Longpoll{h: h, w: w, broadcast: make(chan []byte, bufferSize), Out: make(chan []byte)}
}

func (c Longpoll) Run() {
	c.h.Register(&c)
	c.writePump()
}
func (c Longpoll) Send(message []byte) {
	c.broadcast <- message
}

func (c Longpoll) Close() {
	close(c.broadcast)
}

func (c Longpoll) Write(bytes []byte) (int, error) {
	num, err := c.w.Write(bytes)
	if err != nil {
		return num, err
	}
	if f, ok := c.w.(http.Flusher); ok {
		f.Flush()
	}
	return num, nil
}

func (c Longpoll) writePump() {
	for {
		select {
		case message, ok := <-c.broadcast:
			if !ok {
				return
			}
			c.Out <- message
		}
	}
}
