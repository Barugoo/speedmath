package main

import (
	"context"
	"log"
	"time"
)

type GameStatus string

const (
	GameStatusWaitingForPlayers GameStatus = "waiting_for_players"
	GameStatusStarted           GameStatus = "started"
	GameStatusSendingProblem    GameStatus = "sending_problem"
	GameStatusWaitingForAnswer  GameStatus = "waiting_for_answer"
	GameStatusWinnerPicked      GameStatus = "winner_picked"
)

type Problem struct {
	Expression     string
	ExpectedAnswer int64
}

type Round struct {
	num        int64
	curProblem *Problem
	ctx        context.Context
	cancel     context.CancelFunc
	answerCh   chan *EventAnswer
}

type Game struct {
	status      GameStatus
	curRound    *Round
	curRoundNum int64

	stopGameCh  chan struct{}
	nextRoundCh chan struct{}
	roundCh     chan *Round
	loginCh     chan *EventLogin
	logoutCh    chan *EventLogout

	players map[string]chan<- any
}

func (g *Game) broadcast(msg string) {
	for _, player := range g.players {
		player <- struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		}{
			Type:    "print_text",
			Message: msg,
		}
	}
}

func (g *Game) broadcastNextProblem(r *Round) {
	for _, player := range g.players {
		player <- struct {
			Type     string `json:"type"`
			Message  string `json:"message"`
			RoundNum int64  `json:"round_num"`
		}{
			Type:     "next_problem",
			RoundNum: r.num,
			Message:  "Решите пример: " + r.curProblem.Expression,
		}
	}
}

func (g *Game) pauseGame(until time.Time) {
	for _, player := range g.players {
		player <- struct {
			Type       string    `json:"type"`
			PauseUntil time.Time `json:"pause_until"`
		}{
			Type:       "pause_until",
			PauseUntil: until,
		}
	}
}

func (g *Game) Run() {
	log.Println("running game engine..")
	g.status = GameStatusWaitingForPlayers
	g.curRound = &Round{}
	for {
		var err error
		select {
		case e := <-g.loginCh:
			err = g.handleLogin(e)
		case e := <-g.logoutCh:
			err = g.handleLogout(e)
		case e := <-g.curRound.answerCh:
			err = g.handleAnswer(e)
		case <-g.nextRoundCh:
			g.handleNextRound()
		}
		if err != nil {
			log.Println(err)
		}
	}
}

func NewGame() *Game {
	return &Game{
		players:     make(map[string]chan<- any),
		nextRoundCh: make(chan struct{}, 10),
		roundCh:     make(chan *Round),
		loginCh:     make(chan *EventLogin, 10),
		logoutCh:    make(chan *EventLogout, 10),
		stopGameCh:  make(chan struct{}, 10),
	}
}
