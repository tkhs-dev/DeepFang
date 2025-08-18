package env

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestTurnResultConst(t *testing.T) {
	check := func(value uint32, mask uint32) bool {
		return (value&mask) == value && (((0xFFFFFFFF^mask)|value)&mask) == value
	}

	// check alive section
	if !check(TurnResultAlive, TurnResultMaskAlive) {
		t.Errorf("TurnResultAlive does not match TurnResultMaskAlive")
	}
	if !check(TurnResultDead, TurnResultMaskAlive) {
		t.Errorf("TurnResultDead does not match TurnResultMaskAlive")
	}
	// check result section
	if !check(TurnResultContinue, TurnResultMaskResult) {
		t.Errorf("TurnResultContinue does not match TurnResultMaskResult")
	}
	if !check(TurnResultWin, TurnResultMaskResult) {
		t.Errorf("TurnResultWin does not match TurnResultMaskResult")
	}
	if !check(TurnResultDraw, TurnResultMaskResult) {
		t.Errorf("TurnResultDraw does not match TurnResultMaskResult")
	}
	if !check(TurnResultLoss, TurnResultMaskResult) {
		t.Errorf("TurnResultLoss does not match TurnResultMaskResult")
	}
	// check cause section
	if !check(TurnResultStarvation, TurnResultMaskCause) {
		t.Errorf("TurnResultStarvation does not match TurnResultMaskCause")
	}
	if !check(TurnResultSelfCollision, TurnResultMaskCause) {
		t.Errorf("TurnResultSelfCollision does not match TurnResultMaskCause")
	}
	if !check(TurnResultWallCollision, TurnResultMaskCause) {
		t.Errorf("TurnResultWallCollision does not match TurnResultMaskCause")
	}
	if !check(TurnResultSnakeCollision, TurnResultMaskCause) {
		t.Errorf("TurnResultSnakeCollision does not match TurnResultMaskCause")
	}
	if !check(TurnResultHeadCollision, TurnResultMaskCause) {
		t.Errorf("TurnResultHeadCollision does not match TurnResultMaskCause")
	}
	if !check(TurnResultFoodEaten, TurnResultMaskFood) {
		t.Errorf("TurnResultFood does not match TurnResultMaskFood")
	}
}

func TestExecute_通常移動(t *testing.T) {
	got := *NewGame(Setting{
		Width:  11,
		Height: 11,
	})
	got.Board.placeSnake(Point{0, 5}, 3, 90)
	got.Board.placeSnake(Point{10, 5}, 3, 90)
	got.Board.placeSnake(Point{5, 0}, 3, 90)
	got.Board.placeSnake(Point{5, 10}, 3, 90)
	result, ended, _ := got.Execute([]Direction{Right, Left, Up, Down})

	expected := NewBoardEditor(11, 11).
		placeSnake([]Point{{1, 5}, {0, 5}, {0, 5}}, 89).
		placeSnake([]Point{{9, 5}, {10, 5}, {10, 5}}, 89).
		placeSnake([]Point{{5, 1}, {5, 0}, {5, 0}}, 89).
		placeSnake([]Point{{5, 9}, {5, 10}, {5, 10}}, 89).
		build()

	if d := cmp.Diff(got.Board, expected); len(d) != 0 {
		t.Errorf("Execute() mismatch (-got +expected):\n%s", d)
	}

	expectedResults := []uint32{
		TurnResultAlive | TurnResultContinue,
		TurnResultAlive | TurnResultContinue,
		TurnResultAlive | TurnResultContinue,
		TurnResultAlive | TurnResultContinue,
	}
	for i, res := range result {
		if res != expectedResults[i] {
			t.Errorf("Execute() result mismatch at index %d: got %b, want %b", i, res, expectedResults[i])
		}
	}

	expectEnded := false
	if ended != expectEnded {
		t.Errorf("Execute() ended mismatch: got %v, want %v", ended, expectEnded)
	}
}

func TestExecute_体力(t *testing.T) {
	got := *NewGame(Setting{
		Width:  11,
		Height: 11,
	})
	got.Board.placeSnake(Point{0, 0}, 2, 1)
	got.Board.placeSnake(Point{0, 10}, 2, 90)
	result, ended, _ := got.Execute([]Direction{Right, Right})

	expected := NewBoardEditor(11, 11).
		placeSnake([]Point{{1, 0}, {0, 0}}, 0).
		placeSnake([]Point{{1, 10}, {0, 10}}, 89).
		killSnakes(0).
		build()

	if d := cmp.Diff(got.Board, expected); len(d) != 0 {
		t.Errorf("Execute() mismatch (-got +expected):\n%s", d)
	}

	expectedResults := []uint32{
		TurnResultDead | TurnResultLoss | TurnResultStarvation,
		TurnResultAlive | TurnResultWin,
	}
	for i, res := range result {
		if res != expectedResults[i] {
			t.Errorf("Execute() result mismatch at index %d: got %b, want %b", i, res, expectedResults[i])
		}
	}

	expectEnded := true
	if ended != expectEnded {
		t.Errorf("Execute() ended mismatch: got %v, want %v", ended, expectEnded)
	}
}

