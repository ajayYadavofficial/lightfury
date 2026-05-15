package repository

import "lightfury/internal/model"

type GameRepository interface {
	Save(game *model.Game) error
	Get(gameID int) (*model.Game, error)
	Delete(gameID int) error
	Exists(gameID int) bool
}
