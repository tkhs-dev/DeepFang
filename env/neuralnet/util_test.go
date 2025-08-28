package neuralnet

import (
	"golib/rule"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestBoardToCNNInput(t *testing.T) {
	board := rule.NewBoardEditor(4, 4).
		PlaceSnake([]rule.Point{{0, 0}, {0, 1}}, 100).
		PlaceSnake([]rule.Point{{1, 0}, {1, 1}, {1, 2}}, 100).
		PlaceFoods(rule.Point{2, 2}, rule.Point{3, 3}).
		Build()
	input := boardToCNNInput(board)
	expected := []float32{
		// Head
		0, 0, 0, 0, 0, 0,
		0, 1, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		// Body
		0, 0, 0, 0, 0, 0,
		0, 1, 0, 0, 0, 0,
		0, 1, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		// Tail
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 1, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		// Head
		0, 0, 0, 0, 0, 0,
		0, 0, 1, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		// Body
		0, 0, 0, 0, 0, 0,
		0, 0, 1, 0, 0, 0,
		0, 0, 1, 0, 0, 0,
		0, 0, 1, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		// Tail
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 1, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		// Food
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 1, 0, 0,
		0, 0, 0, 0, 1, 0,
		0, 0, 0, 0, 0, 0,
		// Walls
		1, 1, 1, 1, 1, 1,
		1, 0, 0, 0, 0, 1,
		1, 0, 0, 0, 0, 1,
		1, 0, 0, 0, 0, 1,
		1, 0, 0, 0, 0, 1,
		1, 1, 1, 1, 1, 1,
	}
	if d := cmp.Diff(input, expected); len(d) != 0 {
		t.Errorf("Execute() mismatch (-got +expected):\n%s", d)
	}
}
