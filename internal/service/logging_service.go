package service

import (
	"log"
	"lightfury/internal/model"
	"time"

	"github.com/gorilla/websocket"
)

type LoggingGameService struct {
	inner GameService
}

func NewLoggingGameService(inner GameService) *LoggingGameService {
	return &LoggingGameService{inner: inner}
}

func (l *LoggingGameService) PlayerJoin(playerID int, playerName string, conn *websocket.Conn) error {
	start := time.Now()
	err := l.inner.PlayerJoin(playerID, playerName, conn)
	log.Printf("[PlayerJoin] playerID=%d player=%s err=%v duration=%s", playerID, playerName, err, time.Since(start))
	return err
}

func (l *LoggingGameService) RollDice(gameID int, playerID int) error {
	start := time.Now()
	err := l.inner.RollDice(gameID, playerID)
	log.Printf("[RollDice] gameID=%d playerID=%d err=%v duration=%s", gameID, playerID, err, time.Since(start))
	return err
}

func (l *LoggingGameService) ResumeGame(gameID int, snapshot model.Game) error {
	start := time.Now()
	err := l.inner.ResumeGame(gameID, snapshot)
	log.Printf("[ResumeGame] gameID=%d err=%v duration=%s", gameID, err, time.Since(start))
	return err
}

func (l *LoggingGameService) LeaveGame(gameID int, playerID int) error {
	start := time.Now()
	err := l.inner.LeaveGame(gameID, playerID)
	log.Printf("[LeaveGame] gameID=%d playerID=%d err=%v duration=%s", gameID, playerID, err, time.Since(start))
	return err
}
