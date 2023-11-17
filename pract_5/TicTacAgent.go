package main

import (
	"fmt"
)

const (
	DEBUG bool = true
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

func (b *Game) Set(p Place, fig Figure) {
	b.Board[p.Row][p.Col] = fig
}

func NewGame(size uint) Game {
	b := make([][]Cell, size)
	for i := range b {
		b[i] = make([]Cell, size)
		for j := range b[i] {
			b[i][j] = Empty{}
		}
	}
	return Game{size, b}
}

func (b Game) String() string {
	rs := fmt.Sprintf("Game [%vx%v]", b.Size, b.Size)
	rs += "\n"
	//hDelim := Reduce(func(s string, _ []Cell) string { return s + "--" }, b.Game, "")
	//rs += hDelim
	//rs += "\n"
	for row := range b.Board {
		for col := range b.Board[row] {
			rs += fmt.Sprintf("|%v", b.Board[row][col])
		}
		rs += "|\n"
	}
	//rs += hDelim
	rs += "\n"
	return rs
}

// Common functions

func Step(b Game, fig Figure) []Game {
	if DEBUG {
		Duration(Track("Step"))
	}
	bs := make([]Game, 0, b.Size*b.Size)
	for i, row := range b.Board {
		for j := range row {
			if row[j].IsEmpty() {
				newB := NewGame(3)
				newB.Copy(b)
				newB.Board[i][j] = fig
				bs = append(bs, newB)
			}
		}
	}
	return bs
}

// Main
func main() {
	B := NewGame(3)
	B.Set(Place{1, 1}, X)
	//B.Set(Place{1, 0}, O)
	fmt.Println(B)
	for _, b := range Step(B, X) {
		fmt.Println(b)
	}
}