func TestExecute_体力同時(t *testing.T) {
	got := *NewGame(Setting{
		Width:  11,
		Height: 11,
	})
	got.Board.placeSnake(Point{0, 0}, 2, 1)
	got.Board.placeSnake(Point{0, 10}, 2, 1)
	result, ended, _ := got.Execute([]Direction{Right, Right})

	expected := NewBoardEditor(11, 11).
		placeSnake([]Point{{1, 0}, {0, 0}}, 0).
		placeSnake([]Point{{1, 10}, {0, 10}}, 0).
		killSnakes(0, 1).
		build()

	if d := cmp.Diff(got.Board, expected); len(d) != 0 {
		t.Errorf("Execute() mismatch (-got +expected):\n%s", d)
	}

	expectedResults := []uint32{
		TurnResultDead | TurnResultDraw | TurnResultStarvation,
		TurnResultDead | TurnResultDraw | TurnResultStarvation,
	}
	for i, res := range result {
		if res != expectedResults[i] {
			t.Errorf("Execute() result mismatch at index %d: got %b, want %b", i, res, expectedResults[i])
		}
	}

	expectEnded := true
	if ended != expectEnded {
		t.Errorf("Execute() ended mismatch: got %v, want %v", ended, expectEnded)
	}
}

func TestExecute_体力同時_負け(t *testing.T) {
	got := *NewGame(Setting{
		Width:  11,
		Height: 11,
	})
	got.Board.placeSnake(Point{0, 0}, 2, 1)
	got.Board.placeSnake(Point{0, 10}, 2, 1)
	got.Board.placeSnake(Point{0, 5}, 2, 2)
	result, ended, _ := got.Execute([]Direction{Right, Right, Right})

	expected := NewBoardEditor(11, 11).
		placeSnake([]Point{{1, 0}, {0, 0}}, 0).
		placeSnake([]Point{{1, 10}, {0, 10}}, 0).
		placeSnake([]Point{{1, 5}, {0, 5}}, 1).
		killSnakes(0, 1).
		build()

	if d := cmp.Diff(got.Board, expected); len(d) != 0 {
		t.Errorf("Execute() mismatch (-got +expected):\n%s", d)
	}

	expectedResults := []uint32{
		TurnResultDead | TurnResultLoss | TurnResultStarvation,
		TurnResultDead | TurnResultLoss | TurnResultStarvation,
		TurnResultAlive | TurnResultWin,
	}
	for i, res := range result {
		if res != expectedResults[i] {
			t.Errorf("Execute() result mismatch at index %d: got %b, want %b", i, res, expectedResults[i])
		}
	}

	expectEnded := true
	if ended != expectEnded {
		t.Errorf("Execute() ended mismatch: got %v, want %v", ended, expectEnded)
	}
}

func TestExecute_餌_食べる(t *testing.T) {
	got := *NewGame(Setting{
		Width:  11,
		Height: 11,
	})
	got.Board.placeSnake(Point{0, 0}, 2, 1)
	got.Board.placeSnake(Point{0, 10}, 2, 1)
	got.Board.placeFoods(Point{1, 0})
	result, ended, _ := got.Execute([]Direction{Right, Right})

	expected := NewBoardEditor(11, 11).
		placeSnake([]Point{{1, 0}, {0, 0}, {0, 0}}, 100).
		placeSnake([]Point{{1, 10}, {0, 10}}, 0).
		killSnakes(1).
		build()

	if d := cmp.Diff(got.Board, expected); len(d) != 0 {
		t.Errorf("Execute() mismatch (-got +expected):\n%s", d)
	}

	expectedResults := []uint32{
		TurnResultAlive | TurnResultWin | TurnResultFoodEaten,
		TurnResultDead | TurnResultLoss | TurnResultStarvation,
	}
	for i, res := range result {
		if res != expectedResults[i] {
			t.Errorf("Execute() result mismatch at index %d: got %b, want %b", i, res, expectedResults[i])
		}
	}

	expectEnded := true
	if ended != expectEnded {
		t.Errorf("Execute() ended mismatch: got %v, want %v", ended, expectEnded)
	}
}

