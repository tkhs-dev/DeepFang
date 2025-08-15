package main

/*
#include <stdint.h>

*/
import "C"
import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/BattlesnakeOfficial/rules"
	"github.com/BattlesnakeOfficial/rules/client"
	"github.com/BattlesnakeOfficial/rules/maps"
)

func main() {
}

type Arg struct {
	Setting   *GameSetting `json:"setting"`
	ModelPath string       `json:"model_path"`
	Enemies   []string     `json:"enemies"`
}

type GameSetting struct {
	// ゲームの設定
	Seed            int64
	Width           int
	Height          int
	FoodSpawnChance int
	MinimumFood     int
}

type EpisodeContext struct {
	values map[string]any
}

type BoardState struct {
	*rules.BoardState
	You rules.Snake
}

type Client struct {
	callOnStart func(*EpisodeContext)
	callOnMove  func(*EpisodeContext, *BoardState) int
	callOnEnd   func(*EpisodeContext)
}

type SnakeState struct {
	ID      string
	Client  *Client
	Context *EpisodeContext
}

type GameState struct {
	// ゲームの状態
	gameId      string
	gameSetting *GameSetting
	gameMap     maps.GameMap
	ruleSet     rules.Ruleset
	settings    map[string]string
	snakeStates map[string]SnakeState
}

type StepData struct {
	State  [][][]float32 // CNN入力用盤面データなど
	Action int
	Reward float32
	Done   bool
	Value  float32 // Criticの推定値（Go側で推論するなら）
}

type EpisodeData struct {
	Steps []StepData
}

//export CollectExperience
func CollectExperience(carg *C.char) *C.char {
	arg := &Arg{}
	err := json.Unmarshal([]byte(C.GoString(carg)), arg)
	if err != nil {
		println(err.Error())
		return C.CString("")
	}
	res, err := startEpisode(arg)
	if err != nil {
		return C.CString("")
	}
	bytes, err := json.Marshal(res)
	if err != nil {
		return C.CString("")
	}
	return C.CString(string(bytes))
}

func (ctx *EpisodeContext) SetValue(key string, value any) {
	if ctx.values == nil {
		ctx.values = make(map[string]any)
	}
	ctx.values[key] = value
}
func (ctx *EpisodeContext) GetValue(key string) (any, bool) {
	if ctx.values == nil {
		return nil, false
	}
	value, ok := ctx.values[key]
	return value, ok
}

func startEpisode(arg *Arg) (*EpisodeData, error) {
	setting := arg.Setting
	gameState := &GameState{gameSetting: setting}
	gameState.gameId = "game-" + time.Now().Format("20060102150405")
	gameState.snakeStates = make(map[string]SnakeState)

	clients := make([]*Client, 0)
	for _, enemy := range arg.Enemies {
		if enemy == "" {
			clients = append(clients, &Client{
				callOnStart: func(ctx *EpisodeContext) {
				},
				callOnMove: func(ctx *EpisodeContext, board *BoardState) int {
					return rand.Intn(4) // 0: up, 1: down, 2: left, 3: right
				},
				callOnEnd: func(ctx *EpisodeContext) {
				},
			})
		}
	}

	for i, client := range clients {
		gameState.snakeStates["snake-"+fmt.Sprint(i)] = SnakeState{ID: "snake-" + fmt.Sprint(i), Client: client, Context: &EpisodeContext{}}
	}
	gameMap, _ := maps.GetMap("standard")
	gameState.gameMap = gameMap
	gameState.settings = map[string]string{
		rules.ParamFoodSpawnChance:     fmt.Sprint(setting.FoodSpawnChance),
		rules.ParamMinimumFood:         fmt.Sprint(setting.MinimumFood),
		rules.ParamHazardDamagePerTurn: fmt.Sprint(0),
		rules.ParamShrinkEveryNTurns:   fmt.Sprint(0),
	}
	ruleSet := rules.NewRulesetBuilder().WithSeed(setting.Seed).WithParams(gameState.settings).WithSolo(len(clients) == 1).NamedRuleset("standard")
	gameState.ruleSet = ruleSet

	rand.Seed(gameState.gameSetting.Seed)

	gameOver, boardState, err := initializeBoard(gameState)
	if err != nil {
		return nil, err
	}

	for !gameOver { //メインループ
		gameOver, boardState, err = gameState.createNextBoard(boardState)
		if err != nil {
			return nil, err
		}
	}

	for _, snakeState := range gameState.snakeStates {
		snakeState.Client.callOnEnd(snakeState.Context)
	}

	return &EpisodeData{
		Steps: []StepData{},
	}, nil

}

