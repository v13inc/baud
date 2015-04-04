package stream

import (
	"github.com/garyburd/redigo/redis"
	"github.com/v13inc/baud/repo"
	"github.com/v13inc/baud/utils"
	"net/http"
	"text/template"
)

type Manager struct {
	redis    redis.Conn
	streams  map[string]Stream
	root     Stream
	template *template.Template
}

func NewManager(r redis.Conn, t *template.Template) Manager {
	m := Manager{redis: r, template: t, streams: make(map[string]Stream)}
	m.root = m.NewStream("messages")
	return m
}

func (m *Manager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		m.root.ServeHTTP(w, r)
		return
	}

	name, err := utils.StreamSlug(r.URL)
	if err != nil || name == m.root.name {
		utils.Error(w, r, "Error serving stream", err)
		return
	}

	m.ServeStream(name, w, r)
}

func (m *Manager) ServeStream(name string, w http.ResponseWriter, r *http.Request) {
	stream, found := m.streams[name]
	if !found {
		stream = m.NewStream(name)
	}

	stream.ServeHTTP(w, r)
}

func (m *Manager) NewStream(name string) Stream {
	stream := New(name, repo.NewMessages(m.redis, name), m.template)
	m.streams[name] = stream

	go stream.Run()

	return stream
}