func TestExecute_餌_食べない(t *testing.T) {
	got := *NewGame(Setting{
		Width:  11,
		Height: 11,
	})
	got.Board.placeSnake(Point{0, 0}, 2, 1)
	got.Board.placeSnake(Point{0, 10}, 2, 1)
	got.Board.placeFoods(Point{2, 0})
	result, ended, _ := got.Execute([]Direction{Right, Right})

	expected := NewBoardEditor(11, 11).
		placeSnake([]Point{{1, 0}, {0, 0}}, 0).
		placeSnake([]Point{{1, 10}, {0, 10}}, 0).
		placeFoods(Point{2, 0}).
		killSnakes(0, 1).
		build()

	if d := cmp.Diff(got.Board, expected); len(d) != 0 {
		t.Errorf("Execute() mismatch (-got +expected):\n%s", d)
	}

	expectedResults := []uint32{
		TurnResultDead | TurnResultDraw | TurnResultStarvation,
		TurnResultDead | TurnResultDraw | TurnResultStarvation,
	}
	for i, res := range result {
		if res != expectedResults[i] {
			t.Errorf("Execute() result mismatch at index %d: got %b, want %b", i, res, expectedResults[i])
		}
	}

	expectEnded := true
	if ended != expectEnded {
		t.Errorf("Execute() ended mismatch: got %v, want %v", ended, expectEnded)
	}
}

func TestExecute_壁(t *testing.T) {
	got := *NewGame(Setting{
		Width:  11,
		Height: 11,
	})
	got.Board.placeSnake(Point{0, 5}, 3, 90)
	got.Board.placeSnake(Point{10, 5}, 3, 90)
	got.Board.placeSnake(Point{5, 0}, 3, 90)
	got.Board.placeSnake(Point{5, 10}, 3, 90)
	got.Board.placeSnake(Point{5, 5}, 1, 90)
	result, ended, _ := got.Execute([]Direction{Left, Right, Down, Up, Up})

	expected := NewBoardEditor(11, 11).
		placeSnake([]Point{{-1, 5}, {0, 5}, {0, 5}}, 89).
		placeSnake([]Point{{11, 5}, {10, 5}, {10, 5}}, 89).
		placeSnake([]Point{{5, -1}, {5, 0}, {5, 0}}, 89).
		placeSnake([]Point{{5, 11}, {5, 10}, {5, 10}}, 89).
		placeSnake([]Point{{5, 6}}, 89).
		killSnakes(0, 1, 2, 3).
		build()

	if d := cmp.Diff(got.Board, expected); len(d) != 0 {
		t.Errorf("Execute() mismatch (-got +expected):\n%s", d)
	}

	expectedResults := []uint32{
		TurnResultDead | TurnResultLoss | TurnResultWallCollision,
		TurnResultDead | TurnResultLoss | TurnResultWallCollision,
		TurnResultDead | TurnResultLoss | TurnResultWallCollision,
		TurnResultDead | TurnResultLoss | TurnResultWallCollision,
		TurnResultAlive | TurnResultWin,
	}
	for i, res := range result {
		if res != expectedResults[i] {
			t.Errorf("Execute() result mismatch at index %d: got %b, want %b", i, res, expectedResults[i])
		}
	}

	expectEnded := true
	if ended != expectEnded {
		t.Errorf("Execute() ended mismatch: got %v, want %v", ended, expectEnded)
	}
}

func TestExecute_壁同時(t *testing.T) {
	got := *NewGame(Setting{
		Width:  11,
		Height: 11,
	})
	got.Board.placeSnake(Point{0, 5}, 3, 90)
	got.Board.placeSnake(Point{10, 5}, 3, 90)
	got.Board.placeSnake(Point{5, 0}, 3, 90)
	got.Board.placeSnake(Point{5, 10}, 3, 90)
	result, ended, _ := got.Execute([]Direction{Left, Right, Down, Up})

	expected := NewBoardEditor(11, 11).
		placeSnake([]Point{{-1, 5}, {0, 5}, {0, 5}}, 89).
		placeSnake([]Point{{11, 5}, {10, 5}, {10, 5}}, 89).
		placeSnake([]Point{{5, -1}, {5, 0}, {5, 0}}, 89).
		placeSnake([]Point{{5, 11}, {5, 10}, {5, 10}}, 89).
		killSnakes(0, 1, 2, 3).
		build()

	if d := cmp.Diff(got.Board, expected); len(d) != 0 {
		t.Errorf("Execute() mismatch (-got +expected):\n%s", d)
	}

	expectedResults := []uint32{
		TurnResultDead | TurnResultDraw | TurnResultWallCollision,
		TurnResultDead | TurnResultDraw | TurnResultWallCollision,
		TurnResultDead | TurnResultDraw | TurnResultWallCollision,
		TurnResultDead | TurnResultDraw | TurnResultWallCollision,
	}
	for i, res := range result {
		if res != expectedResults[i] {
			t.Errorf("Execute() result mismatch at index %d: got %b, want %b", i, res, expectedResults[i])
		}
	}

	expectEnded := true
	if ended != expectEnded {
		t.Errorf("Execute() ended mismatch: got %v, want %v", ended, expectEnded)
	}
}

