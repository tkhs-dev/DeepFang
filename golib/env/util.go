package env

func CoordsToLayer(coords []Point, width int, height int) Layer {
	layer := make(Layer, height)
	for i := range layer {
		layer[i] = make([]int, width)
	}

	for _, coord := range coords {
		if coord.X < 0 || coord.X >= width || coord.Y < 0 || coord.Y >= height {
			continue // Skip out-of-bounds coordinates
		}
		layer[coord.Y][coord.X] = 1.0
	}

	return layer
}

func All[T any](slice []T, predicate func(T) bool) bool {
	for _, v := range slice {
		if !predicate(v) {
			return false
		}
	}
	return true
}