func initializeBoard(gameState *GameState) (bool, *rules.BoardState, error) {
	// 盤面を初期化する
	snakeIds := []string{}
	for _, snakeState := range gameState.snakeStates {
		snakeIds = append(snakeIds, snakeState.ID)
	}
	boardState, err := maps.SetupBoard(gameState.gameMap.ID(), gameState.ruleSet.Settings(), gameState.gameSetting.Width, gameState.gameSetting.Height, snakeIds)
	if err != nil {
		return false, nil, err
	}

	gameOver, boardState, err := gameState.ruleSet.Execute(boardState, nil)
	if err != nil {
		return false, nil, err
	}
	for _, snakeState := range gameState.snakeStates {
		snakeState.Client.callOnStart(snakeState.Context)
	}
	return gameOver, boardState, nil
}

func (gameState *GameState) createNextBoard(boardState *rules.BoardState) (bool, *rules.BoardState, error) {
	boardState, err := maps.PreUpdateBoard(gameState.gameMap, boardState, gameState.ruleSet.Settings())
	if err != nil {
		return false, nil, err
	}
	var moves []rules.SnakeMove
	for _, snakeState := range gameState.snakeStates {
		for _, snake := range boardState.Snakes {
			if snake.ID == snakeState.ID && snake.EliminatedCause == rules.NotEliminated {
				action := gameState.getSnakeAction(boardState, snakeState)
				moves = append(moves, action)
			}
		}
	}

	gameOver, boardState, err := gameState.ruleSet.Execute(boardState, moves)
	if err != nil {
		return false, nil, err
	}

	boardState, err = maps.PostUpdateBoard(gameState.gameMap, boardState, gameState.ruleSet.Settings())
	if err != nil {
		return false, nil, err
	}
	boardState.Turn++
	return gameOver, boardState, nil
}

func (gameState *GameState) getSnakeAction(boardState *rules.BoardState, snakeState SnakeState) rules.SnakeMove {
	you := rules.Snake{}
	for _, snake := range boardState.Snakes {
		if snake.ID == snakeState.ID {
			you = snake
			break
		}
	}
	actionNum := snakeState.Client.callOnMove(
		snakeState.Context,
		&BoardState{
			BoardState: boardState,
			You:        you,
		},
	)
	var action string
	switch actionNum {
	case 0:
		action = "up"
	case 1:
		action = "down"
	case 2:
		action = "left"
	case 3:
		action = "right"
	}
	return rules.SnakeMove{
		ID:   snakeState.ID,
		Move: action,
	}
}

func convertRulesSnake(snake rules.Snake, snakeState SnakeState) client.Snake {
	return client.Snake{
		ID:      snake.ID,
		Name:    snakeState.ID,
		Health:  snake.Health,
		Body:    client.CoordFromPointArray(snake.Body),
		Latency: "0",
		Head:    client.CoordFromPoint(snake.Body[0]),
		Length:  len(snake.Body),
		Shout:   "",
		Customizations: client.Customizations{
			Head:  "",
			Tail:  "",
			Color: "",
		},
	}
}

func convertRulesSnakes(snakes []rules.Snake, snakeStates map[string]SnakeState, includeEliminated bool) []client.Snake {
	a := make([]client.Snake, 0)
	for _, snake := range snakes {
		if snake.EliminatedCause == rules.NotEliminated || includeEliminated {
			a = append(a, convertRulesSnake(snake, snakeStates[snake.ID]))
		}
	}
	return a
}

func convertStateToBoard(boardState *rules.BoardState, snakeStates map[string]SnakeState) client.Board {
	return client.Board{
		Height:  boardState.Height,
		Width:   boardState.Width,
		Food:    client.CoordFromPointArray(boardState.Food),
		Hazards: client.CoordFromPointArray(boardState.Hazards),
		Snakes:  convertRulesSnakes(boardState.Snakes, snakeStates, false),
	}
}