func TestExecute_自己衝突(t *testing.T) {
	got := *NewGame(Setting{
		Width:  11,
		Height: 11,
	})
	got.Board.placeSnake(Point{0, 0}, 3, 90)
	got.Board.placeSnake(Point{5, 5}, 3, 90)
	got.Execute([]Direction{Right, Right})
	result, ended, _ := got.Execute([]Direction{Right, Left})

	expected := NewBoardEditor(11, 11).
		placeSnake([]Point{{2, 0}, {1, 0}, {0, 0}}, 88).
		placeSnake([]Point{{5, 5}, {6, 5}, {5, 5}}, 88).
		killSnakes(1).
		build()

	if d := cmp.Diff(got.Board, expected); len(d) != 0 {
		t.Errorf("Execute() mismatch (-got +expected):\n%s", d)
	}

	expectedResults := []uint32{
		TurnResultAlive | TurnResultWin,
		TurnResultDead | TurnResultLoss | TurnResultSelfCollision,
	}
	for i, res := range result {
		if res != expectedResults[i] {
			t.Errorf("Execute() result mismatch at index %d: got %b, want %b", i, res, expectedResults[i])
		}
	}

	expectEnded := true
	if ended != expectEnded {
		t.Errorf("Execute() ended mismatch: got %v, want %v", ended, expectEnded)
	}
}

func TestExecute_自己衝突同時(t *testing.T) {
	got := *NewGame(Setting{
		Width:  11,
		Height: 11,
	})
	got.Board.placeSnake(Point{0, 0}, 3, 90)
	got.Board.placeSnake(Point{5, 5}, 3, 90)
	got.Execute([]Direction{Right, Right})
	result, ended, _ := got.Execute([]Direction{Left, Left})

	expected := NewBoardEditor(11, 11).
		placeSnake([]Point{{0, 0}, {1, 0}, {0, 0}}, 88).
		placeSnake([]Point{{5, 5}, {6, 5}, {5, 5}}, 88).
		killSnakes(0, 1).
		build()

	if d := cmp.Diff(got.Board, expected); len(d) != 0 {
		t.Errorf("Execute() mismatch (-got +expected):\n%s", d)
	}

	expectedResults := []uint32{
		TurnResultDead | TurnResultDraw | TurnResultSelfCollision,
		TurnResultDead | TurnResultDraw | TurnResultSelfCollision,
	}
	for i, res := range result {
		if res != expectedResults[i] {
			t.Errorf("Execute() result mismatch at index %d: got %b, want %b", i, res, expectedResults[i])
		}
	}

	expectEnded := true
	if ended != expectEnded {
		t.Errorf("Execute() ended mismatch: got %v, want %v", ended, expectEnded)
	}
}

func TestExecute_衝突(t *testing.T) {
	got := *NewGame(Setting{
		Width:  11,
		Height: 11,
	})
	got.Board.placeSnake(Point{5, 4}, 3, 90)
	got.Board.placeSnake(Point{5, 5}, 3, 90)
	result, ended, _ := got.Execute([]Direction{Up, Right})

	expected := NewBoardEditor(11, 11).
		placeSnake([]Point{{5, 5}, {5, 4}, {5, 4}}, 89).
		placeSnake([]Point{{6, 5}, {5, 5}, {5, 5}}, 89).
		killSnakes(0).
		build()

	if d := cmp.Diff(got.Board, expected); len(d) != 0 {
		t.Errorf("Execute() mismatch (-got +expected):\n%s", d)
	}

	expectedResults := []uint32{
		TurnResultDead | TurnResultLoss | TurnResultSnakeCollision,
		TurnResultAlive | TurnResultWin,
	}
	for i, res := range result {
		if res != expectedResults[i] {
			t.Errorf("Execute() result mismatch at index %d: got %b, want %b", i, res, expectedResults[i])
		}
	}

	expectEnded := true
	if ended != expectEnded {
		t.Errorf("Execute() ended mismatch: got %v, want %v", ended, expectEnded)
	}
}

