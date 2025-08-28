package neuralnet

import "golib/rule"

func GetInputTensor(board rule.Board) *InputData {
	return &InputData{
		CNNInput:     boardToCNNInput(board),
		FeatureInput: boardToFeatureInput(board),
	}
}

func boardToCNNInput(board rule.Board) []float32 {
	width, height := board.Width+2, board.Height+2
	size := width * height
	zeroChannel := make([]float32, size*3)
	result := make([]float32, (len(board.Snakes)*3+2)*size)
	for i, snake := range board.Snakes {
		if snake.Alive {
			// Head
			headIndex := (snake.Head.Y+1)*width + (snake.Head.X + 1)
			result[3*i*size+headIndex] = 1.0
			// Body
			for _, bodyPart := range snake.Body {
				bodyIndex := (bodyPart.Y+1)*width + (bodyPart.X + 1)
				result[(3*i+1)*size+bodyIndex] = 1.0
			}
			// Tail
			tailIndex := (snake.Body[len(snake.Body)-1].Y+1)*width + (snake.Body[len(snake.Body)-1].X + 1)
			result[(3*i+2)*size+tailIndex] = 1.0
		} else {
			copy(result[3*i*size:3*(i+1)*size], zeroChannel)
		}
	}
	// Food
	foodIndex := 3 * len(board.Snakes) * size
	for _, food := range board.Food.ToSlice() {
		foodIndexPosition := (food.Y+1)*width + (food.X + 1)
		result[foodIndex+foodIndexPosition] = 1.0
	}
	// Walls
	wallIndex := foodIndex + size
	for x := 0; x < width; x++ {
		wallTopIndex := x
		wallBottomIndex := (height-1)*width + x
		result[wallIndex+wallTopIndex] = 1.0
		result[wallIndex+wallBottomIndex] = 1.0
	}
	for y := 0; y < height; y++ {
		wallLeftIndex := y * width
		wallRightIndex := y*width + (width - 1)
		result[wallIndex+wallLeftIndex] = 1.0
		result[wallIndex+wallRightIndex] = 1.0
	}
	return result
}

func boardToFeatureInput(board rule.Board) []float32 {
	result := make([]float32, 0)
	for _, snake := range board.Snakes {
		if snake.Alive {
			result = append(result, float32(snake.Health)/100.0)                             // Normalize health
			result = append(result, float32(snake.Length)/float32(board.Width*board.Height)) // Normalize length
		} else {
			result = append(result, 0.0, 0.0) // Dead snakes contribute zero
		}
	}
	return result
}
