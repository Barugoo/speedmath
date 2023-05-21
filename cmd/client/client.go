package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

type EventType string

const (
	EventTypePrintText   EventType = "print_text"
	EventTypeNextProblem EventType = "next_problem"
	EventTypePauseUntil  EventType = "pause_until"
)

type ServerEvent struct {
	Type       EventType `json:"type"`
	Message    string    `json:"message"`
	RoundNum   int64     `json:"round_num"`
	PauseUntil time.Time `json:"pause_until"`
}

func handleConn(conn *websocket.Conn) <-chan ServerEvent {
	ch := make(chan ServerEvent, 10)
	go func() {
		for {
			var m ServerEvent
			if err := conn.ReadJSON(&m); err != nil {
				panic(err)
			}
			ch <- m
		}
	}()
	return ch
}

type InputEvent struct {
	RoundNum  int64 `json:"round_num"`
	Timestamp int64 `json:"timestamp"`
	Answer    int64 `json:"answer"`
}

func handleInput(r io.Reader) <-chan InputEvent {
	ch := make(chan InputEvent, 10)
	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			if err := scanner.Err(); err != nil {
				panic(err)
			}

			t := time.Now().Unix()
			ans, err := strconv.ParseInt(scanner.Text(), 10, 64)
			if err != nil {
				log.Println("incorrect input!")
				continue
			}
			ch <- InputEvent{
				Answer:    ans,
				Timestamp: t,
			}
		}
	}()
	return ch
}

func main() {
	var username string
	flag.StringVar(&username, "u", "", "username")
	flag.Parse()

	if username == "" {
		log.Fatal("please provide username -u")
	}

	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws", RawQuery: "username=" + username}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	answerCh := handleInput(os.Stdin)
	outputCh := handleConn(c)

	var currentProblemNum int64
	for {
		select {
		case a := <-answerCh:
			a.RoundNum = currentProblemNum
			if err := c.WriteJSON(a); err != nil {
				log.Fatal("unable to send answer to server", err)
			}
		case o := <-outputCh:
			switch o.Type {
			case EventTypePrintText:
				log.Println("server:", o.Message)
			case EventTypeNextProblem:
				currentProblemNum = o.RoundNum
				log.Println("next problem:", o.Message)
			case EventTypePauseUntil:
				time.Sleep(time.Until(o.PauseUntil))
			}
		}
	}
}
