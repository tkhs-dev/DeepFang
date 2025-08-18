package env

import (
	"fmt"
	"math/rand"

	"github.com/BattlesnakeOfficial/rules"
)

type Setting struct {
	Width           int   `json:"width"`
	Height          int   `json:"height"`
	Seed            int64 `json:"seed"`
	FoodSpawnChance int   `json:"food_spawn_chance"`
	MinimumFood     int   `json:"minimum_food"`
	InitialTurn     int   `json:"initial_turn"`
}

type SnakeSetting struct {
	InitialLength int     `json:"length"`
	InitialHealth int     `json:"health"`
	InitialBody   []Point `json:"body"`
}

type Game struct {
	Setting Setting
	Board   Board
	Rand    Rand
	Turn    int
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

func NewGame(setting Setting) *Game {
	return &Game{
		Setting: setting,
		Board:   *NewBoard(setting.Width, setting.Height),
		Rand:    NewStandardRand(setting.Seed),
		Turn:    setting.InitialTurn,
	}
}

func (g *Game) Setup(snakeSettings []SnakeSetting) {
	err := PlaceSnakesAutomatically(&g.Board, g.Rand, snakeSettings)
	if err != nil {
		return
	}
	err = PlaceFoodAutomatically(g.Rand, &g.Board)
	if err != nil {
		println("Error placing food:", err.Error())
		return
	}
	println("Game setup complete with", len(g.Board.Snakes), "snakes and", g.Board.Food.Cardinality(), "food items.")
}

func (g *Game) Execute(moves []Direction) ([]uint32, bool, error) {
	if len(moves) != len(g.Board.Snakes) {
		return nil, false, fmt.Errorf("number of moves (%d) does not match number of snakes (%d)", len(moves), len(g.Board.Snakes))
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
		nextHead := snake.Head.GetMovedPoint(move)
		// Check if the snake eats food
		eaten := g.Board.isFood(nextHead)
		snake.move(nextHead, eaten)
		if eaten {
			result[i] |= TurnResultFoodEaten
			vBoard.Food.Remove(nextHead)
		} else if snake.Health <= 0 {
			result[i] |= TurnResultStarvation
			snake.Alive = false // Snake dies from starvation
			eliminatedSnakeCount++
			continue
		}

		// Check for wall collisions
		if snake.Head.X < 0 || snake.Head.X >= g.Board.Width || snake.Head.Y < 0 || snake.Head.Y >= g.Board.Height {
			result[i] |= TurnResultWallCollision
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
				vsnake.Alive = false
				eliminatedSnakeCount++
			} else if headCollision {
				if snake.Length <= collidedSnake.Length {
					result[i] |= TurnResultHeadCollision
					if snake.Length == collidedSnake.Length && vBoard.Snakes[collidedSnake.Id].Alive {
						result[collidedSnake.Id] |= TurnResultHeadCollision
						vBoard.Snakes[collidedSnake.Id].Alive = false
						eliminatedSnakeCount++
					}
					vsnake.Alive = false
					eliminatedSnakeCount++
				}
			} else {
				result[i] |= TurnResultSnakeCollision
				vsnake.Alive = false
				eliminatedSnakeCount++
			}
		}
	}

	g.Board = vBoard // Update the original board with the modified state
	ended := false
	for i, snake := range g.Board.Snakes {
		if !snake.Alive {
			result[i] |= TurnResultDead
			if aliveSnakeCount == eliminatedSnakeCount {
				if len(g.Board.Snakes) == 1 {
					result[i] |= TurnResultLoss
				} else {
					result[i] |= TurnResultDraw
				}
				ended = true
			} else {
				result[i] |= TurnResultLoss
			}
		} else {
			result[i] |= TurnResultAlive
			if len(g.Board.Snakes) > 1 && aliveSnakeCount-eliminatedSnakeCount == 1 {
				result[i] |= TurnResultWin
				ended = true
			} else {
				result[i] |= TurnResultContinue
			}
		}
	}

	g.Turn++
	// New food placement
	foodNeeded := foodDemand(len(g.Board.Food.ToSlice()), g.Setting)
	if foodNeeded > 0 {
		err := placeFoodRandomly(&g.Board, g.Rand, foodNeeded)
		if err != nil {
			println("Error placing food:", err.Error())
			return nil, ended, err
		}
	}

	return result, ended, nil
}

