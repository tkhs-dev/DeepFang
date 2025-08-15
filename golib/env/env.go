package env

import (
	"math/rand"

	mapset "github.com/deckarep/golang-set/v2"
)

type Layer [][]float32
type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Sneak struct {
	Body      []Point
	Head      Point
	Length    int
	Health    int
	Alive     bool
	BodyLayer Layer
	HeadLayer Layer
}

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
	onMoveCallback    func(ctx *GameContext, board *Board, snake *Sneak) int
	onTurnEndCallback func(ctx *GameContext, board *Board, snake *Sneak, result uint32)
}

func NewGame(setting Setting) *Game {
	gameMap := Board{
		Width:     setting.Width,
		Height:    setting.Height,
		Food:      mapset.NewSet[Point](),
		Snakes:    make([]Sneak, 0),
		FoodLayer: newLayer(setting.Width, setting.Height),
	}
	return &Game{
		Setting: setting,
		Board:   gameMap,
		Rand:    NewStandardRand(setting.Seed),
	}
}

func (g *Game) Execute(moves []Direction) []uint32 {
	result := make([]uint32, len(g.Board.Snakes))
	eatenFood := mapset.NewSet[Point]()
	aliveSnakeCount := 0
	eliminatedSnakeCount := 0
	for i, snake := range g.Board.Snakes {
		if !snake.Alive {
			continue // Skip dead snakes
		}
		aliveSnakeCount++
		if i < len(moves) {
			move := moves[i]
			nextHead := snake.Head
			nextHead.move(move)
			println("Snake", i, "moving to (", nextHead.X, ",", nextHead.Y, ")")
			// Check if the snake eats food
			eaten := g.Board.FoodLayer.getValue(nextHead.X, nextHead.Y) != 0
			snake.move(nextHead, eaten)
			if eaten {
				eatenFood.Add(nextHead) // Add the food to the eaten set
				result[i] |= TurnResultFoodEaten
			}
			g.Board.Snakes[i] = snake // Update the snake in the map
		}
	}
	// Remove eaten food from the board
	g.Board.Food = g.Board.Food.Difference(eatenFood)

	// Update the food layer
	for food := range eatenFood.Iter() {
		g.Board.FoodLayer.setValue(food.X, food.Y, 0.0)
	}
	deadByCollision := make([]bool, len(g.Board.Snakes))
	drawPossible := make([]bool, len(g.Board.Snakes))
	// Check for collisions
	for i, snake := range g.Board.Snakes {
		if !snake.Alive {
			continue // Skip dead snakes
		}
		// Check for starvation
		if snake.Health <= 0 {
			result[i] |= TurnResultStarvation
			println("Snake has starved")
			snake.Alive = false // Snake dies from starvation
			eliminatedSnakeCount++
			g.Board.Snakes[i] = snake // Update the snake in the map
			drawPossible[i] = true    // Mark as draw possible
			continue
		}
		// Check for wall collisions
		if snake.Head.X < 0 || snake.Head.X >= g.Board.Width || snake.Head.Y < 0 || snake.Head.Y >= g.Board.Height {
			result[i] |= TurnResultWallCollision
			println("Collision detected with wall")
			snake.Alive = false // Snake dies on wall collision
			eliminatedSnakeCount++
			g.Board.Snakes[i] = snake // Update the snake in the map
			drawPossible[i] = true
			continue
		}
		// Check for body collisions
		for j, otherSnake := range g.Board.Snakes {
			if !deadByCollision[i] && otherSnake.Alive {
				if otherSnake.BodyLayer.getValue(snake.Head.X, snake.Head.Y) != 0 {
					if i == j {
						result[i] |= TurnResultSelfCollision
						println("Self-collision detected")
					} else {
						result[i] |= TurnResultSnakeCollision
						println("Collision detected with another snake's body")
					}
					deadByCollision[i] = true
					drawPossible[i] = true // Mark as draw possible
					eliminatedSnakeCount++
				}
			}
		}

		// Check for head collisions
		for j, otherSnake := range g.Board.Snakes {
			if !deadByCollision[i] && otherSnake.Alive && i != j {
				if snake.Head.X == otherSnake.Head.X && snake.Head.Y == otherSnake.Head.Y {
					if snake.Length <= otherSnake.Length {
						result[i] |= TurnResultHeadCollision
						println(i, j, "Head collision detected with another snake's head")
						if snake.Length == otherSnake.Length {
							drawPossible[i] = true // Mark as draw possible
						}
						deadByCollision[i] = true
						eliminatedSnakeCount++
						break
					} else {
						println(i, j, "Head collision detected with another snake's head, but this snake is longer")
					}
				}
			}
		}
	}

	// Update the alive status of snakes based on collisions
	for i, snake := range g.Board.Snakes {
		if deadByCollision[i] {
			snake.Alive = false // Snake dies from collision
			println("Snake has died due to collision")
		}
		if snake.Alive {
			result[i] |= TurnResultAlive
			if aliveSnakeCount-eliminatedSnakeCount == 1 {
				result[i] |= TurnResultWin // If this snake is the only one alive, it wins
			} else {
				result[i] |= TurnResultContinue // Snake is still alive and game continues
			}
		} else {
			result[i] |= TurnResultDead
			if aliveSnakeCount == eliminatedSnakeCount {
				if drawPossible[i] {
					result[i] |= TurnResultDraw // If all snakes are dead and this one is marked
				} else {
					result[i] |= TurnResultLoss // If all snakes are dead and this one is not marked, it loses
				}
			} else {
				result[i] |= TurnResultLoss // If this snake is dead but others are alive, it loses
			}
		}
		g.Board.Snakes[i] = snake // Update the snake in the map
	}
	return result
}

