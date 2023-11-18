package main

import (
	"fmt"
	"slices"
)

const (
	DEBUG bool = true
)

type Maybe[T any] interface{}

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

//func check1(b Board) Maybe[Figure] {
//	// по строкам
//	for _, row := range b.GetBoard() {
//		f := row[0]
//		if f.IsEmpty() {
//			continue
//		}
//		for j := range row[1:] {
//			if row[j] != f {
//				break
//			}
//			return f
//		}
//	}
//	return nil
//}

//func check2(b Board) Maybe[Figure] {
//	tspB := TransposeOpt(b)
//
//	b.SetBoard(TransposeOpt(b.GetBoard()))
//	byCol
//
//	return false, byRow
//}

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
	rs += "\n"
	//hDelim := Reduce(func(s string, _ []Cell) string { return s + "--" }, b.Board, "")
	for _, row := range b {
		for _, cell := range row {
			rs += fmt.Sprintf("|%v", cell)
		}
		rs += "|\n"
	}
	rs += "\n"
	return rs
}

// Common functions

// Корректный ход в игре. Ячейка не может быть перезаписана
func Step(b Board, fig Figure, p Place) Maybe[Board] {
	if DEBUG {
		Duration(Track("Step"))
	}
	newG := NewBoard(uint(b.Size()))
	CopyBoard(newG, b)
	if !slices.Contains(b.BoardPlaces(), p) {
		return fmt.Errorf("Некорректный ход: адрес ячейки задан неверно")
	}
	if cell, err := b.Get(p); err != nil {
		panic(err)
	} else if cell.IsEmpty() {
		return fmt.Errorf("Некорректный ход: ячейка %v занята", p)
	}
	if err := newG.Set(p, fig); err != nil {
		return err
	}
	return newG
}

func Steps(b Board, fig Figure) []Maybe[Board] {
	if DEBUG {
		Duration(Track("Steps"))
	}
	return Map(func(p Place) Maybe[Board] { b := Step(b, fig, p); return b }, b.BoardPlaces())
}

// Main
func main() {
	G := NewBoard(3)
	fmt.Println(G)
	if err := G.Set(Place{1, 1}, X); err != nil {
		panic(err)
	}
	if err := G.Set(Place{1, 2}, O); err != nil {
		panic(err)
	}
	fmt.Println(G)
	G = TransposeOpt(G)
	fmt.Println(G)

	//for _, b := range Steps(G, X) {
	//	fmt.Println(b)
	//}
}
