package service

import (
	"github.com/gorilla/websocket"
	"lightfury/internal/model"
)

type GameService interface {
	PlayerJoin(playerID int, playerName string, conn *websocket.Conn) error
	RollDice(gameID int, playerID int) error
	ResumeGame(gameID int, snapshot model.Game) error
	LeaveGame(gameID int, playerID int) error
}
