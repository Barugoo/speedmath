package main

import (
	"context"
	"fmt"
	"time"
)

const minPlayersToStart = 2

func (g *Game) handleAnswer(e *EventAnswer) error {
	g.broadcast(fmt.Sprintf("Игрок %s прислал ответ: %d", e.Username, e.Answer))
	if e.Answer == g.curRound.curProblem.ExpectedAnswer {
		g.broadcast(fmt.Sprintf("Игрок %s победил!", e.Username))

		g.curRound.cancel()
		g.curRoundNum++
		g.nextRoundCh <- struct{}{}
		return nil
	}
	g.broadcast(fmt.Sprintf("Ответ игрока %s неверный:(", e.Username))
	return nil
}

func (g *Game) handleLogin(e *EventLogin) error {
	if _, ok := g.players[e.Username]; ok {
		return fmt.Errorf("player already exists")
	}
	g.players[e.Username] = e.OutputCh

	g.broadcast(fmt.Sprintf("Поприветствуйте %s!", e.Username))

	if len(g.players) >= minPlayersToStart {
		g.broadcast("Игра началась!")
		g.nextRoundCh <- struct{}{}
	}
	return nil
}

func (g *Game) handleNextRound() {
	g.broadcast("Генерируем вопрос...")
	g.pauseGame(time.Now().UTC().Add(time.Second))

	expr, expectedAnswer := generateProblem()

	ctx, cancel := context.WithCancel(context.Background())

	nextRound := Round{
		num: g.curRound.num + 1,
		curProblem: &Problem{
			Expression:     expr,
			ExpectedAnswer: expectedAnswer,
		},
		ctx:      ctx,
		cancel:   cancel,
		answerCh: make(chan *EventAnswer, 10),
	}

	g.curRound = &nextRound
	g.broadcastNextProblem(&nextRound)
	g.roundCh <- &nextRound
}

func (g *Game) handleLogout(e *EventLogout) error {
	if _, ok := g.players[e.Username]; !ok {
		return fmt.Errorf("player not found")
	}
	delete(g.players, e.Username)

	g.broadcast(fmt.Sprintf("Нас покинул %s:(", e.Username))

	if len(g.players) < minPlayersToStart {
		g.broadcast("Игра приостановлена из-за нехватки игроков!")
		g.stopGameCh <- struct{}{}
	}
	return nil
}