func TestExecute_衝突同時(t *testing.T) {
	got := *NewGame(Setting{
		Width:  11,
		Height: 11,
	})
	got.Board.placeSnake(Point{4, 4}, 3, 90)
	got.Board.placeSnake(Point{6, 5}, 3, 90)
	got.Board.placeSnake(Point{4, 6}, 3, 90)
	got.Execute([]Direction{Right, Left, Down})
	result, ended, _ := got.Execute([]Direction{Up, Down, Down})

	expected := NewBoardEditor(11, 11).
		placeSnake([]Point{{5, 5}, {5, 4}, {4, 4}}, 88).
		placeSnake([]Point{{5, 4}, {5, 5}, {6, 5}}, 88).
		placeSnake([]Point{{4, 4}, {4, 5}, {4, 6}}, 88).
		killSnakes(0, 1, 2).
		build()

	if d := cmp.Diff(got.Board, expected); len(d) != 0 {
		t.Errorf("Execute() mismatch (-got +expected):\n%s", d)
	}

	expectedResults := []uint32{
		TurnResultDead | TurnResultDraw | TurnResultSnakeCollision,
		TurnResultDead | TurnResultDraw | TurnResultSnakeCollision,
		TurnResultDead | TurnResultDraw | TurnResultSnakeCollision,
	}
	for i, res := range result {
		if res != expectedResults[i] {
			t.Errorf("Execute() result mismatch at index %d: got %b, want %b", i, res, expectedResults[i])
		}
	}

	expectEnded := true
	if ended != expectEnded {
		t.Errorf("Execute() ended mismatch: got %v, want %v", ended, expectEnded)
	}
}

func TestExecute_衝突_餌食(t *testing.T) {
	got := *NewGame(Setting{
		Width:  11,
		Height: 11,
	})
	got.Board.placeSnake(Point{7, 0}, 2, 90)
	got.Board.placeSnake(Point{4, 0}, 2, 90)
	got.Board.placeSnake(Point{2, 0}, 2, 90)
	got.Board.placeFoods(Point{9, 0}, Point{6, 0})
	got.Execute([]Direction{Right, Right, Right})
	result, ended, _ := got.Execute([]Direction{Right, Right, Right})

	expected := NewBoardEditor(11, 11).
		placeSnake([]Point{{9, 0}, {8, 0}, {7, 0}}, 100).
		placeSnake([]Point{{6, 0}, {5, 0}, {4, 0}}, 100).
		placeSnake([]Point{{4, 0}, {3, 0}}, 88).
		killSnakes(2).
		build()

	if d := cmp.Diff(got.Board, expected); len(d) != 0 {
		t.Errorf("Execute() mismatch (-got +expected):\n%s", d)
	}

	expectedResults := []uint32{
		TurnResultAlive | TurnResultContinue | TurnResultFoodEaten,
		TurnResultAlive | TurnResultContinue | TurnResultFoodEaten,
		TurnResultDead | TurnResultLoss | TurnResultSnakeCollision,
	}
	for i, res := range result {
		if res != expectedResults[i] {
			t.Errorf("Execute() result mismatch at index %d: got %b, want %b", i, res, expectedResults[i])
		}
	}

	expectEnded := false
	if ended != expectEnded {
		t.Errorf("Execute() ended mismatch: got %v, want %v", ended, expectEnded)
	}
}

func TestExecute_頭衝突(t *testing.T) {
	got := *NewGame(Setting{
		Width:  11,
		Height: 11,
	})
	got.Board.placeSnake(Point{5, 4}, 4, 90)
	got.Board.placeSnake(Point{4, 5}, 3, 90)
	result, ended, _ := got.Execute([]Direction{Up, Right})

	expected := NewBoardEditor(11, 11).
		placeSnake([]Point{{5, 5}, {5, 4}, {5, 4}, {5, 4}}, 89).
		placeSnake([]Point{{5, 5}, {4, 5}, {4, 5}}, 89).
		killSnakes(1).
		build()

	if d := cmp.Diff(got.Board, expected); len(d) != 0 {
		t.Errorf("Execute() mismatch (-got +expected):\n%s", d)
	}

	expectedResults := []uint32{
		TurnResultAlive | TurnResultWin,
		TurnResultDead | TurnResultLoss | TurnResultHeadCollision,
	}
	for i, res := range result {
		if res != expectedResults[i] {
			t.Errorf("Execute() result mismatch at index %d: got %b, want %b", i, res, expectedResults[i])
		}
	}

	expectEnded := true
	if ended != expectEnded {
		t.Errorf("Execute() ended mismatch: got %v, want %v", ended, expectEnded)
	}
}

