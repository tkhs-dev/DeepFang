package env

import (
	"fmt"
	env "golib/rule"
)

type Experience struct {
	Board  *env.Board
	Actor  []float32
	Critic float32
	Action int
	Reward float32
	Done   bool
}

type EpisodeData struct {
	Steps []Experience
}

type EpisodeArg struct {
	Setting *env.Setting `json:"setting"`
}

func StartEpisode(arg *EpisodeArg, clients []Client) error {
	setting := arg.Setting
	game := env.NewGame(*setting)

	snakeSettings := make([]env.SnakeSetting, len(clients))
	for i := range clients {
		snakeSettings[i] = env.SnakeSetting{
			InitialLength: 3,
			InitialHealth: 100,
		}
	}
	err := game.Setup(snakeSettings)
	if err != nil {
		return fmt.Errorf("failed to setup game: %w", err)
	}

	ctxs := make([]*GameContext, len(clients))
	for i, c := range clients {
		ctxs[i] = &GameContext{}
		c.OnStartCallback(ctxs[i])
	}
	for {
		actions := make([]env.Direction, len(clients))
		for i, c := range clients {
			board := game.Board.Copy()
			action := c.OnMoveCallback(ctxs[i], &board, &board.Snakes[i])
			actions[i] = action
		}

		results, ended, err := game.Execute(actions)
		if err != nil {
			return err
		}
		for i, c := range clients {
			board := game.Board.Copy()
			snake := &board.Snakes[i]
			c.OnTurnEndCallback(ctxs[i], &board, snake, results[i])
		}
		if ended {
			break
		}
	}
	return nil
}
