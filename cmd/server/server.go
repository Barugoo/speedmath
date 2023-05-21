package main

import (
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Server struct {
	game     *Game
	upgrader *websocket.Upgrader
	aw       *AnswerWorker
}

func NewServer(game *Game) *Server {
	return &Server{
		game: game,
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		aw: NewAnswerWorker(game.roundCh),
	}
}

func (s *Server) ConnectGame(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("username")
	if len(name) == 0 {
		http.Error(w, "no username was provded", http.StatusBadRequest)
		return
	}

	outCh := make(chan any, 10)
	s.game.loginCh <- &EventLogin{
		Username: name,
		OutputCh: outCh,
	}

	wsconn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	wg := &sync.WaitGroup{}

	wg.Add(2)
	go s.handleIncomingMessages(wg, wsconn, name)
	go s.handleOutcomingMessages(wg, wsconn, outCh)

	log.Printf("user %s connected, handling events..", name)

	wg.Wait()
	wsconn.Close()
}

func (s *Server) handleOutcomingMessages(wg *sync.WaitGroup, wsconn *websocket.Conn, outCh chan any) {
	defer wg.Done()
	for out := range outCh {
		if err := wsconn.WriteJSON(&out); err != nil {
			log.Println(err)
			continue
		}
	}
}

func (s *Server) handleIncomingMessages(wg *sync.WaitGroup, wsconn *websocket.Conn, username string) {
	defer wg.Done()
	for {
		var e EventAnswer
		if err := wsconn.ReadJSON(&e); err != nil {
			var e *websocket.CloseError
			if errors.As(err, &e) {
				s.game.logoutCh <- &EventLogout{
					Username: username,
				}
				return
			}
			log.Println(err)
			return
		}

		e.Username = username
		s.aw.PushAnswer(&e)
	}
}
