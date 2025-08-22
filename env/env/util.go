package env

import "golib/rule"

func IsMoveSafe(board *rule.Board, id int, move rule.Direction) bool {
	if board == nil || id < 0 || id >= len(board.Snakes) {
		return false // Invalid board or snake ID
	}

	snake := &board.Snakes[id]
	if !snake.Alive {
		return false // Snake is dead, cannot move
	}

	nextHead := snake.Head.GetMovedPoint(move)

	// Check if the next head position is out of bounds
	if nextHead.X < 0 || nextHead.X >= board.Width || nextHead.Y < 0 || nextHead.Y >= board.Height {
		return false // Out of bounds
	}

	// Check if the next head position collides with any other snake's body
	return !board.IsPointOccupied(nextHead, false)
}
