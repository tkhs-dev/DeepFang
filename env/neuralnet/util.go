package neuralnet

import "golib/rule"

func GetInputTensor(board rule.Board) *InputData {
	return &InputData{
		CNNInput:     boardToCNNInput(board),
		FeatureInput: boardToFeatureInput(board),
	}
}

func boardToCNNInput(board rule.Board) []float32 {
	zeroChannel := make([]float32, board.Width*board.Height*3)
	size := board.Width * board.Height
	result := make([]float32, (len(board.Snakes)*3+1)*size)
	for i, snake := range board.Snakes {
		if snake.Alive {
			// Head
			headIndex := snake.Head.Y*board.Width + snake.Head.X
			result[3*i*size+headIndex] = 1.0
			// Body
			for _, bodyPart := range snake.Body {
				bodyIndex := bodyPart.Y*board.Width + bodyPart.X
				result[(3*i+1)*size+bodyIndex] = 1.0
			}
			// Tail
			tailIndex := snake.Body[len(snake.Body)-1].Y*board.Width + snake.Body[len(snake.Body)-1].X
			result[(3*i+2)*size+tailIndex] = 1.0
		} else {
			copy(result[3*i*size:3*(i+1)*size], zeroChannel)
		}
	}
	// Food
	foodIndex := 3 * len(board.Snakes) * size
	for _, food := range board.Food.ToSlice() {
		foodIndexPosition := food.Y*board.Width + food.X
		result[foodIndex+foodIndexPosition] = 1.0
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
