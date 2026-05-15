package repository

import (
	"lightfury/internal/model"
	"testing"
)

func newTestGame(id int) *model.Game {
	return &model.Game{ID: id, WinningScore: 50, State: model.StateWaiting}
}

func TestSaveAndGet(t *testing.T) {
	repo := NewInMemoryGameRepository()
	game := newTestGame(1)

	repo.Save(game)
	got, err := repo.Get(1)

	if err != nil || got.ID != 1 {
		t.Fatalf("expected game 1, got err=%v", err)
	}
}

func TestGetNotFound(t *testing.T) {
	repo := NewInMemoryGameRepository()
	_, err := repo.Get(99)
	if err == nil {
		t.Fatal("expected error for missing game")
	}
}

func TestExists(t *testing.T) {
	repo := NewInMemoryGameRepository()
	repo.Save(newTestGame(2))

	if !repo.Exists(2) {
		t.Fatal("expected game 2 to exist")
	}
	if repo.Exists(99) {
		t.Fatal("expected game 99 to not exist")
	}
}

func TestDelete(t *testing.T) {
	repo := NewInMemoryGameRepository()
	repo.Save(newTestGame(3))
	repo.Delete(3)

	if repo.Exists(3) {
		t.Fatal("expected game 3 to be deleted")
	}
}

func TestDeleteNotFound(t *testing.T) {
	repo := NewInMemoryGameRepository()
	err := repo.Delete(99)
	if err == nil {
		t.Fatal("expected error when deleting missing game")
	}
}
