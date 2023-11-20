package main

import (
	"fmt"
	"testing"

	"github.com/denilai/maybe"
)

type Test[A, W any] struct {
	Arg  A
	Func func(A) W
	Want W
}

func perform[A any, B comparable](msg string, x Test[A, B], t *testing.T) {
	want := x.Want
	arg := x.Arg
	res := x.Func(arg)
	if want != res {
		t.Fatalf("%v\n%v. Ожидался `%v`, получен `%v`", arg, msg, want, res)
	}
}

func TestNoMainDiagWinner(t *testing.T) {
	games := make([]Test[Board, maybe.Maybe[Figure]], 0)
	// Есть победитель по строке и по вторичной диагонали, но не по главной диагонали
	G1 := Board{{X, Empty{}, O}, {O, O, O}, {X, Empty{}, O}}
	// Есть победитель по столбцу и по строке, но не по главной диагонали
	G2 := Board{{Empty{}, O, O}, {O, O, O}, {O, O, Empty{}}}
	// Пустое поле
	G3 := Board{{Empty{}, Empty{}, Empty{}}, {Empty{}, Empty{}, Empty{}}, {Empty{}, Empty{}, Empty{}}}
	games = append(games, Test[Board, maybe.Maybe[Figure]]{Arg: G1, Func: MainDiagWinner, Want: maybe.Nothing[Figure]()})
	games = append(games, Test[Board, maybe.Maybe[Figure]]{Arg: G2, Func: MainDiagWinner, Want: maybe.Nothing[Figure]()})
	games = append(games, Test[Board, maybe.Maybe[Figure]]{Arg: G3, Func: MainDiagWinner, Want: maybe.Nothing[Figure]()})
	for i, game := range games {
		perform(fmt.Sprintf("(%v) Oшибка поиска победителя по главной диагонали (отсутствие победителя)", i), game, t)
	}
}

func TestDiagMainWinner(t *testing.T) {
	games := make([]Test[Board, maybe.Maybe[Figure]], 0)
	G1 := Board{{O, X, Empty{}}, {O, O, X}, {O, Empty{}, O}}
	G2 := Board{{X, O, Empty{}}, {O, X, X}, {O, Empty{}, X}}
	games = append(games, Test[Board, maybe.Maybe[Figure]]{Arg: G1, Func: MainDiagWinner, Want: maybe.Just(O)})
	games = append(games, Test[Board, maybe.Maybe[Figure]]{Arg: G2, Func: MainDiagWinner, Want: maybe.Just(X)})
	for i, game := range games {
		perform(fmt.Sprintf("(%v) Oшибка поиска победителя по главной диагонали", i), game, t)
	}
}

func TestNoColWinner(t *testing.T) {
	games := make([]Test[Board, maybe.Maybe[Figure]], 0)
	// Есть победитель по строке, но не по столбцу
	G1 := Board{{O, O, O}, {Empty{}, O, X}, {X, Empty{}, Empty{}}}
	// Есть победитель по диагонали и по строке, но не по столбцу
	G2 := Board{{Empty{}, Empty{}, O}, {O, O, O}, {O, O, Empty{}}}
	// Пустое поле
	G3 := Board{{Empty{}, Empty{}, Empty{}}, {Empty{}, Empty{}, Empty{}}, {Empty{}, Empty{}, Empty{}}}
	games = append(games, Test[Board, maybe.Maybe[Figure]]{Arg: G1, Func: ColWinner, Want: maybe.Nothing[Figure]()})
	games = append(games, Test[Board, maybe.Maybe[Figure]]{Arg: G2, Func: ColWinner, Want: maybe.Nothing[Figure]()})
	games = append(games, Test[Board, maybe.Maybe[Figure]]{Arg: G3, Func: ColWinner, Want: maybe.Nothing[Figure]()})
	for i, game := range games {
		perform(fmt.Sprintf("(%v) Oшибка поиска победителя по столбцу (отсутствие победителя)", i), game, t)
	}
}

func TestColWinner(t *testing.T) {
	games := make([]Test[Board, maybe.Maybe[Figure]], 0)
	G1 := Board{{O, X, Empty{}}, {O, O, X}, {O, Empty{}, Empty{}}}
	G2 := Board{{X, O, O}, {O, O, O}, {X, O, X}}
	G3 := Board{{O, Empty{}, X}, {X, O, X}, {X, X, X}}
	games = append(games, Test[Board, maybe.Maybe[Figure]]{Arg: G1, Func: ColWinner, Want: maybe.Just(O)})
	games = append(games, Test[Board, maybe.Maybe[Figure]]{Arg: G2, Func: ColWinner, Want: maybe.Just(O)})
	games = append(games, Test[Board, maybe.Maybe[Figure]]{Arg: G3, Func: ColWinner, Want: maybe.Just(X)})
	for i, game := range games {
		perform(fmt.Sprintf("(%v) Oшибка поиска победителя по столбцу", i), game, t)
	}
}

func TestNoRowWinner(t *testing.T) {
	games := make([]Test[Board, maybe.Maybe[Figure]], 0)
	// Есть победитель по диагонали
	G1 := Board{{O, Empty{}, O}, {Empty{}, O, X}, {X, Empty{}, O}}
	// Есть победитель по столбцу, но не по строке
	G2 := Board{{O, Empty{}, Empty{}}, {O, O, X}, {O, O, Empty{}}}
	// Пустое поле
	G3 := Board{{Empty{}, Empty{}, Empty{}}, {Empty{}, Empty{}, Empty{}}, {Empty{}, Empty{}, Empty{}}}
	games = append(games, Test[Board, maybe.Maybe[Figure]]{Arg: G1, Func: RowWinner, Want: maybe.Nothing[Figure]()})
	games = append(games, Test[Board, maybe.Maybe[Figure]]{Arg: G2, Func: RowWinner, Want: maybe.Nothing[Figure]()})
	games = append(games, Test[Board, maybe.Maybe[Figure]]{Arg: G3, Func: RowWinner, Want: maybe.Nothing[Figure]()})
	for i, game := range games {
		perform(fmt.Sprintf("(%v) Oшибка поиска победителя по строке (отсутствие победителя)", i), game, t)
	}
}

func TestRowWinner(t *testing.T) {
	games := make([]Test[Board, maybe.Maybe[Figure]], 0)
	G1 := Board{{O, O, O}, {Empty{}, O, X}, {X, Empty{}, Empty{}}}
	G2 := Board{{X, X, O}, {O, O, O}, {X, Empty{}, X}}
	G3 := Board{{O, Empty{}, O}, {X, O, X}, {X, X, X}}
	games = append(games, Test[Board, maybe.Maybe[Figure]]{Arg: G1, Func: RowWinner, Want: maybe.Just(O)})
	games = append(games, Test[Board, maybe.Maybe[Figure]]{Arg: G2, Func: RowWinner, Want: maybe.Just(O)})
	games = append(games, Test[Board, maybe.Maybe[Figure]]{Arg: G3, Func: RowWinner, Want: maybe.Just(X)})
	for i, game := range games {
		perform(fmt.Sprintf("(%v) Oшибка поиска победителя по строке", i), game, t)
	}
}
