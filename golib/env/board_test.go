package env

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPlaceSnake(t *testing.T) {
	got := *NewGame(Setting{
		Width:  11,
		Height: 11,
	})
	got.Board.placeSnake(Point{2, 2}, 3, 90)
	got.Board.placeSnake(Point{3, 3}, 2, 0)

	expected := NewBoardEditor(11, 11).
		placeSnake([]Point{{2, 2}, {2, 2}, {2, 2}}, 90).
		placeSnake([]Point{{3, 3}, {3, 3}}, 0).
		build()

	if d := cmp.Diff(got.Board, expected); len(d) != 0 {
		t.Errorf("placeSnake() mismatch (-got +expected):\n%s", d)
	}
}

func TestPlaceFood(t *testing.T) {
	got := *NewGame(Setting{
		Width:  11,
		Height: 11,
	})
	got.Board.placeFoods(Point{2, 2})
	got.Board.placeSnake(Point{2, 2}, 2, 90)

	expected := NewBoardEditor(11, 11).
		placeFoods(Point{2, 2}).
		placeSnake([]Point{{2, 2}, {2, 2}}, 90).
		build()

	if d := cmp.Diff(got.Board, expected); len(d) != 0 {
		t.Errorf("placeFoods() mismatch (-got +expected):\n%s", d)
	}
}
