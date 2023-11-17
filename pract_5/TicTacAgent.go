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

func (g *Game) Set(p Place, fig Figure) {
	g.Board[p.Row][p.Col] = fig
}

func NewGame(size uint) Game {
	g := make([][]Cell, size)
	for i := range g {
		g[i] = make([]Cell, size)
		for j := range g[i] {
			g[i][j] = Empty{}
		}
	}
	return Game{size, g}
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

func Step(g Game, fig Figure, p Place) Game {
	if DEBUG {
		Duration(Track("Step"))
	}
	newG := NewGame(g.Size)
	newG.Copy(g)
	newG.Board[p.Row][p.Col] = fig
	return newG
}

func Steps(g Game, fig Figure) []Game {
	if DEBUG {
		Duration(Track("Steps"))
	}
	return Map(func(p Place) Game { return Step(g, fig, p) }, g.BoardPlaces())
}

// Main
func main() {
	G := NewGame(3)
	G.Set(Place{1, 1}, X)
	fmt.Println(G)
	for _, g := range StepsM(G, X) {
		fmt.Println(g)
	}
}
