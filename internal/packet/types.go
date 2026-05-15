package packet

const (
	TypeGameStarted = "game_started"
	TypeTurnUpdate  = "turn_update"
	TypeGameEnded   = "game_ended"
	TypeError       = "error"
	TypeGameResumed = "game_resumed"

	// inbound
	TypeRollDice  = "roll_dice"
	TypeLeaveGame = "leave_game"
)

// Inbound — base struct to read the type field first
type InboundBase struct {
	Type string `json:"type"`
}

type RollDice struct {
	Type   string `json:"type"`
	GameID int    `json:"game_id"`
}

type LeaveGame struct {
	Type   string `json:"type"`
	GameID int    `json:"game_id"`
}

type PlayerInfo struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Score int    `json:"score"`
}

type GameStarted struct {
	Type        string       `json:"type"`
	GameID      int          `json:"game_id"`
	Players     []PlayerInfo `json:"players"`
	CurrentTurn int          `json:"current_turn"`
}

type TurnUpdate struct {
	Type        string       `json:"type"`
	GameID      int          `json:"game_id"`
	Rolled      int          `json:"rolled"`
	Players     []PlayerInfo `json:"players"`
	CurrentTurn int          `json:"current_turn"`
}

type GameEnded struct {
	Type   string `json:"type"`
	GameID int    `json:"game_id"`
	Winner string `json:"winner"`
}

type GameResumed struct {
	Type        string       `json:"type"`
	GameID      int          `json:"game_id"`
	Players     []PlayerInfo `json:"players"`
	CurrentTurn int          `json:"current_turn"`
}

type ErrorPacket struct {
	Type    string `json:"type"`
	Code    string `json:"code"`
	Message string `json:"message"`
}
