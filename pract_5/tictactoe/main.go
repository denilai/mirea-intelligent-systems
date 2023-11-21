package main

import (
	"fmt"
	"log"
	"slices"

	"github.com/denilai/maybe"
)

const (
	DEBUG bool = false
)

type Figure byte

const (
	O Figure = iota
	X
	FigureCount
)

type Place struct {
	Row, Col uint
}

type Cell interface {
	IsEmpty() bool
	String() string
}

type Empty struct{}

type Board [][]Cell

// Empty methods

func (Empty) IsEmpty() bool  { return true }
func (Empty) String() string { return " " }

// Figure methods
func Next(fig Figure) Figure {
	switch fig {
	case X:
		return O
	case O:
		return X
	default:
		panic("Некорректная фигура для игры. Ожидалось (X|O)")
	}

}

func (fig Figure) IsEmpty() bool { return false }

func (fig Figure) String() string {
	switch fig {
	case X:
		return "X"
	case O:
		return "O"
	default:
		return "N/A"
	}
}

// Board methods
func (b Board) Size() int { return len(b[0]) }

// Возвращает Just(f), если срез заполнен одинаковой Figure, иначе Nothing
func IsRepeated(xs []Cell) maybe.Maybe[Figure] {
	if f, ok := xs[0].(Figure); ok && All(func(c Cell) bool { return c == xs[0] }, xs) {
		return maybe.Just(f)
	} else {
		return maybe.Nothing[Figure]()
	}
}

func RowWinner(b Board) maybe.Maybe[Figure] {
	if DEBUG {
		Duration(Track("RowWinner"))
	}
	// по строкам
	rowsCheck := Map(IsRepeated, b)
	for i := range rowsCheck {
		if rowsCheck[i].HasValue() {
			return rowsCheck[i]
		}
	}
	return maybe.Nothing[Figure]()
}

func ColWinner(b Board) maybe.Maybe[Figure] {
	if DEBUG {
		Duration(Track("ColWinner"))
	}
	return RowWinner(Board(Transpose(b)))
}

func MainDiagWinner(b Board) maybe.Maybe[Figure] {
	if DEBUG {
		Duration(Track("MainDiagWinner"))
	}
	diag := make([]Cell, b.Size())
	for i := range b {
		diag[i] = b[i][i]
	}
	return IsRepeated(diag)
}

func SecDiagWinner(b Board) maybe.Maybe[Figure] {
	if DEBUG {
		Duration(Track("SecDiagWinner"))
	}
	diag := make([]Cell, b.Size())
	for i := range b {
		diag[i] = b[i][b.Size()-1-i]
	}
	return IsRepeated(diag)
}

func Winner(b Board) maybe.Maybe[Figure] {
	if DEBUG {
		Duration(Track("Winner"))
	}
	rowWinner, colWinner, mainDiagWinner, secDiagWinner := RowWinner(b), ColWinner(b), MainDiagWinner(b), SecDiagWinner(b)
	winners := [4]maybe.Maybe[Figure]{rowWinner, colWinner, mainDiagWinner, secDiagWinner}
	if DEBUG {
		log.Printf("%28v: %v\n", "Check by row:", rowWinner)
		log.Printf("%28v: %v\n", "Check by col:", colWinner)
		log.Printf("%28v: %v\n", "Check by main diagonal:", mainDiagWinner)
		log.Printf("%28v: %v\n", "Check by secondary diagonal:", secDiagWinner)
	}
	for _, w := range winners {
		if w != maybe.Nothing[Figure]() {
			return w
		}
	}
	return maybe.Nothing[Figure]()
}

func (b Board) BoardPlaces() []Place {
	size := b.Size()
	ps := make([]Place, size*size)
	for i := range ps {
		ps[i] = Place{Row: uint(i / size), Col: uint(i % size)}
	}
	return ps
}

func CopyBoard(dst, src Board) error {
	if dst.Size() != src.Size() {
		return fmt.Errorf("Размеры доcок не совпадают")
	}
	for i := range src {
		copy(dst[i], src[i])
	}
	return nil

	//	for i := range src {
	//		for j := range src[i] {
	//			if el, errGet := src.Get(Place{Row: uint(i), Col: uint(j)}); errGet != nil {
	//				return errGet
	//			} else if errSet := dst.Set(Place{Row: uint(i), Col: uint(j)}, el); errSet != nil {
	//				return errSet
	//			}
	//		}
	//	}
	//
	// return nil
}

