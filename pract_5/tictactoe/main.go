package main

import (
	"fmt"
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
	return RowWinner(Board(TransposeOpt(b)))
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
	fmt.Println("Board check")
	fmt.Println(b)
	fmt.Println("Find winner")
	fmt.Printf("%28v: %v\n", "Check by row", RowWinner(b))
	fmt.Printf("%28v: %v\n", "Check by col", ColWinner(b))
	fmt.Printf("%28v: %v\n", "Check by main diagonal", MainDiagWinner(b))
	fmt.Printf("%28v: %v\n", "Check by secondary diagonal", SecDiagWinner(b))

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
		dst[i] = src[i]
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
	//hDelim := Reduce(func(s string, _ []Cell) string { return s + "--" }, b.Board, "")
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
func Step(b Board, fig Figure, p Place) maybe.Maybe[Board] {
	if DEBUG {
		Duration(Track("Step"))
	}
	newG := NewBoard(uint(b.Size()))
	CopyBoard(newG, b)
	if !slices.Contains(b.BoardPlaces(), p) {
		return maybe.Nothing[Board]() //fmt.Errorf("Некорректный ход: адрес ячейки задан неверно")
	}
	if cell, err := b.Get(p); err != nil {
		panic(err)
	} else if cell.IsEmpty() {
		return maybe.Nothing[Board]() //fmt.Errorf("Некорректный ход: ячейка %v занята", p)
	}
	if err := newG.Set(p, fig); err != nil {
		return maybe.Nothing[Board]()
	}
	return maybe.Just(newG)
}

func Steps(b Board, fig Figure) []maybe.Maybe[Board] {
	if DEBUG {
		Duration(Track("Steps"))
	}
	return Map(func(p Place) maybe.Maybe[Board] { b := Step(b, fig, p); return b }, b.BoardPlaces())
}

// Main
func main() {
	G1 := NewBoard(3)
	if err := G1.Set(Place{0, 0}, X); err != nil {
		panic(err)
	}
	if err := G1.Set(Place{1, 0}, X); err != nil {
		panic(err)
	}
	if err := G1.Set(Place{2, 0}, X); err != nil {
		panic(err)
	}
	G1r := Board(TransposeOpt(G1))
	G2 := NewBoard(3)
	if err := G2.Set(Place{0, 1}, X); err != nil {
		panic(err)
	}
	if err := G2.Set(Place{1, 1}, X); err != nil {
		panic(err)
	}
	if err := G2.Set(Place{2, 1}, X); err != nil {
		panic(err)
	}
	G2r := Board(TransposeOpt(G2))
	G3 := NewBoard(3)
	if err := G3.Set(Place{0, 2}, O); err != nil {
		panic(err)
	}
	if err := G3.Set(Place{1, 2}, O); err != nil {
		panic(err)
	}
	if err := G3.Set(Place{2, 2}, O); err != nil {
		panic(err)
	}
	G3r := Board(TransposeOpt(G3))
	G4 := NewBoard(3)
	if err := G4.Set(Place{0, 0}, O); err != nil {
		panic(err)
	}
	if err := G4.Set(Place{1, 1}, O); err != nil {
		panic(err)
	}
	if err := G4.Set(Place{2, 2}, O); err != nil {
		panic(err)
	}
	G5 := NewBoard(3)
	if err := G5.Set(Place{0, 2}, O); err != nil {
		panic(err)
	}
	if err := G5.Set(Place{1, 1}, O); err != nil {
		panic(err)
	}
	if err := G5.Set(Place{2, 0}, O); err != nil {
		panic(err)
	}
	Winner(G1)
	Winner(G1r)
	Winner(G2)
	Winner(G2r)
	Winner(G3)
	Winner(G3r)
	Winner(G4)
	Winner(G5)

	//for _, b := range Steps(G, X) {
	//	fmt.Println(b)
	//}
}
