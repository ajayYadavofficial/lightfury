package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"lightfury/internal/packet"
	"lightfury/internal/service"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func ServeWS(svc service.GameService, w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	playerName := q.Get("name")
	playerIDStr := q.Get("id")

	if playerName == "" || playerIDStr == "" {
		http.Error(w, "missing name or id query param", http.StatusBadRequest)
		return
	}
	playerID, err := strconv.Atoi(playerIDStr)
	if err != nil {
		http.Error(w, "id must be an integer", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
		return
	}
	defer conn.Close()

	if err := svc.PlayerJoin(playerID, playerName, conn); err != nil {
		conn.WriteJSON(packet.ErrorPacket{Type: packet.TypeError, Code: "JOIN_FAILED", Message: err.Error()})
		return
	}

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("ws read error player=%s: %v", playerName, err)
			break
		}

		var base packet.InboundBase
		if err := json.Unmarshal(msg, &base); err != nil {
			conn.WriteJSON(packet.ErrorPacket{Type: packet.TypeError, Code: "BAD_PACKET", Message: "invalid JSON"})
			continue
		}

		switch base.Type {
		case packet.TypeRollDice:
			var pkt packet.RollDice
			json.Unmarshal(msg, &pkt)
			if err := svc.RollDice(pkt.GameID, playerID); err != nil {
				conn.WriteJSON(packet.ErrorPacket{Type: packet.TypeError, Code: "ROLL_FAILED", Message: err.Error()})
			}

		case packet.TypeLeaveGame:
			var pkt packet.LeaveGame
			json.Unmarshal(msg, &pkt)
			svc.LeaveGame(pkt.GameID, playerID)
			return

		default:
			conn.WriteJSON(packet.ErrorPacket{Type: packet.TypeError, Code: "UNKNOWN_TYPE", Message: "unknown packet type"})
		}
	}
}
