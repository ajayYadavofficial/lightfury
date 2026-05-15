# Lightfury — Game Server Backend

A Go-based game server backend for a turn-based 2-player dice game with crash-recovery semantics. Supports real-time WebSocket communication, game state persistence, and mid-game resume via a REST API.

---

## Tech Stack

| Layer | Choice |
|---|---|
| Language | Go |
| Transport | WebSocket (`gorilla/websocket`) |
| State | In-memory repository (`sync.RWMutex` map) |
| Logging | Decorator pattern wrapping `GameService` |

---

## Architecture

```
cmd/main.go                           — entry point, wires all dependencies
config/config.go                      — constants (MaxPlayers, WinningScore, Port)
internal/
  model/game.go                       — Player, Game structs, GameState enum
  repository/game_repository.go       — GameRepository interface
  repository/in_memory_repository.go  — thread-safe in-memory implementation
  service/game_service.go             — GameService interface
  service/game_service_impl.go        — business logic
  service/logging_service.go          — logging decorator (wraps GameService)
  handler/ws_handler.go               — WebSocket upgrade and packet routing
  handler/http_handler.go             — REST endpoints
  packet/types.go                     — inbound/outbound JSON packet structs
```

### Design Patterns Used
- **Repository Pattern** — data access behind a `GameRepository` interface, decoupled from business logic
- **Decorator Pattern** — `LoggingGameService` wraps `GameService` and adds structured logging to every method without touching business logic
- **Interface-driven design** — both `GameService` and `GameRepository` are interfaces; concrete implementations are injected at startup

---

## Game Rules

- Strictly 2 players per game (configurable via `config.MaxPlayers`)
- Turn-based: players alternate rolling a dice (1–6)
- Scores are cumulative — first player to reach or pass **50 points** wins
- If a player leaves mid-game, the other player is declared the winner

---

## API Reference

### WebSocket

**Connect as a player**
```
ws://localhost:8080/ws?name=<name>&id=<player_id>
```
Connecting automatically joins the waiting queue. The game starts once 2 players are connected.

**Inbound packets (client → server)**

Roll dice:
```json
{ "type": "roll_dice", "game_id": 1 }
```

Leave game:
```json
{ "type": "leave_game", "game_id": 1 }
```

**Outbound packets (server → client)**

Game started:
```json
{
  "type": "game_started",
  "game_id": 1,
  "players": [
    { "id": 1, "name": "Alice", "score": 0 },
    { "id": 2, "name": "Bob",   "score": 0 }
  ],
  "current_turn": 0
}
```

Turn update:
```json
{
  "type": "turn_update",
  "game_id": 1,
  "rolled": 4,
  "players": [
    { "id": 1, "name": "Alice", "score": 4 },
    { "id": 2, "name": "Bob",   "score": 0 }
  ],
  "current_turn": 1
}
```

Game ended:
```json
{ "type": "game_ended", "game_id": 1, "winner": "Alice" }
```

Game resumed:
```json
{
  "type": "game_resumed",
  "game_id": 1,
  "players": [
    { "id": 1, "name": "Alice", "score": 20 },
    { "id": 2, "name": "Bob",   "score": 15 }
  ],
  "current_turn": 0
}
```

---

### REST Endpoints

#### Health Check
```
GET /health
```
```bash
curl http://localhost:8080/health
```
Response:
```json
{ "status": "ok" }
```

---

#### Resume Game
```
POST /api/v1/resume-game
```
Restores a game from a snapshot. Used to recover games after a server crash — the caller provides the last known game state and the server resumes from that point, notifying both connected players.

Returns `409 Conflict` if the game is already running on this instance (duplicate resume guard).

```bash
curl -X POST http://localhost:8080/api/v1/resume-game \
  -H "Content-Type: application/json" \
  -d '{
    "game_id": 1,
    "snapshot": {
      "id": 1,
      "players": [
        { "id": 1, "name": "Alice", "score": 20 },
        { "id": 2, "name": "Bob",   "score": 15 }
      ],
      "current_turn": 0,
      "is_ended": false,
      "winning_score": 50,
      "state": "RUNNING"
    }
  }'
```

Response:
```json
{ "status": "resumed" }
```

---

## Running the Server

```bash
go run cmd/main.go
```

Server starts on `:8080`.

---

