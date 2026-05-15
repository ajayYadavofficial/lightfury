package model

import "github.com/gorilla/websocket"

type GameState string

const (
	StateWaiting GameState = "WAITING"
	StateRunning GameState = "RUNNING"
	StateEnded   GameState = "ENDED"
)

type Player struct {
	ID    int             `json:"id"    bson:"id"`
	Name  string          `json:"name"  bson:"name"`
	Score int             `json:"score" bson:"score"`
	Conn  *websocket.Conn `json:"-"     bson:"-"`
}

type Game struct {
	ID           int       `json:"id"            bson:"id"`
	Players      []*Player `json:"players"       bson:"players"`
	CurrentTurn  int       `json:"current_turn"  bson:"current_turn"`
	IsEnded      bool      `json:"is_ended"      bson:"is_ended"`
	WinningScore int       `json:"winning_score" bson:"winning_score"`
	Winner       *Player   `json:"winner"        bson:"winner"`
	State        GameState `json:"state"         bson:"state"`
}
