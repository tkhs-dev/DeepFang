package env

import (
	"fmt"
	"math/rand"
)

type Setting struct {
	Width           int   `json:"width"`
	Height          int   `json:"height"`
	Seed            int64 `json:"seed"`
	FoodSpawnChance int   `json:"food_spawn_chance"`
	MinimumFood     int   `json:"minimum_food"`
}

type Game struct {
	Setting Setting
	Board   Board
	Rand    Rand
}

const (
	TurnResultMaskAlive  uint32 = 0b1
	TurnResultMaskResult uint32 = 0b11 << 1
	TurnResultMaskCause  uint32 = 0b111 << 3
	TurnResultMaskFood   uint32 = 0b1 << 6
)

const (
	TurnResultAlive          uint32 = 0
	TurnResultDead           uint32 = 1
	TurnResultContinue       uint32 = 0 << 1
	TurnResultWin            uint32 = 1 << 1
	TurnResultDraw           uint32 = 2 << 1
	TurnResultLoss           uint32 = 3 << 1
	TurnResultStarvation     uint32 = 1 << 3
	TurnResultSelfCollision  uint32 = 2 << 3
	TurnResultWallCollision  uint32 = 3 << 3
	TurnResultSnakeCollision uint32 = 4 << 3
	TurnResultHeadCollision  uint32 = 5 << 3
	TurnResultFoodEaten      uint32 = 1 << 6
)

type Direction int

const (
	Up Direction = iota
	Down
	Left
	Right
)

type GameContext struct {
	values map[string]interface{}
}

type Client struct {
	onStartCallback   func(ctx *GameContext)
	onMoveCallback    func(ctx *GameContext, board *Board, snake *Snake) int
	onTurnEndCallback func(ctx *GameContext, board *Board, snake *Snake, result uint32)
}

func NewGame(setting Setting) *Game {
	return &Game{
		Setting: setting,
		Board:   *NewBoard(setting.Width, setting.Height),
		Rand:    NewStandardRand(setting.Seed),
	}
}

func (g *Game) Execute(moves []Direction) ([]uint32, error) {
	if len(moves) != len(g.Board.Snakes) {
		return nil, fmt.Errorf("number of moves (%d) does not match number of snakes (%d)", len(moves), len(g.Board.Snakes))
	}

	result := make([]uint32, len(g.Board.Snakes))
	aliveSnakeCount := 0
	eliminatedSnakeCount := 0

	vBoard := g.Board.Copy() // Create a copy of the board to avoid modifying the original during the turn
	for i := range g.Board.Snakes {
		snake := &vBoard.Snakes[i]
		if !snake.Alive {
			continue // Skip dead snakes
		}
		aliveSnakeCount++

		move := moves[i]
		nextHead := snake.Head.getMovedPoint(move)
		println("Snake", i, "moving to (", nextHead.X, ",", nextHead.Y, ")")
		// Check if the snake eats food
		eaten := g.Board.isFood(nextHead)
		snake.move(nextHead, eaten)
		if eaten {
			result[i] |= TurnResultFoodEaten
			vBoard.Food.Remove(nextHead)
			println("Snake", i, "has eaten food at (", nextHead.X, ",", nextHead.Y, ")")
		} else if snake.Health <= 0 {
			result[i] |= TurnResultStarvation
			println("Snake", i, "has starved")
			snake.Alive = false // Snake dies from starvation
			eliminatedSnakeCount++
			continue
		}

		// Check for wall collisions
		if snake.Head.X < 0 || snake.Head.X >= g.Board.Width || snake.Head.Y < 0 || snake.Head.Y >= g.Board.Height {
			result[i] |= TurnResultWallCollision
			println("Collision detected with wall")
			snake.Alive = false // Snake dies on wall collision
			eliminatedSnakeCount++
		}
	}
	// Update the board
	g.Board = vBoard.Copy()

	// Check for collisions
	for i := range g.Board.Snakes {
		snake := &g.Board.Snakes[i]
		vsnake := &vBoard.Snakes[i]
		if !vsnake.Alive {
			continue // Skip dead snakes
		}

		if collidedSnake, selfCollision, headCollision := g.Board.getCollision(snake); collidedSnake != nil {
			if selfCollision {
				result[i] |= TurnResultSelfCollision
				println("Self-collision detected")
				vsnake.Alive = false
				eliminatedSnakeCount++
			} else if headCollision {
				if snake.Length <= collidedSnake.Length {
					result[i] |= TurnResultHeadCollision
					println(i, "Head collision detected with another snake's head")
					if snake.Length == collidedSnake.Length && vBoard.Snakes[collidedSnake.Id].Alive {
						result[collidedSnake.Id] |= TurnResultHeadCollision
						vBoard.Snakes[collidedSnake.Id].Alive = false
						eliminatedSnakeCount++
					}
					vsnake.Alive = false
					eliminatedSnakeCount++
				} else {
					println(i, "Head collision detected with another snake's head, but this snake is longer")
				}
			} else {
				result[i] |= TurnResultSnakeCollision
				println("Collision detected with another snake's body")
				vsnake.Alive = false
				eliminatedSnakeCount++
			}
		}
	}

	g.Board = vBoard // Update the original board with the modified state

	for i, snake := range g.Board.Snakes {
		if !snake.Alive {
			result[i] |= TurnResultDead
			println("Snake", i, "is dead")
			if aliveSnakeCount == eliminatedSnakeCount {
				result[i] |= TurnResultDraw
				println("All snakes are dead, the game is a draw")
			} else {
				result[i] |= TurnResultLoss
				println("Snake", i, "has lost the game")
			}
		} else {
			result[i] |= TurnResultAlive
			println("Snake", i, "is alive")
			if aliveSnakeCount-eliminatedSnakeCount == 1 {
				result[i] |= TurnResultWin
				println("Snake", i, "has won the game")
			} else {
				result[i] |= TurnResultContinue
				println("Snake", i, "continues to the next turn")
			}
		}
	}

	return result, nil
}

func foodDemand(numCurrentFood int, setting Setting) int {
	// Calculate the number of food items needed based on the game settings
	minFood := setting.MinimumFood
	foodSpawnChance := setting.FoodSpawnChance

	if numCurrentFood < minFood {
		return minFood - numCurrentFood
	}
	if foodSpawnChance > 0 && (100-rand.Intn(100)) < foodSpawnChance {
		return 1
	}

	return 0
}

//func placeFoodRandomly(n int) {
//	unoccupiedPoints := rules.GetUnoccupiedPoints(b, false, false)
//	g.placeFoodRandomlyAtPositions(rand, b, editor, n, unoccupiedPoints)
//}
//
//func placeFoodRandomlyAtPositions(rand rules.Rand, b *rules.BoardState, editor Editor, n int, positions []rules.Point) {
//	if len(positions) < n {
//		n = len(positions)
//	}
//
//	rand.Shuffle(len(positions), func(i int, j int) {
//		positions[i], positions[j] = positions[j], positions[i]
//	})
//
//	for i := 0; i < n; i++ {
//		editor.AddFood(positions[i])
//	}
//}