func (s *Sneak) move(nextHead Point, eaten bool) {
	if !s.Alive {
		println("Error: Cannot move a dead snake")
		return // Snake is dead, cannot move
	}
	if len(s.Body) == 0 {
		println("Error: Snake has no body to move")
		return // No body to move
	}

	// Move the body
	if eaten {
		s.Body = append([]Point{nextHead}, s.Body...)
		s.Health = 100
		s.Length += 1
	} else {
		tail := s.Body[len(s.Body)-1]
		if len(s.Body) <= 1 || !tail.Equal(s.Body[len(s.Body)-2]) {
			s.BodyLayer.setValue(tail.X, tail.Y, 0.0)
		}
		s.Body = append([]Point{nextHead}, s.Body[:len(s.Body)-1]...)
		s.Health -= 1
	}
	// clear the previous head layer value
	s.HeadLayer.setValue(s.Head.X, s.Head.Y, 0.0)

	// Set the new layer value
	if s.Length > 1 {
		s.BodyLayer.setValue(s.Head.X, s.Head.Y, 1.0)
	}
	s.HeadLayer.setValue(nextHead.X, nextHead.Y, 1.0)
	s.Head = nextHead
}

func newLayer(width, height int) Layer {
	layer := make([][]float32, height)
	for i := range layer {
		(layer)[i] = make([]float32, width)
	}
	return layer
}

func (l *Layer) setValue(x, y int, value float32) {
	// Check if the coordinates are within bounds
	if x < 0 || x >= len((*l)[0]) || y < 0 || y >= len(*l) {
		println("Error: Coordinates out of bounds")
		return // Out of bounds
	}
	(*l)[y][x] = value
}

func (l *Layer) getValue(x, y int) float32 {
	// Check if the coordinates are within bounds
	if x < 0 || x >= len((*l)[0]) || y < 0 || y >= len(*l) {
		println("Error: Coordinates out of bounds")
		return 0.0 // Out of bounds
	}
	return (*l)[y][x]
}

func (p *Point) move(direction Direction) {
	switch direction {
	case Up:
		p.Y++
	case Down:
		p.Y--
	case Left:
		p.X--
	case Right:
		p.X++
	}
}

func (p *Point) Equal(other Point) bool {
	return p.X == other.X && p.Y == other.Y
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
