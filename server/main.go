package main

import (
	"flag"
	"github.com/garyburd/redigo/redis"
	"github.com/v13inc/baud/repo"
	"github.com/v13inc/baud/settings"
	"github.com/v13inc/baud/stream"
	"log"
	"net/http"
	"text/template"
)

var (
	settingsPath = flag.String("settings", "settings.json", "Settings path")
	r            redis.Conn
	messages     repo.Messages
	temp         *template.Template
	home         stream.Stream
)

func main() {
	flag.Parse()

	err := settings.Init(*settingsPath)
	if err != nil {
		log.Fatal("Error initializing settings: ", err)
	}

	r, err = redis.Dial("tcp", settings.S.RedisAddress)
	if err != nil {
		log.Fatal("Error connecting to Redis: ", err)
	}

	temp, err = template.ParseFiles("index.html")
	if err != nil {
		log.Fatal("Error parsing template: ", err)
	}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	manager := stream.NewManager(r, temp)
	http.Handle("/", &manager)

	if err = http.ListenAndServe(settings.S.ServeAddress, nil); err != nil {
		log.Fatal("Error setting up http.ListenAndServe: ", err)
	}
}