func TestExecute_頭衝突_餌(t *testing.T) {
	got := *NewGame(Setting{
		Width:  11,
		Height: 11,
	})
	got.Board.placeSnake(Point{5, 4}, 4, 90)
	got.Board.placeSnake(Point{4, 5}, 3, 90)
	got.Board.placeFoods(Point{5, 5})
	result, ended, _ := got.Execute([]Direction{Up, Right})

	expected := NewBoardEditor(11, 11).
		placeSnake([]Point{{5, 5}, {5, 4}, {5, 4}, {5, 4}, {5, 4}}, 100).
		placeSnake([]Point{{5, 5}, {4, 5}, {4, 5}, {4, 5}}, 100).
		killSnakes(1).
		build()

	if d := cmp.Diff(got.Board, expected); len(d) != 0 {
		t.Errorf("Execute() mismatch (-got +expected):\n%s", d)
	}

	expectedResults := []uint32{
		TurnResultAlive | TurnResultWin | TurnResultFoodEaten,
		TurnResultDead | TurnResultLoss | TurnResultHeadCollision | TurnResultFoodEaten,
	}
	for i, res := range result {
		if res != expectedResults[i] {
			t.Errorf("Execute() result mismatch at index %d: got %b, want %b", i, res, expectedResults[i])
		}
	}

	expectEnded := true
	if ended != expectEnded {
		t.Errorf("Execute() ended mismatch: got %v, want %v", ended, expectEnded)
	}
}

func TestExecute_頭衝突_一部(t *testing.T) {
	got := *NewGame(Setting{
		Width:  11,
		Height: 11,
	})
	got.Board.placeSnake(Point{5, 4}, 4, 90)
	got.Board.placeSnake(Point{4, 5}, 3, 90)
	got.Board.placeSnake(Point{0, 0}, 1, 90)
	result, ended, _ := got.Execute([]Direction{Up, Right, Right})

	expected := NewBoardEditor(11, 11).
		placeSnake([]Point{{5, 5}, {5, 4}, {5, 4}, {5, 4}}, 89).
		placeSnake([]Point{{5, 5}, {4, 5}, {4, 5}}, 89).
		placeSnake([]Point{{1, 0}}, 89).
		killSnakes(1).
		build()

	if d := cmp.Diff(got.Board, expected); len(d) != 0 {
		t.Errorf("Execute() mismatch (-got +expected):\n%s", d)
	}

	expectedResults := []uint32{
		TurnResultAlive | TurnResultContinue,
		TurnResultDead | TurnResultLoss | TurnResultHeadCollision,
		TurnResultAlive | TurnResultContinue,
	}
	for i, res := range result {
		if res != expectedResults[i] {
			t.Errorf("Execute() result mismatch at index %d: got %b, want %b", i, res, expectedResults[i])
		}
	}

	expectEnded := false
	if ended != expectEnded {
		t.Errorf("Execute() ended mismatch: got %v, want %v", ended, expectEnded)
	}
}

func TestExecute_頭衝突同時_引き分け(t *testing.T) {
	got := *NewGame(Setting{
		Width:  11,
		Height: 11,
	})
	got.Board.placeSnake(Point{5, 4}, 4, 90)
	got.Board.placeSnake(Point{4, 5}, 3, 90)
	got.Board.placeSnake(Point{6, 5}, 4, 90)
	result, ended, _ := got.Execute([]Direction{Up, Right, Left})

	expected := NewBoardEditor(11, 11).
		placeSnake([]Point{{5, 5}, {5, 4}, {5, 4}, {5, 4}}, 89).
		placeSnake([]Point{{5, 5}, {4, 5}, {4, 5}}, 89).
		placeSnake([]Point{{5, 5}, {6, 5}, {6, 5}, {6, 5}}, 89).
		killSnakes(0, 1, 2).
		build()

	if d := cmp.Diff(got.Board, expected); len(d) != 0 {
		t.Errorf("Execute() mismatch (-got +expected):\n%s", d)
	}

	expectedResults := []uint32{
		TurnResultDead | TurnResultDraw | TurnResultHeadCollision,
		TurnResultDead | TurnResultDraw | TurnResultHeadCollision,
		TurnResultDead | TurnResultDraw | TurnResultHeadCollision,
	}
	for i, res := range result {
		if res != expectedResults[i] {
			t.Errorf("Execute() result mismatch at index %d: got %b, want %b", i, res, expectedResults[i])
		}
	}

	expectEnded := true
	if ended != expectEnded {
		t.Errorf("Execute() ended mismatch: got %v, want %v", ended, expectEnded)
	}
}

