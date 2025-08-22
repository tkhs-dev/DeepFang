package env

import (
	"golib/rule"

	mapset "github.com/deckarep/golang-set/v2"
)

type GameContext struct {
	values map[string]any
}

type Client interface {
	OnStartCallback(ctx *GameContext)
	OnMoveCallback(ctx *GameContext, board *rule.Board, snake *rule.Snake) rule.Direction
	OnTurnEndCallback(ctx *GameContext, board *rule.Board, snake *rule.Snake, result uint32)
}

func (ctx *GameContext) SetValue(key string, value any) {
	if ctx.values == nil {
		ctx.values = make(map[string]any)
	}
	ctx.values[key] = value
}
func (ctx *GameContext) GetValue(key string) (any, bool) {
	if ctx.values == nil {
		return nil, false
	}
	value, ok := ctx.values[key]
	return value, ok
}

type SimpleSafeClient struct {
}

func NewSimpleSafeClient() *SimpleSafeClient {
	return &SimpleSafeClient{}
}

func (s SimpleSafeClient) OnStartCallback(ctx *GameContext) {
}

func (s SimpleSafeClient) OnMoveCallback(ctx *GameContext, board *rule.Board, snake *rule.Snake) rule.Direction {
	actions := mapset.NewSet(rule.Up, rule.Down, rule.Left, rule.Right)
	safeActions := mapset.NewSet[rule.Direction]()
	for action := range actions.Iter() {
		if IsMoveSafe(board, snake.Id, action) {
			safeActions.Add(action)
		}
	}
	action, set := safeActions.Pop()
	if !set {
		println("No valid actions available, returning Right by default")
		return rule.Up // Default action if no valid actions are available
	}
	println("Selected action:", action.String())
	return action
}

func (s SimpleSafeClient) OnTurnEndCallback(ctx *GameContext, board *rule.Board, snake *rule.Snake, result uint32) {
}
