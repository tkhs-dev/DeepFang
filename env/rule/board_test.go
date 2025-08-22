package rule

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
		PlaceSnake([]Point{{2, 2}, {2, 2}, {2, 2}}, 90).
		PlaceSnake([]Point{{3, 3}, {3, 3}}, 0).
		Build()

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
		PlaceFoods(Point{2, 2}).
		PlaceSnake([]Point{{2, 2}, {2, 2}}, 90).
		Build()

	if d := cmp.Diff(got.Board, expected); len(d) != 0 {
		t.Errorf("PlaceFoods() mismatch (-got +expected):\n%s", d)
	}
}
