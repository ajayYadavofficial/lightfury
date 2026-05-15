package repository

import (
	"fmt"
	"lightfury/internal/model"
	"sync"
)

type InMemoryGameRepository struct {
	mu    sync.RWMutex
	games map[int]*model.Game
}

func NewInMemoryGameRepository() *InMemoryGameRepository {
	return &InMemoryGameRepository{
		games: make(map[int]*model.Game),
	}
}

func (r *InMemoryGameRepository) Save(game *model.Game) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.games[game.ID] = game
	return nil
}

func (r *InMemoryGameRepository) Get(gameID int) (*model.Game, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	game, ok := r.games[gameID]
	if !ok {
		return nil, fmt.Errorf("game %d not found", gameID)
	}
	return game, nil
}

func (r *InMemoryGameRepository) Delete(gameID int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.games[gameID]; !ok {
		return fmt.Errorf("game %d not found", gameID)
	}
	delete(r.games, gameID)
	return nil
}

func (r *InMemoryGameRepository) Exists(gameID int) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.games[gameID]
	return ok
}