func TestExecute_頭衝突同時_全引き分け(t *testing.T) {
	got := *NewGame(Setting{
		Width:  11,
		Height: 11,
	})
	got.Board.placeSnake(Point{5, 4}, 3, 90)
	got.Board.placeSnake(Point{4, 5}, 3, 90)
	got.Board.placeSnake(Point{6, 5}, 3, 90)
	result, ended, _ := got.Execute([]Direction{Up, Right, Left})

	expected := NewBoardEditor(11, 11).
		placeSnake([]Point{{5, 5}, {5, 4}, {5, 4}}, 89).
		placeSnake([]Point{{5, 5}, {4, 5}, {4, 5}}, 89).
		placeSnake([]Point{{5, 5}, {6, 5}, {6, 5}}, 89).
		killSnakes(0, 1, 2).
		build()

	if d := cmp.Diff(got.Board, expected); len(d) != 0 {
		t.Errorf("Execute() mismatch (-got +expected):\n%s", d)
	}

	expectedResults := []uint32{
		TurnResultDead | TurnResultDraw | TurnResultHeadCollision,
		TurnResultDead | TurnResultDraw | TurnResultHeadCollision,
		TurnResultDead | TurnResultDraw | TurnResultHeadCollision,
	}
	for i, res := range result {
		if res != expectedResults[i] {
			t.Errorf("Execute() result mismatch at index %d: got %b, want %b", i, res, expectedResults[i])
		}
	}

	expectEnded := true
	if ended != expectEnded {
		t.Errorf("Execute() ended mismatch: got %v, want %v", ended, expectEnded)
	}
}

func TestExecute_頭衝突同時_勝ち(t *testing.T) {
	got := *NewGame(Setting{
		Width:  11,
		Height: 11,
	})
	got.Board.placeSnake(Point{5, 4}, 4, 90)
	got.Board.placeSnake(Point{4, 5}, 3, 90)
	got.Board.placeSnake(Point{6, 5}, 3, 90)
	result, ended, _ := got.Execute([]Direction{Up, Right, Left})

	expected := NewBoardEditor(11, 11).
		placeSnake([]Point{{5, 5}, {5, 4}, {5, 4}, {5, 4}}, 89).
		placeSnake([]Point{{5, 5}, {4, 5}, {4, 5}}, 89).
		placeSnake([]Point{{5, 5}, {6, 5}, {6, 5}}, 89).
		killSnakes(1, 2).
		build()

	if d := cmp.Diff(got.Board, expected); len(d) != 0 {
		t.Errorf("Execute() mismatch (-got +expected):\n%s", d)
	}

	expectedResults := []uint32{
		TurnResultAlive | TurnResultWin,
		TurnResultDead | TurnResultLoss | TurnResultHeadCollision,
		TurnResultDead | TurnResultLoss | TurnResultHeadCollision,
	}
	for i, res := range result {
		if res != expectedResults[i] {
			t.Errorf("Execute() result mismatch at index %d: got %b, want %b", i, res, expectedResults[i])
		}
	}

	expectEnded := true
	if ended != expectEnded {
		t.Errorf("Execute() ended mismatch: got %v, want %v", ended, expectEnded)
	}
}

func TestExecute_頭衝突同時_多発(t *testing.T) {
	got := *NewGame(Setting{
		Width:  11,
		Height: 11,
	})
	got.Board.placeSnake(Point{4, 4}, 2, 90)
	got.Board.placeSnake(Point{6, 4}, 2, 90)
	got.Board.placeSnake(Point{4, 5}, 2, 90)
	got.Board.placeSnake(Point{6, 5}, 2, 90)
	result, ended, _ := got.Execute([]Direction{Right, Left, Right, Left})

	expected := NewBoardEditor(11, 11).
		placeSnake([]Point{{5, 4}, {4, 4}}, 89).
		placeSnake([]Point{{5, 4}, {6, 4}}, 89).
		placeSnake([]Point{{5, 5}, {4, 5}}, 89).
		placeSnake([]Point{{5, 5}, {6, 5}}, 89).
		killSnakes(0, 1, 2, 3).
		build()

	if d := cmp.Diff(got.Board, expected); len(d) != 0 {
		t.Errorf("Execute() mismatch (-got +expected):\n%s", d)
	}

	expectedResults := []uint32{
		TurnResultDead | TurnResultDraw | TurnResultHeadCollision,
		TurnResultDead | TurnResultDraw | TurnResultHeadCollision,
		TurnResultDead | TurnResultDraw | TurnResultHeadCollision,
		TurnResultDead | TurnResultDraw | TurnResultHeadCollision,
	}
	for i, res := range result {
		if res != expectedResults[i] {
			t.Errorf("Execute() result mismatch at index %d: got %b, want %b", i, res, expectedResults[i])
		}
	}

	expectEnded := true
	if ended != expectEnded {
		t.Errorf("Execute() ended mismatch: got %v, want %v", ended, expectEnded)
	}
}