## Test Plan

### Prerequisites
- Server running: `go run cmd/main.go`
- Postman (or any WebSocket client) for WS connections
- `curl` for REST calls

---

### Test 1 — Full Game (Normal Flow)

**Step 1:** Open two WebSocket connections in Postman:
```
ws://localhost:8080/ws?name=Alice&id=1
ws://localhost:8080/ws?name=Bob&id=2
```

**Step 2:** Both connections receive `game_started`. Note the `game_id` and each player's `id`.

**Step 3:** Alice sends `roll_dice` (it is `current_turn: 0`, Alice's turn):
```json
{ "type": "roll_dice", "game_id": 1 }
```
Both players receive `turn_update` with updated scores and `current_turn: 1`.

**Step 4:** Bob sends `roll_dice`:
```json
{ "type": "roll_dice", "game_id": 1 }
```
Both players receive `turn_update` with `current_turn: 0`.

**Step 5:** Alternate turns until a player's cumulative score reaches 50. Both players receive:
```json
{ "type": "game_ended", "game_id": 1, "winner": "Alice" }
```

---

### Test 2 — Player Leaves Mid-Game

**Step 1:** Connect both players (same as Test 1 Step 1). Wait for `game_started`.

**Step 2:** Alice sends a few `roll_dice` packets, then sends:
```json
{ "type": "leave_game", "game_id": 1 }
```

**Expected:** Bob receives:
```json
{ "type": "game_ended", "game_id": 1, "winner": "Bob" }
```

**Step 3:** Alice reconnects with the same id — should succeed (connection was cleaned up):
```
ws://localhost:8080/ws?name=Alice&id=1
```

---

### Test 3 — Duplicate Player ID Rejected

**Step 1:** Connect Alice:
```
ws://localhost:8080/ws?name=Alice&id=1
```

**Step 2:** Open a second connection with the same id:
```
ws://localhost:8080/ws?name=Alice&id=1
```

**Expected:** Second connection receives and closes:
```json
{ "type": "error", "code": "JOIN_FAILED", "message": "player id 1 already connected" }
```

---

### Test 4 — Resume Game (Crash Recovery)

**Step 1:** Connect both players and start a game:
```
ws://localhost:8080/ws?name=Alice&id=1
ws://localhost:8080/ws?name=Bob&id=2
```

**Step 2:** Play a few turns so scores are non-zero. Note the current scores and whose turn it is.

**Step 3:** Simulate a crash — evict the game from the server:
```bash
curl -X DELETE http://localhost:8080/debug/game/1
```
Response: `{"status":"deleted"}`

**Step 4:** Resume the game with the last known state (use the scores from Step 2):
```bash
curl -X POST http://localhost:8080/api/v1/resume-game \
  -H "Content-Type: application/json" \
  -d '{
    "game_id": 1,
    "snapshot": {
      "id": 1,
      "players": [
        { "id": 1, "name": "Alice", "score": 20 },
        { "id": 2, "name": "Bob",   "score": 15 }
      ],
      "current_turn": 0,
      "is_ended": false,
      "winning_score": 50,
      "state": "RUNNING"
    }
  }'
```

**Expected:** Both WebSocket connections receive:
```json
{
  "type": "game_resumed",
  "game_id": 1,
  "players": [
    { "id": 1, "name": "Alice", "score": 20 },
    { "id": 2, "name": "Bob",   "score": 15 }
  ],
  "current_turn": 0
}
```

**Step 5:** Continue rolling dice — game proceeds normally from the resumed state.

**Step 6:** Call resume again with the same `game_id` — should be rejected:
```bash
curl -X POST http://localhost:8080/api/v1/resume-game \
  -H "Content-Type: application/json" \
  -d '{
    "game_id": 1,
    "snapshot": {
      "id": 1,
      "players": [],
      "current_turn": 0,
      "is_ended": false,
      "winning_score": 50,
      "state": "RUNNING"
    }
  }'
```
**Expected:** `409 Conflict` — `{"error":"game 1 already running on this instance"}`

---

### Test 5 — Health Check

```bash
curl http://localhost:8080/health
```
**Expected:** `{"status":"ok"}`

---

## Out of Scope

- WebSocket disconnect/reconnect handling
- Persistent storage (Redis, Postgres)
- Authentication
- Graceful shutdown
