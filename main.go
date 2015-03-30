package main

import (
	"flag"
	"github.com/garyburd/redigo/redis"
	"log"
	"net/http"
)

var (
	settingsPath = flag.String("settings", "settings.json", "Settings path")
	r            redis.Conn
)

func main() {
	flag.Parse()

	err := initSettings()
	if err != nil {
		log.Fatal("Error initializing settings", err)
	}

	go h.run()

	r, err = redis.Dial("tcp", settings.RedisAddress)
	if err != nil {
		log.Fatal("Error connecting to Redis:", err)
	}

	if err = initHandlers(); err != nil {
		log.Fatal("Error initializing handlers:", err)
	}

	if err = http.ListenAndServe(settings.ServeAddress, nil); err != nil {
		log.Fatal("Error setting up http.ListenAndServe:", err)
	}
}
