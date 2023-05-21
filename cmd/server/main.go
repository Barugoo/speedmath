package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

const port = ":8080"

func main() {
	game := NewGame()
	s := NewServer(game)

	go game.Run()

	r := mux.NewRouter()
	r.HandleFunc("/ws", s.ConnectGame)
	log.Println("running websocket server on " + port)
	http.ListenAndServe(port, r)
}
