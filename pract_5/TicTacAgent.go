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

type Game struct {
	Size  uint
	Board [][]Cell
}

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

// Game methods

func (g Game) IsEnd() (bool, Maybe[Figure]) {
	byRow := func() Maybe[Figure] {
		// по строкам
		for _, row := range g.Board {
			f := row[0]
			if f.IsEmpty() {
				continue
			}
			for j := range row[1:] {
				if row[j] != f {
					break
				}
				return f
			}
		}
		return nil
	}()
	if byRow != nil {
		return true, byRow
	}
	return false, nil
}

func (g Game) BoardPlaces() []Place {
	size := g.Size
	ps := make([]Place, size*size)
	for i := range ps {
		ps[i] = Place{Row: uint(uint(i) / size), Col: uint(uint(i) % size)}
	}
	return ps
}

func (g *Game) Copy(g1 Game) error {
	if g.Size != g1.Size {
		return fmt.Errorf("Размеры доcок не совпадают")
	}
	for i := range g.Board {
		for j := range g.Board[i] {
			g.Board[i][j] = g1.Board[i][j]
		}
	}
	return nil
}

// Безусловно заменяет фигуру в ячейке (ячейка может быть перезаписана)
func (g *Game) Set(p Place, fig Figure) error {
	if !slices.Contains(g.BoardPlaces(), p) {
		return fmt.Errorf("Некорректный ход: адрес ячейки задан неверно")
	}
	g.Board[p.Row][p.Col] = fig
	return nil
}

func (b Board) Clear() {
	for i := range b {
		for j := range b[i] {
			b[i][j] = Empty{}
		}
	}
}

func NewGame(size uint) Game {
	b := Board(NewMatrix[Cell](size, size))
	b.Clear()
	return Game{size, b}
}

func (g Game) String() string {
	rs := fmt.Sprintf("Game [%vx%v]", g.Size, g.Size)
	rs += "\n"
	//hDelim := Reduce(func(s string, _ []Cell) string { return s + "--" }, g.Game, "")
	for row := range g.Board {
		for col := range g.Board[row] {
			rs += fmt.Sprintf("|%v", g.Board[row][col])
		}
		rs += "|\n"
	}
	rs += "\n"
	return rs
}

// Common functions

// Корректный ход в игре. Ячейка не может быть перезаписана
func Step(g Game, fig Figure, p Place) Maybe[Game] {
	if DEBUG {
		Duration(Track("Step"))
	}
	newG := NewGame(g.Size)
	newG.Copy(g)
	if !slices.Contains(g.BoardPlaces(), p) {
		return fmt.Errorf("Некорректный ход: адрес ячейки задан неверно")
	}
	if !g.Board[p.Row][p.Col].IsEmpty() {
		return fmt.Errorf("Некорректный ход: ячейка %v занята", p)
	}
	newG.Set(p, fig)
	return newG
}

func Steps(g Game, fig Figure) []Maybe[Game] {
	if DEBUG {
		Duration(Track("Steps"))
	}
	return Map(func(p Place) Maybe[Game] { g := Step(g, fig, p); return g }, g.BoardPlaces())
}

// Main
func main() {
	fmt.Println(NewMatrix[Empty](3, 3))
	G := NewGame(3)
	fmt.Println(G)
	err := G.Set(Place{1, 1}, X)
	if err != nil {
		panic(err)
	}
	err = G.Set(Place{1, 2}, O)
	if err != nil {
		panic(err)
	}
	fmt.Println(G)
	G.Board = TransposeOpt[Cell](G.Board)
	fmt.Println(G)

	//for _, g := range Steps(G, X) {
	//	fmt.Println(g)
	//}
}