func (d Direction) String() string {
	switch d {
	case Up:
		return "Up"
	case Down:
		return "Down"
	case Left:
		return "Left"
	case Right:
		return "Right"
	default:
		return "Unknown Direction"
	}
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

func PlaceFoodAutomatically(rand Rand, b *Board) error {
	if b.Width == b.Height && b.Width >= BoardSizeSmall {
		return PlaceFoodFixed(b, rand)
	}

	return placeFoodRandomly(b, rand, len(b.Snakes))
}

func placeFoodRandomly(board *Board, rand Rand, n int) error {
	unoccupiedPoints := board.GetUnoccupiedPoints()
	placeFoodRandomlyAtPositions(board, rand, n, unoccupiedPoints.ToSlice())
	return nil
}

func placeFoodRandomlyAtPositions(board *Board, rand rules.Rand, n int, positions []Point) {
	if len(positions) < n {
		n = len(positions)
	}

	rand.Shuffle(len(positions), func(i int, j int) {
		positions[i], positions[j] = positions[j], positions[i]
	})

	board.placeFoods(positions[:n]...)
}

func PlaceFoodFixed(board *Board, rand Rand) error {
	width, height := board.Width, board.Height
	centerCoord := Point{X: (width - 1) / 2, Y: (height - 1) / 2}

	isSmallBoard := width*height < BoardSizeMedium*BoardSizeMedium

	// Up to 4 snakes can be placed such that food is nearby on small boards.
	// Otherwise, we skip this and only try to place food in the center.
	snakes := board.Snakes
	if len(snakes) <= 4 || !isSmallBoard {
		// Place 1 food within exactly 2 moves of each snake, but never towards the center or in a corner
		for _, snake := range snakes {
			snakeHead := snake.Head
			possibleFoodLocations := []Point{
				{X: snakeHead.X - 1, Y: snakeHead.Y - 1},
				{X: snakeHead.X - 1, Y: snakeHead.Y + 1},
				{X: snakeHead.X + 1, Y: snakeHead.Y - 1},
				{X: snakeHead.X + 1, Y: snakeHead.Y + 1},
			}

			// Remove any invalid/unwanted positions
			var availableFoodLocations []Point
			for _, p := range possibleFoodLocations {

				// Don't place in the center
				if centerCoord == p {
					continue
				}

				// Ignore points already occupied by food or hazards
				if board.IsPointOccupied(p, true) {
					continue
				}

				// Food must be further than snake from center on at least one axis
				isAwayFromCenter := false
				if p.X < snakeHead.X && snakeHead.X < centerCoord.X {
					isAwayFromCenter = true
				} else if centerCoord.X < snakeHead.X && snakeHead.X < p.X {
					isAwayFromCenter = true
				} else if p.Y < snakeHead.Y && snakeHead.Y < centerCoord.Y {
					isAwayFromCenter = true
				} else if centerCoord.Y < snakeHead.Y && snakeHead.Y < p.Y {
					isAwayFromCenter = true
				}
				if !isAwayFromCenter {
					continue
				}

				// Don't spawn food in corners
				if (p.X == 0 || p.X == (width-1)) && (p.Y == 0 || p.Y == (height-1)) {
					continue
				}

				availableFoodLocations = append(availableFoodLocations, p)
			}

			if len(availableFoodLocations) <= 0 {
				return rules.ErrorNoRoomForFood
			}

			// Select randomly from available locations
			placedFood := availableFoodLocations[rand.Intn(len(availableFoodLocations))]
			board.placeFoods(placedFood)
		}
	}

	// Finally, try to place 1 food in center of board for dramatic purposes
	if !board.IsPointOccupied(centerCoord, true) {
		board.placeFoods(centerCoord)
	}

	return nil
}

func PlaceSnakesAutomatically(b *Board, rand Rand, snakeSettings []SnakeSetting) error {

	if b.Width == b.Height {
		// we don't allow > 8 snakes on very small boards
		if len(snakeSettings) > 8 && b.Width < BoardSizeSmall {
			return fmt.Errorf("too many snakes (%d) for board size %dx%d", len(snakeSettings), b.Width, b.Height)
		}

		// we can do fixed placement for up to 8 snakes on minimum sized boards
		if len(snakeSettings) <= 8 && b.Width >= BoardSizeSmall {
			return PlaceSnakesFixed(b, rand, snakeSettings)
		}

		// for > 8 snakes, we can do distributed placement
		if b.Width >= BoardSizeMedium {
			return PlaceManySnakesDistributed(b, rand, snakeSettings)
		}
	}

	// last resort for unexpected board sizes we'll just randomly place snakes
	return PlaceSnakesRandomly(b, rand, snakeSettings)
}

func PlaceSnakesFixed(b *Board, rand Rand, snakeSettings []SnakeSetting) error {
	// Create start 8 points
	mn, md, mx := 1, (b.Width-1)/2, b.Width-2
	cornerPoints := []Point{
		{X: mn, Y: mn},
		{X: mn, Y: mx},
		{X: mx, Y: mn},
		{X: mx, Y: mx},
	}
	cardinalPoints := []Point{
		{X: mn, Y: md},
		{X: md, Y: mn},
		{X: md, Y: mx},
		{X: mx, Y: md},
	}

	// Sanity check
	if len(snakeSettings) > (len(cornerPoints) + len(cardinalPoints)) {
		return fmt.Errorf("too many snakes (%d) for fixed placement on board size %dx%d", len(b.Snakes), b.Width, b.Height)
	}

	// Randomly order them
	rand.Shuffle(len(cornerPoints), func(i int, j int) {
		cornerPoints[i], cornerPoints[j] = cornerPoints[j], cornerPoints[i]
	})
	rand.Shuffle(len(cardinalPoints), func(i int, j int) {
		cardinalPoints[i], cardinalPoints[j] = cardinalPoints[j], cardinalPoints[i]
	})

	var startPoints []Point
	if rand.Intn(2) == 0 {
		startPoints = append(startPoints, cornerPoints...)
		startPoints = append(startPoints, cardinalPoints...)
	} else {
		startPoints = append(startPoints, cardinalPoints...)
		startPoints = append(startPoints, cornerPoints...)
	}

	// Assign to snakes in order given
	for i := range snakeSettings {
		b.placeSnake(startPoints[i], snakeSettings[i].InitialLength, snakeSettings[i].InitialHealth)
	}

	return nil
}

// PlaceManySnakesDistributed is a placement algorithm that works for up to 16 snakes
// It is intended for use on large boards and distributes snakes relatively evenly,
// and randomly, across quadrants.
func PlaceManySnakesDistributed(b *Board, rand Rand, snakeSettings []SnakeSetting) error {
	// this placement algorithm supports up to 16 snakes
	if len(snakeSettings) > 16 {
		return fmt.Errorf("too many snakes (%d) for distributed placement on board size %dx%d", len(snakeSettings), b.Width, b.Height)
	}

	quadHSpace := b.Width / 2
	quadVSpace := b.Height / 2

	hOffset := quadHSpace / 3
	vOffset := quadVSpace / 3

	quads := make([]RandomPositionBucket, 4)

	// quad 1
	quads[0] = RandomPositionBucket{}
	quads[0].Fill(
		Point{X: hOffset, Y: vOffset},
		Point{X: quadHSpace - hOffset, Y: vOffset},
		Point{X: hOffset, Y: quadVSpace - vOffset},
		Point{X: quadHSpace - hOffset, Y: quadVSpace - vOffset},
	)

	// quad 2
	quads[1] = RandomPositionBucket{}
	for _, p := range quads[0].positions {
		quads[1].Fill(Point{X: b.Width - p.X - 1, Y: p.Y})
	}

	// quad 3
	quads[2] = RandomPositionBucket{}
	for _, p := range quads[0].positions {
		quads[2].Fill(Point{X: p.X, Y: b.Height - p.Y - 1})
	}

	// quad 4
	quads[3] = RandomPositionBucket{}
	for _, p := range quads[0].positions {
		quads[3].Fill(Point{X: b.Width - p.X - 1, Y: b.Height - p.Y - 1})
	}

	currentQuad := rand.Intn(4) // randomly pick a quadrant to start from
	// evenly distribute snakes across quadrants, randomly, by rotating through the quadrants
	for i := range snakeSettings {
		p, err := quads[currentQuad].Take(rand)
		if err != nil {
			return err
		}
		b.placeSnake(p, snakeSettings[i].InitialLength, snakeSettings[i].InitialHealth)

		currentQuad = (currentQuad + 1) % 4
	}

	return nil
}

type RandomPositionBucket struct {
	positions []Point
}

func (rpb *RandomPositionBucket) Fill(p ...Point) {
	rpb.positions = append(rpb.positions, p...)
}

func (rpb *RandomPositionBucket) Take(rand Rand) (Point, error) {
	if len(rpb.positions) == 0 {
		return Point{}, fmt.Errorf("no positions available to take from RandomPositionBucket")
	}

	// randomly pick the next position
	idx := rand.Intn(len(rpb.positions))
	p := rpb.positions[idx]

	// remove that position from the list using the fast slice removal method
	rpb.positions[idx] = rpb.positions[len(rpb.positions)-1]
	rpb.positions = rpb.positions[:len(rpb.positions)-1]

	return p, nil
}

func PlaceSnakesRandomly(b *Board, rand Rand, snakeSettings []SnakeSetting) error {
	for i := 0; i < len(b.Snakes); i++ {
		unoccupiedPoints := removeCenterCoord(b, GetEvenUnoccupiedPoints(b))
		if len(unoccupiedPoints) <= 0 {
			return fmt.Errorf("no unoccupied points available to place snake %d", i)
		}
		p := unoccupiedPoints[rand.Intn(len(unoccupiedPoints))]
		b.placeSnake(p, snakeSettings[i].InitialLength, snakeSettings[i].InitialHealth)
	}
	return nil
}

func removeCenterCoord(b *Board, points []Point) []Point {
	centerCoord := Point{X: (b.Width - 1) / 2, Y: (b.Height - 1) / 2}
	var noCenterPoints []Point
	for _, p := range points {
		if p != centerCoord {
			noCenterPoints = append(noCenterPoints, p)
		}
	}

	return noCenterPoints
}

func GetEvenUnoccupiedPoints(b *Board) []Point {
	// Start by getting unoccupied points
	unoccupiedPoints := b.GetUnoccupiedPoints()

	// Create a new array to hold points that are  even
	evenUnoccupiedPoints := []Point{}

	for _, point := range unoccupiedPoints.ToSlice() {
		if ((point.X + point.Y) % 2) == 0 {
			evenUnoccupiedPoints = append(evenUnoccupiedPoints, point)
		}
	}
	return evenUnoccupiedPoints
}
