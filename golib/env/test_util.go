package env

import mapset "github.com/deckarep/golang-set/v2"

type BoardEditor struct {
	Width  int
	Height int
	Food   mapset.Set[Point]
	Snakes []Sneak
}

func NewBoardEditor(width, height int) *BoardEditor {
	be := &BoardEditor{
		Width:  width,
		Height: height,
		Food:   mapset.NewSet[Point](),
		Snakes: []Sneak{},
	}
	return be
}

func (be *BoardEditor) placeSnake(bodies []Point, health int) *BoardEditor {
	snake := Sneak{
		Body:      bodies,
		Head:      bodies[0],
		Length:    len(bodies),
		Health:    health,
		Alive:     true,
		BodyLayer: CoordsToLayer(bodies[1:], be.Width, be.Height),
		HeadLayer: CoordsToLayer([]Point{bodies[0]}, be.Width, be.Height),
	}
	be.Snakes = append(be.Snakes, snake)
	return be
}

func (be *BoardEditor) placeFoods(coords ...Point) *BoardEditor {
	for _, coord := range coords {
		if coord.X < 0 || coord.X >= be.Width || coord.Y < 0 || coord.Y >= be.Height {
			continue // Skip out-of-bounds coordinates
		}
		be.Food.Add(coord)
	}
	return be
}

func (be *BoardEditor) killSnakes(indexes ...int) *BoardEditor {
	for _, index := range indexes {
		if index < 0 || index >= len(be.Snakes) {
			continue // Skip out-of-bounds indexes
		}
		be.Snakes[index].Alive = false
	}
	return be
}

func (be *BoardEditor) build() Board {
	return Board{
		Width:     be.Width,
		Height:    be.Height,
		Food:      be.Food,
		Snakes:    be.Snakes,
		FoodLayer: CoordsToLayer(be.Food.ToSlice(), be.Width, be.Height),
	}
}