func TestExecute_多要因_引き分け(t *testing.T) {
	got := *NewGame(Setting{
		Width:  11,
		Height: 11,
	})
	got.Board.placeSnake(Point{4, 4}, 2, 90)
	got.Board.placeSnake(Point{6, 4}, 2, 90)
	got.Board.placeSnake(Point{4, 3}, 2, 90)
	got.Board.placeSnake(Point{0, 0}, 2, 90)
	got.Board.placeSnake(Point{0, 10}, 2, 1)
	result, ended, _ := got.Execute([]Direction{Right, Left, Up, Left, Left})

	expected := NewBoardEditor(11, 11).
		placeSnake([]Point{{5, 4}, {4, 4}}, 89).
		placeSnake([]Point{{5, 4}, {6, 4}}, 89).
		placeSnake([]Point{{4, 4}, {4, 3}}, 89).
		placeSnake([]Point{{-1, 0}, {0, 0}}, 89).
		placeSnake([]Point{{-1, 10}, {0, 10}}, 0).
		killSnakes(0, 1, 2, 3, 4).
		build()

	if d := cmp.Diff(got.Board, expected); len(d) != 0 {
		t.Errorf("Execute() mismatch (-got +expected):\n%s", d)
	}

	expectedResults := []uint32{
		TurnResultDead | TurnResultDraw | TurnResultHeadCollision,
		TurnResultDead | TurnResultDraw | TurnResultHeadCollision,
		TurnResultDead | TurnResultDraw | TurnResultSnakeCollision,
		TurnResultDead | TurnResultDraw | TurnResultWallCollision,
		TurnResultDead | TurnResultDraw | TurnResultStarvation,
	}
	for i, res := range result {
		if res != expectedResults[i] {
			t.Errorf("Execute() result mismatch at index %d: got %b, want %b", i, res, expectedResults[i])
		}
	}

	expectEnded := true
	if ended != expectEnded {
		t.Errorf("Execute() ended mismatch: got %v, want %v", ended, expectEnded)
	}
}

func TestExecute_ソロ(t *testing.T) {
	got := *NewGame(Setting{
		Width:  11,
		Height: 11,
	})
	got.Board.placeSnake(Point{4, 4}, 2, 90)
	result, ended, _ := got.Execute([]Direction{Right})

	expected := NewBoardEditor(11, 11).
		placeSnake([]Point{{5, 4}, {4, 4}}, 89).
		build()

	if d := cmp.Diff(got.Board, expected); len(d) != 0 {
		t.Errorf("Execute() mismatch (-got +expected):\n%s", d)
	}

	expectedResults := []uint32{
		TurnResultAlive | TurnResultContinue,
	}
	for i, res := range result {
		if res != expectedResults[i] {
			t.Errorf("Execute() result mismatch at index %d: got %b, want %b", i, res, expectedResults[i])
		}
	}

	expectEnded := false
	if ended != expectEnded {
		t.Errorf("Execute() ended mismatch: got %v, want %v", ended, expectEnded)
	}
}

func TestExecute_ソロ_負け(t *testing.T) {
	got := *NewGame(Setting{
		Width:  11,
		Height: 11,
	})
	got.Board.placeSnake(Point{10, 4}, 2, 90)
	result, ended, _ := got.Execute([]Direction{Right})

	expected := NewBoardEditor(11, 11).
		placeSnake([]Point{{11, 4}, {10, 4}}, 89).
		killSnakes(0).
		build()

	if d := cmp.Diff(got.Board, expected); len(d) != 0 {
		t.Errorf("Execute() mismatch (-got +expected):\n%s", d)
	}

	expectedResults := []uint32{
		TurnResultDead | TurnResultLoss | TurnResultWallCollision,
	}
	for i, res := range result {
		if res != expectedResults[i] {
			t.Errorf("Execute() result mismatch at index %d: got %b, want %b", i, res, expectedResults[i])
		}
	}

	expectEnded := true
	if ended != expectEnded {
		t.Errorf("Execute() ended mismatch: got %v, want %v", ended, expectEnded)
	}
}
