package env

import mapset "github.com/deckarep/golang-set/v2"

type Board struct {
	Width     int
	Height    int
	Food      mapset.Set[Point]
	Snakes    []Sneak
	FoodLayer Layer
	Layer     BoardLayers
}

type BoardLayers struct {
	FoodLayer  Layer
	SnakeLayer []SnakeLayer
}

type SnakeLayer struct {
	BodyLayer Layer
	HeadLayer Layer
}

func NewBoard(width, height int) *Board {
	return &Board{
		Width:     width,
		Height:    height,
		Food:      mapset.NewSet[Point](),
		Snakes:    []Sneak{},
		FoodLayer: newLayer(width, height),
	}
}

func (b *Board) placeSnake(head Point, length int, health int) {
	snake := Sneak{
		Body:      make([]Point, length),
		Head:      head,
		Length:    length,
		Health:    health,
		Alive:     true,
		BodyLayer: newLayer(b.Width, b.Height),
		HeadLayer: newLayer(b.Width, b.Height),
	}
	// Initialize the snake's body
	for i := 0; i < length; i++ {
		snake.Body[i] = Point{
			X: head.X,
			Y: head.Y,
		}
	}
	// Set the head layer value
	snake.HeadLayer.setValue(head.X, head.Y, 1.0)
	// Set the body layer values
	snake.BodyLayer.setValue(head.X, head.Y, 1.0)
	b.Snakes = append(b.Snakes, snake)
}

func (b *Board) placeFoods(coords ...Point) {
	for _, coord := range coords {
		b.Food.Add(coord)
		b.FoodLayer.setValue(coord.X, coord.Y, 1.0)
	}
}
