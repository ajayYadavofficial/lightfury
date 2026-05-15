package service

import (
	"fmt"
	"lightfury/config"
	"lightfury/internal/model"
	"lightfury/internal/packet"
	"lightfury/internal/repository"
	"math/rand"
	"sync"

	"github.com/gorilla/websocket"
)

type GameServiceImpl struct {
	repo         repository.GameRepository
	mu           sync.Mutex
	waitingQueue []*model.Player
	nextGameID   int
	activeConns  map[int]*websocket.Conn
}

func NewGameService(repo repository.GameRepository) *GameServiceImpl {
	return &GameServiceImpl{repo: repo, nextGameID: 1, activeConns: make(map[int]*websocket.Conn)}
}

func (s *GameServiceImpl) PlayerJoin(playerID int, playerName string, conn *websocket.Conn) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.activeConns[playerID]; exists {
		return fmt.Errorf("player id %d already connected", playerID)
	}

	s.activeConns[playerID] = conn
	player := &model.Player{ID: playerID, Name: playerName, Conn: conn}
	s.waitingQueue = append(s.waitingQueue, player)

	if len(s.waitingQueue) >= config.MaxPlayers {
		players := s.waitingQueue[:config.MaxPlayers]
		s.waitingQueue = s.waitingQueue[config.MaxPlayers:]
		s.startGame(players)
	}
	return nil
}

func (s *GameServiceImpl) startGame(players []*model.Player) {
	game := &model.Game{
		ID:           s.nextGameID,
		Players:      players,
		CurrentTurn:  0,
		WinningScore: config.WinningScore,
		State:        model.StateRunning,
	}
	s.nextGameID++
	s.repo.Save(game)

	pkt := packet.GameStarted{
		Type:        packet.TypeGameStarted,
		GameID:      game.ID,
		Players:     toPlayerInfos(game.Players),
		CurrentTurn: game.CurrentTurn,
	}
	broadcast(game.Players, pkt)
}

func (s *GameServiceImpl) RollDice(gameID int, playerID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	game, err := s.repo.Get(gameID)
	if err != nil {
		return err
	}
	if game.State != model.StateRunning {
		return fmt.Errorf("game %d is not running", gameID)
	}

	current := game.Players[game.CurrentTurn]
	if current.ID != playerID {
		return fmt.Errorf("not player %d's turn", playerID)
	}

	rolled := rand.Intn(6) + 1
	current.Score += rolled

	if current.Score >= game.WinningScore {
		game.State = model.StateEnded
		game.IsEnded = true
		game.Winner = current
		s.repo.Save(game)
		broadcast(game.Players, packet.GameEnded{
			Type:   packet.TypeGameEnded,
			GameID: game.ID,
			Winner: current.Name,
		})
		return nil
	}

	game.CurrentTurn = (game.CurrentTurn + 1) % len(game.Players)
	s.repo.Save(game)

	broadcast(game.Players, packet.TurnUpdate{
		Type:        packet.TypeTurnUpdate,
		GameID:      game.ID,
		Rolled:      rolled,
		Players:     toPlayerInfos(game.Players),
		CurrentTurn: game.CurrentTurn,
	})
	return nil
}

func (s *GameServiceImpl) LeaveGame(gameID int, playerID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	game, err := s.repo.Get(gameID)
	if err != nil {
		return err
	}

	if game.State == model.StateRunning {
		var winner *model.Player
		for _, p := range game.Players {
			if p.ID != playerID {
				winner = p
				break
			}
		}
		game.State = model.StateEnded
		game.IsEnded = true
		game.Winner = winner
		s.repo.Save(game)
		broadcast(game.Players, packet.GameEnded{
			Type:   packet.TypeGameEnded,
			GameID: game.ID,
			Winner: winner.Name,
		})
	}

	delete(s.activeConns, playerID)
	return s.repo.Delete(gameID)
}

func (s *GameServiceImpl) ResumeGame(gameID int, snapshot model.Game) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.repo.Exists(gameID) {
		return fmt.Errorf("game %d already running on this instance", gameID)
	}

	snapshot.ID = gameID
	for _, p := range snapshot.Players {
		p.Conn = s.activeConns[p.ID]
	}
	s.repo.Save(&snapshot)

	broadcast(snapshot.Players, packet.GameResumed{
		Type:        packet.TypeGameResumed,
		GameID:      gameID,
		Players:     toPlayerInfos(snapshot.Players),
		CurrentTurn: snapshot.CurrentTurn,
	})
	return nil
}

func broadcast(players []*model.Player, pkt any) {
	for _, p := range players {
		if p.Conn != nil {
			p.Conn.WriteJSON(pkt)
		}
	}
}

func toPlayerInfos(players []*model.Player) []packet.PlayerInfo {
	infos := make([]packet.PlayerInfo, len(players))
	for i, p := range players {
		infos[i] = packet.PlayerInfo{ID: p.ID, Name: p.Name, Score: p.Score}
	}
	return infos
}
