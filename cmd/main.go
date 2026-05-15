package main

import (
	"log"
	"net/http"

	"lightfury/config"
	"lightfury/internal/handler"
	"lightfury/internal/repository"
	"lightfury/internal/service"
)

func main() {
	repo := repository.NewInMemoryGameRepository()
	svc := service.NewLoggingGameService(service.NewGameService(repo))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handler.HealthHandler)
	mux.HandleFunc("POST /api/v1/resume-game", handler.ResumeGameHandler(svc))
	mux.HandleFunc("DELETE /debug/game/{id}", handler.DeleteGameHandler(repo))
	mux.HandleFunc("GET /ws", func(w http.ResponseWriter, r *http.Request) {
		handler.ServeWS(svc, w, r)
	})

	log.Printf("Lightfury server listening on %s", config.ServerPort)
	if err := http.ListenAndServe(config.ServerPort, mux); err != nil {
		log.Fatal(err)
	}
}
