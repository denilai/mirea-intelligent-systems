package main

import (
	"fmt"
	"testing"

	"github.com/denilai/maybe"
)

type Test[A, W any] struct {
	Arg         A
	Func        func(A) W
	EFunc       func(A) (W, error)
	ExpectError bool
	Want        W
}

func perform[A any, B comparable](msg string, x Test[A, B], t *testing.T) {
	want := x.Want
	arg := x.Arg
	var res B
	if x.Func != nil {
		res = x.Func(arg)
	} else {
		res, _ = x.EFunc(arg)
	}
	if want != res {
		t.Fatalf("%v\n%v. Ожидался `%v`, получен `%v`", arg, msg, want, res)
	}
}

func TestDecode2(t *testing.T) {
	src := "O X X O X"
	want := Board{{O, Empty{}, X}, {Empty{}, X, Empty{}}, {O, Empty{}, X}}
	if b, err := Decode(src); err != nil {
		t.Fatal(err)
	} else if b.String() != want.String() {
		t.Fatalf("Ошибка декодирования поля")
	}
}

func TestEmptyDecode(t *testing.T) {
	src := "         "
	want := Board{{Empty{}, Empty{}, Empty{}}, {Empty{}, Empty{}, Empty{}}, {Empty{}, Empty{}, Empty{}}}
	if b, err := Decode(src); err != nil {
		t.Fatalf("Error")
	} else if b.String() != want.String() {
		t.Fatalf("Ошибка декодирования поля")
	}
}

func TestDecode(t *testing.T) {
	src1 := "X        "
	src2 := "O X X O X"
	src3 := "         "
	want1 := Board{{X, Empty{}, Empty{}}, {Empty{}, Empty{}, Empty{}}, {Empty{}, Empty{}, Empty{}}}
	want2 := Board{{O, Empty{}, X}, {Empty{}, X, Empty{}}, {O, Empty{}, X}}
	want3 := Board{{Empty{}, Empty{}, Empty{}}, {Empty{}, Empty{}, Empty{}}, {Empty{}, Empty{}, Empty{}}}
	srcs := []string{src1, src2, src3}
	wants := []Board{want1, want2, want3}
	tests := make([]Test[string, Board], len(srcs))
	for i := 0; i < len(srcs); i++ {
		tests[i] = Test[string, Board]{Want: wants[i], Arg: srcs[i], EFunc: Decode}
	}
	for _, test := range tests {
		if b, err := Decode(test.Arg); err != nil {
			t.Fatal(err)
		} else {
			fmt.Println(b)
			if b.String() != test.Want.String() {
				t.Fatalf("Ошибка декодирования поля")
			}
		}
	}
}

func TestEncode(t *testing.T) {
	G1 := NewBoard(3)
	G1.Set(Place{0, 0}, X)
	G1.Set(Place{0, 1}, O)
	G1.Set(Place{0, 2}, X)
	G1.Set(Place{1, 0}, O)
	want := "XOXO     "
	if res, _ := Encode(G1); want != res {
		t.Fatalf("Ожидался `%v`, получен `%v`", want, res)
	}
}
func TestEncodeEmpty(t *testing.T) {
	G1 := NewBoard(3)
	want := "         "
	if res, _ := Encode(G1); want != res {
		t.Fatalf("Ожидался `%v`, получен `%v`", want, res)
	}
}

func TestCopyBoard(t *testing.T) {
	G1 := NewBoard(3)
	copyG1 := NewBoard(3)
	if err := copyG1.Set(Place{Row: 1, Col: 1}, X); err != nil {
		panic(err)
	}
	if cell, err := G1.Get(Place{Row: 1, Col: 1}); err != nil {
		panic(err)
	} else if !cell.IsEmpty() {
		t.Fatalf("Ошибка копирования. Скопированы ссылки на ячейки, а не значения")
	}
}

func TestNextMove(t *testing.T) {
	G1 := NewBoard(3)
	G2 := NewBoard(3)
	G3 := NewBoard(3)
	G2.Set(Place{1, 2}, X)
	G2.Set(Place{0, 0}, O)
	G3.Set(Place{0, 0}, X)
	G3.Set(Place{1, 0}, O)
	G3.Set(Place{2, 0}, X)
	//G1.Set(Place{1, 1}, X))
	//G1.Set(Place{1, 0}, O})
	games := make([]Board, 0, 3)
	wants := make([]Figure, 0, 3)
	games = append(games, G1, G2, G3)
	wants = append(wants, X, X, O)
	for i := range games {
		next := games[i].NextMove()
		if next != wants[i] {
			t.Fatalf("%v\nОшибка определения следующего хода\n  Ожидалось: %v\n  Получено: %v\n", games[i], wants[i], next)
		}
	}
}

func TestNoWinner(t *testing.T) {
	G1 := NewBoard(3)
	games := make([]Test[Board, maybe.Maybe[Figure]], 0, 3)
	games = append(games, Test[Board, maybe.Maybe[Figure]]{Arg: G1, Func: Winner, Want: maybe.Nothing[Figure]()})
	for i, game := range games {
		perform(fmt.Sprintf("(%v) Oшибка поиска победителя (отсутствие победителя)", i), game, t)
	}

}

func TestWinner(t *testing.T) {
	G1 := Board{{X, Empty{}, O}, {O, O, O}, {X, Empty{}, O}}
	G2 := Board{{X, Empty{}, O}, {X, O, O}, {X, O, Empty{}}}
	G3 := Board{{Empty{}, O, O}, {X, O, X}, {O, X, Empty{}}}
	games := make([]Test[Board, maybe.Maybe[Figure]], 0, 3)
	games = append(games, Test[Board, maybe.Maybe[Figure]]{Arg: G1, Func: Winner, Want: maybe.Just(O)})
	games = append(games, Test[Board, maybe.Maybe[Figure]]{Arg: G2, Func: Winner, Want: maybe.Just(X)})
	games = append(games, Test[Board, maybe.Maybe[Figure]]{Arg: G3, Func: Winner, Want: maybe.Just(O)})
	for i, game := range games {
		perform(fmt.Sprintf("(%v) Oшибка поиска победителя", i), game, t)
	}
}

func TestNoMainDiagWinner(t *testing.T) {
	games := make([]Test[Board, maybe.Maybe[Figure]], 0, 3)
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
	games := make([]Test[Board, maybe.Maybe[Figure]], 0, 2)
	G1 := Board{{O, X, Empty{}}, {O, O, X}, {O, Empty{}, O}}
	G2 := Board{{X, O, Empty{}}, {O, X, X}, {O, Empty{}, X}}
	games = append(games, Test[Board, maybe.Maybe[Figure]]{Arg: G1, Func: MainDiagWinner, Want: maybe.Just(O)})
	games = append(games, Test[Board, maybe.Maybe[Figure]]{Arg: G2, Func: MainDiagWinner, Want: maybe.Just(X)})
	for i, game := range games {
		perform(fmt.Sprintf("(%v) Oшибка поиска победителя по главной диагонали", i), game, t)
	}
}

func TestNoColWinner(t *testing.T) {
	games := make([]Test[Board, maybe.Maybe[Figure]], 0, 3)
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
	games := make([]Test[Board, maybe.Maybe[Figure]], 0, 3)
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
	games := make([]Test[Board, maybe.Maybe[Figure]], 0, 3)
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
