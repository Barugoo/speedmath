package main

import (
	"log"
	"sort"
	"sync"
	"time"
)

type AnswerWorker struct {
	curRoundNum int64
	mu          *sync.Mutex
	buf         []*EventAnswer
	roundCh     chan *Round
}

func (b *AnswerWorker) handle() {
	go func() {
		for {
			nextRound, ok := <-b.roundCh
			log.Println(nextRound, ok)
			if !ok {
				return
			}
			b.handleRound(nextRound)
			b.buf = b.buf[:0]
		}
	}()
}

func (b *AnswerWorker) handleRound(r *Round) {
	b.curRoundNum = r.num

	ticker := time.NewTicker(time.Second * 1)
	for {
		select {
		case <-ticker.C:
			log.Println("tick")
			b.mu.Lock()
			sort.Slice(b.buf, func(i, j int) bool {
				return b.buf[i].Timestamp < b.buf[j].Timestamp
			})
			for _, e := range b.buf {
				select {
				case r.answerCh <- e:
				case <-r.ctx.Done():
					return
				}
			}
			b.buf = b.buf[:0]
			b.mu.Unlock()
		case <-r.ctx.Done():
			return
		}
	}
}

func (b *AnswerWorker) PushAnswer(e *EventAnswer) {
	if b.curRoundNum != e.RoundNum {
		return
	}
	b.mu.Lock()
	b.buf = append(b.buf, e)
	b.mu.Unlock()
}

func NewAnswerWorker(roundCh chan *Round) *AnswerWorker {
	aw := &AnswerWorker{
		mu:      &sync.Mutex{},
		buf:     make([]*EventAnswer, 0, 10),
		roundCh: roundCh,
	}
	go aw.handle()
	return aw
}
