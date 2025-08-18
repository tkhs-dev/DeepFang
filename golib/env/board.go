package env

import (
	"slices"

	mapset "github.com/deckarep/golang-set/v2"
)

type Board struct {
	Width  int
	Height int
	Food   mapset.Set[Point]
	Snakes []Snake
}

type Layer [][]int

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}
type Snake struct {
	Id     int
	Body   []Point
	Head   Point
	Length int
	Health int
	Alive  bool
}

//func newLayer(width, height int) Layer {
//	layer := make([][]int, height)
//	for i := range layer {
//		(layer)[i] = make([]int, width)
//	}
//	return layer
//}
//
//func (l *Layer) setValue(x, y int, value int) {
//	// Check if the coordinates are within bounds
//	if x < 0 || x >= len((*l)[0]) || y < 0 || y >= len(*l) {
//		println("Error: Coordinates out of bounds")
//		return // Out of bounds
//	}
//	(*l)[y][x] = value
//}
//
//func (l *Layer) getValue(x, y int) int {
//	// Check if the coordinates are within bounds
//	if x < 0 || x >= len((*l)[0]) || y < 0 || y >= len(*l) {
//		println("Error: Coordinates out of bounds")
//		return 0.0 // Out of bounds
//	}
//	return (*l)[y][x]
//}
//
//func (l *Layer) addValue(x, y int, value int) {
//	// Check if the coordinates are within bounds
//	if x < 0 || x >= len((*l)[0]) || y < 0 || y >= len(*l) {
//		println("Error: Coordinates out of bounds")
//		return // Out of bounds
//	}
//	(*l)[y][x] += value
//}

func (s *Snake) move(nextHead Point, eaten bool) {
	if !s.Alive {
		println("Error: Cannot GetMovedPoint a dead snake")
		return // Snake is dead, cannot GetMovedPoint
	}
	if len(s.Body) == 0 {
		println("Error: Snake has no body to GetMovedPoint")
		return // No body to GetMovedPoint
	}

	// GetMovedPoint the body
	if eaten {
		s.Body = append([]Point{nextHead}, s.Body...)
		s.Health = 100
		s.Length += 1
	} else {
		s.Body = append([]Point{nextHead}, s.Body[:len(s.Body)-1]...)
		s.Health -= 1
	}
	// Update the head
	s.Head = nextHead
}

func (p *Point) GetMovedPoint(direction Direction) Point {
	dx, dy := 0, 0
	switch direction {
	case Up:
		dx, dy = 0, 1
	case Down:
		dx, dy = 0, -1
	case Left:
		dx, dy = -1, 0
	case Right:
		dx, dy = 1, 0
	}
	return Point{
		X: p.X + dx,
		Y: p.Y + dy,
	}
}

func (p *Point) Equal(other Point) bool {
	return p.X == other.X && p.Y == other.Y
}

func NewBoard(width, height int) *Board {
	return &Board{
		Width:  width,
		Height: height,
		Food:   mapset.NewSet[Point](),
		Snakes: make([]Snake, 0),
	}
}

func (b *Board) Copy() Board {
	newBoard := *b                                     // Create a shallow copy of the board
	newBoard.Food = b.Food.Clone().(mapset.Set[Point]) // Clone the food set
	newBoard.Snakes = make([]Snake, len(b.Snakes))     // Create a new slice for snakes
	for i, snake := range b.Snakes {
		newSnake := snake                              // Create a shallow copy of the snake
		newSnake.Body = make([]Point, len(snake.Body)) // Create a new slice for the body
		copy(newSnake.Body, snake.Body)                // Copy the body points
		newBoard.Snakes[i] = newSnake                  // Assign the copied snake to the new board
	}
	return newBoard
}

func (b *Board) placeSnake(head Point, length int, health int) {
	snake := Snake{
		Id:     len(b.Snakes),
		Body:   make([]Point, length),
		Head:   head,
		Length: length,
		Health: health,
		Alive:  true,
	}
	// Initialize the snake's body
	for i := 0; i < length; i++ {
		snake.Body[i] = Point{
			X: head.X,
			Y: head.Y,
		}
	}
	b.Snakes = append(b.Snakes, snake)
}

func (b *Board) placeSnakeAt(body []Point, health int) {
	if len(body) == 0 {
		println("Error: Cannot place a snake with an empty body")
		return // Cannot place a snake with an empty body
	}
	head := body[0]
	snake := Snake{
		Id:     len(b.Snakes),
		Body:   body,
		Head:   head,
		Length: len(body),
		Health: health,
		Alive:  true,
	}
	b.Snakes = append(b.Snakes, snake)
}

func (b *Board) placeFoods(coords ...Point) {
	for _, coord := range coords {
		b.Food.Add(coord)
	}
}

func (b *Board) isFood(point Point) bool {
	return b.Food.ContainsOne(point)
}

func (b *Board) getCollision(snake *Snake) (*Snake, bool, bool) {
	for _, other := range b.Snakes {
		if !other.Alive {
			continue // Skip dead snakes
		}
		if snake.Id != other.Id && snake.Head.Equal(other.Head) {
			return &other, false, true // Collision with another snake's head
		}
		if len(other.Body) > 1 && slices.Contains(other.Body[1:], snake.Head) {
			return &other, other.Id == snake.Id, false // Collision with another snake's body
		}
	}
	return nil, false, false // No collision
}

func (b *Board) GetUnoccupiedPoints() mapset.Set[Point] {
	points := mapset.NewSet[Point]()
	for x := 0; x < b.Width; x++ {
		for y := 0; y < b.Height; y++ {
			point := Point{X: x, Y: y}
			if !b.Food.ContainsOne(point) && !slices.ContainsFunc(b.Snakes, func(s Snake) bool {
				return s.Head.Equal(point) || slices.Contains(s.Body, point)
			}) {
				points.Add(point)
			}
		}
	}
	return points
}

func (b *Board) IsPointOccupied(point Point, includeFood bool) bool {
	if includeFood && b.Food.ContainsOne(point) {
		return true // Point is occupied by food
	}
	for _, snake := range b.Snakes {
		if snake.Head.Equal(point) || slices.Contains(snake.Body, point) {
			return true // Point is occupied by a snake
		}
	}
	return false // Point is unoccupied
}