func (b Board) Get(p Place) (Cell, error) {
	if !slices.Contains(b.BoardPlaces(), p) {
		return nil, fmt.Errorf("Некорректный ход: адрес ячейки задан неверно")
	}
	return b[p.Row][p.Col], nil
}

// Безусловно заменяет фигуру в ячейке (ячейка может быть перезаписана)
func (b *Board) Set(p Place, cell Cell) error {
	if !slices.Contains(b.BoardPlaces(), p) {
		return fmt.Errorf("Некорректный ход: адрес ячейки задан неверно")
	}
	(*b)[p.Row][p.Col] = cell
	return nil
}

func (b Board) Clear() {
	for i := range b {
		for j := range b[i] {
			b.Set(Place{Row: uint(i), Col: uint(j)}, Empty{})
		}
	}
}

func NewBoard(size uint) Board {
	b := Board(NewMatrix[Cell](size, size))
	b.Clear()
	return b
}

func (b Board) String() string {
	rs := fmt.Sprintf("Board [%vx%v]", b.Size(), b.Size())
	//rs += "\n|"
	//hDelim := Reduce(func(s string, _ []Cell) string { return s + "--" }, b, "")
	for _, row := range b {
		rs += "\n|"
		for _, cell := range row {
			rs += fmt.Sprintf("%v|", cell)
		}
	}
	return rs
}

// Common functions

// Корректный ход в игре. Ячейка не может быть перезаписана
func Step(mb maybe.Maybe[Board], fig Figure, p Place) maybe.Maybe[Board] {
	if DEBUG {
		Duration(Track("Step"))
	}
	if DEBUG {
		log.Println(mb)
		log.Printf("Ход %v на %v", fig, p)
	}
	if !mb.HasValue() {
		if DEBUG {
			fmt.Println("Поле не существует (Nohting)")
		}
		return maybe.Nothing[Board]()
	}
	b := mb.FromJust()
	if Winner(b).HasValue() {
		if DEBUG {
			fmt.Println("Партия окончена")
		}
		return maybe.Nothing[Board]()
	}
	newG := NewBoard(uint(b.Size()))
	CopyBoard(newG, b)
	if !slices.Contains(b.BoardPlaces(), p) {
		if DEBUG {
			fmt.Println("Некорректный ход: адрес ячейки задан неверно")
		}
		return maybe.Nothing[Board]()
	}
	if cell, err := b.Get(p); err != nil {
		panic(err)
	} else if !cell.IsEmpty() {
		if DEBUG {
			fmt.Printf("Некорректный ход: ячейка %v занята\n", p)
		}
		return maybe.Nothing[Board]() //
	}
	if err := newG.Set(p, fig); err != nil {
		return maybe.Nothing[Board]()
	}
	return maybe.Just(newG)
}

func RecSteps(mb maybe.Maybe[Board], fig Figure) []maybe.Maybe[Board] {
	if !mb.HasValue() {
		return *(new([]maybe.Maybe[Board]))
	} else {
		b := mb.FromJust()
		// TODO -- подобрать оптимальный размер среза
		res := make([]maybe.Maybe[Board], 0)
		boards := Map(func(p Place) maybe.Maybe[Board] { return Step(mb, fig, p) }, b.BoardPlaces())
		//return boards
		for _, b := range boards {
			res = append(res, b)
			res = append(res, RecSteps(b, Next(fig))...)
		}
		return res
	}
	//return *(new([]maybe.Maybe[Board]))
}

func Steps(b Board, fig Figure) []maybe.Maybe[Board] {
	if DEBUG {
		Duration(Track("Steps"))
	}
	return Map(func(p Place) maybe.Maybe[Board] { b := Step(maybe.Just(b), fig, p); return b }, b.BoardPlaces())
}

// Main
func main() {
	G1 := Board{{Empty{}, Empty{}, Empty{}}, {Empty{}, Empty{}, Empty{}}, {Empty{}, Empty{}, Empty{}}}
	bs := RecSteps(maybe.Just(G1), X)
	fbs := Filter(func(mb maybe.Maybe[Board]) bool { return mb.HasValue() }, bs)
	//fmt.Println(fbs)
	fmt.Println(len(fbs))

}
