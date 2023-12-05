package main

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"log"
	"math"
	"slices"

	"github.com/denilai/maybe"
)

type LogLvl int

const (
	ERROR LogLvl = iota
	INFO
	DEBUG
)
const Gamer Figure = X

var ll = INFO

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
	Int() int
}

type Empty struct{}

type Board [][]Cell

// Empty methods

func (Empty) IsEmpty() bool  { return true }
func (Empty) String() string { return " " }
func (Empty) Int() int       { return 0 }

// Figure methods

func (fig Figure) Int() int {
	switch fig {
	case X:
		return 1
	case O:
		return 2
	default:
		panic("Некорректная фигура для игры. Ожидалось (X|O)")
	}
}
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

// Функция возвращает фигуру, за которую нужно сделать следующий шаг на доске. Gamer ходит первым
func (b Board) NextMove() Figure {
	cells := b.Flatten()
	gamerMoves := Filter(func(c Cell) bool { return c.String() == Gamer.String() }, cells)
	opponentMoves := Filter(func(c Cell) bool { return c.String() == Next(Gamer).String() }, cells)
	if len(gamerMoves) > len(opponentMoves) {
		return Next(Gamer)
	} else {
		return Gamer
	}
}

// Функция возвращает `true`, если положение фигур на поле удовлетворяет правилам игры `Крестики-нолики`
// `false` в ином случае
func (b Board) IsValid() bool {
	cells := b.Flatten()
	xs := Filter(func(c Cell) bool { return c.String() == "X" }, cells)
	os := Filter(func(c Cell) bool { return c.String() == "O" }, cells)
	lxs, los := len(xs), len(os)
	if lxs-los > 1 || lxs-los < 0 {
		return false
	}
	return true
}

// Инициализация поля для иры в Крестики-нолики с помощью списка int. См. Cell.Int
func Initialize(xs []int) Board {
	size := math.Sqrt(float64(len(xs)))
	b := NewBoard(uint(size))
	for i := range b {
		for j := range b[i] {
			switch xs[int(float64(i)*size)+j] {
			case X.Int():
				b[i][j] = X
			case O.Int():
				b[i][j] = O
			case Empty{}.Int():
				b[i][j] = Empty{}
			}
		}
	}
	return b
}

// Возвращает одномерный список ячеек
func (b Board) Flatten() []Cell {
	cs := make([]Cell, 0, b.Size()*b.Size())
	for _, row := range b {
		for _, cell := range row {
			cs = append(cs, cell)
		}
	}
	return cs
}

func (b Board) Size() int { return len(b[0]) }

// Возвращает Just(f), если срез заполнен одинаковой Figure, иначе Nothing
func IsRepeated(xs []Cell) maybe.Maybe[Figure] {
	if f, ok := xs[0].(Figure); ok && All(func(c Cell) bool { return c == xs[0] }, xs) {
		return maybe.Just(f)
	} else {
		return maybe.Nothing[Figure]()
	}
}

// Определяет победителя партии в Крестики-нолики по строкам. При отсутствии победителя возвращает maybe.Nothing
func RowWinner(b Board) maybe.Maybe[Figure] {
	if ll == DEBUG {
		Duration(Track("RowWinner"))
	}
	var winner maybe.Maybe[Figure]
	// по строкам
	rowsCheck := Map(IsRepeated, b)
	for i := range rowsCheck {
		if rowsCheck[i].HasValue() {
			winner = rowsCheck[i]
			if ll == DEBUG {
				log.Printf("%28v: %v\n", "Check by row(col):", winner)
			}
			return winner
		}
	}
	winner = maybe.Nothing[Figure]()
	if ll == DEBUG {
		log.Printf("%28v: %v\n", "Check by row(col):", winner)
	}
	return winner
}

// Определяет победителя партии в Крестики-нолики по столбцам. При отсутствии победителя возвращает maybe.Nothing
func ColWinner(b Board) maybe.Maybe[Figure] {
	if ll == DEBUG {
		Duration(Track("ColWinner"))
	}
	return RowWinner(Board(Transpose(b)))
}

// Определяет победителя партии в Крестики-нолики по главной диагонали. При отсутствии победителя возвращает maybe.Nothing
func MainDiagWinner(b Board) maybe.Maybe[Figure] {
	if ll == DEBUG {
		Duration(Track("MainDiagWinner"))
	}
	diag := make([]Cell, b.Size())
	for i := range b {
		diag[i] = b[i][i]
	}
	winner := IsRepeated(diag)
	if ll == DEBUG {
		log.Printf("%28v: %v\n", "Check by main diag:", winner)
	}
	return winner
}

// Определяет победителя партии в Крестики-нолики по вторичной диагонали. При отсутствии победителя возвращает maybe.Nothing
func SecDiagWinner(b Board) maybe.Maybe[Figure] {
	if ll == DEBUG {
		Duration(Track("SecDiagWinner"))
	}
	diag := make([]Cell, b.Size())
	for i := range b {
		diag[i] = b[i][b.Size()-1-i]
	}
	winner := IsRepeated(diag)
	if ll == DEBUG {
		log.Printf("%28v: %v\n", "Check by sec diag:", winner)
	}
	return winner
}

// Определяет победителя партии в Крестики-нолики. При отсутствии победителя возвращает maybe.Nothing
func Winner(b Board) maybe.Maybe[Figure] {
	if ll == DEBUG {
		Duration(Track("Winner"))
	}
	fs := []func(Board) maybe.Maybe[Figure]{RowWinner, ColWinner, MainDiagWinner, SecDiagWinner}
	for _, f := range fs {
		if w := f(b); w != maybe.Nothing[Figure]() {
			return w
		}
	}
	return maybe.Nothing[Figure]()
}

// Функция возвращает все свободные (незанятые фигурами) места на поля для игры в Крестики-нолики
func (b Board) EmptyPlaces() []Place {
	size := b.Size()
	ps := make([]Place, 0, size*size)
	for i := range ps {
		p := Place{Row: uint(i / size), Col: uint(i % size)}
		if cell, err := b.Get(p); err == nil && !cell.IsEmpty() {
			ps = append(ps, Place{Row: uint(i / size), Col: uint(i % size)})
		}
	}
	return ps
}

// Функция возвращает все места на поля для игры в Крестики-нолики
func (b Board) Places() []Place {
	size := b.Size()
	ps := make([]Place, size*size)
	for i := range ps {
		ps[i] = Place{Row: uint(i / size), Col: uint(i % size)}
	}
	return ps
}

// Функция копирования поля
func CopyBoard(dst, src Board) error {
	if dst.Size() != src.Size() {
		return fmt.Errorf("Размеры доcок не совпадают")
	}
	for i := range src {
		copy(dst[i], src[i])
	}
	return nil
}

// Получение значения ячейки поля для игры в Крестики-нолики по меcту
func (b Board) Get(p Place) (Cell, error) {
	if !slices.Contains(b.Places(), p) {
		return nil, fmt.Errorf("Некорректный ход: адрес ячейки задан неверно")
	}
	return b[p.Row][p.Col], nil
}

// Безусловно заменяет фигуру в ячейке поля для игры в Крестики-нолики (ячейка может быть перезаписана)
func (b *Board) Set(p Place, cell Cell) error {
	if !slices.Contains(b.Places(), p) {
		return fmt.Errorf("Некорректный ход: адрес ячейки задан неверно")
	}
	(*b)[p.Row][p.Col] = cell
	return nil
}

// Очищает поле для игры в Крестики-нолики
func (b Board) Clear() {
	for i := range b {
		for j := range b[i] {
			b.Set(Place{Row: uint(i), Col: uint(j)}, Empty{})
		}
	}
}

// Создание нового объекта типа Board (для игры в Крестики-нолики) заданной размерности
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

func Encode(t Board) (string, error) {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	if err := e.Encode(Map(func(c Cell) int { return c.Int() }, t.Flatten())); err != nil {
		return "", fmt.Errorf("Failed gob Encode: %v", err)
	}
	return base64.StdEncoding.EncodeToString(b.Bytes()), nil
}

func Decode(src string) (Board, error) {
	obj := *new([]int)
	if by, err := base64.StdEncoding.DecodeString(src); err != nil {
		return *new(Board), fmt.Errorf("Failed base64 Decode: %v", err)
	} else {
		b := bytes.Buffer{}
		b.Write(by)
		d := gob.NewDecoder(&b)
		if err := d.Decode(&obj); err != nil {
			return *new(Board), fmt.Errorf("Failed gob Decode: %v", err)
		} else {
			return Initialize(obj), nil
		}
	}
}

// Корректный ход в партии игры Крестики-нолики. Ячейка не может быть перезаписана
func Step(b Board, fig Figure, p Place) maybe.Maybe[Board] {
	if ll == DEBUG {
		Duration(Track("Step"))
	}
	if ll == INFO {
		log.Printf("Ход %v на %v", fig, p)
	}
	if Winner(b).HasValue() {
		if ll == INFO {
			fmt.Println("Партия окончена")
		}
		return maybe.Nothing[Board]()
	}
	if !slices.Contains(b.Places(), p) {
		if ll == INFO {
			fmt.Println("Некорректный ход: адрес ячейки задан неверно")
		}
		return maybe.Nothing[Board]()
	}
	newG := NewBoard(uint(b.Size()))
	CopyBoard(newG, b)
	if cell, err := b.Get(p); err != nil {
		panic(err)
	} else if !cell.IsEmpty() {
		if ll == INFO {
			fmt.Printf("Некорректный ход: ячейка %v занята\n", p)
		}
		return maybe.Nothing[Board]() //
	}
	if err := newG.Set(p, fig); err != nil {
		panic(err)
	}
	if ll == INFO {
		log.Printf("Совершённый ход %v на %v\n", fig, p)
		log.Println(newG)
	}
	return maybe.Just(newG)
}

func RecSteps2(scoreMap map[string]float64, boards []Board) map[string]float64 {
	if ll == DEBUG {
		Duration(Track("RecSteps2"))
	}
	if len(boards) == 0 {
		return scoreMap
	}
	ini := boards[0]
	code, err := Encode(ini)
	if err != nil {
		panic(err)
	}
	if ll == INFO {
		log.Printf("Текущее поле: %v", ini)
	}
	if _, ok := scoreMap[code]; !ok {
		if ll == INFO {
			log.Println("Новая комбинация")
		}
		if !Winner(ini).HasValue() {
			if ll == INFO {
				log.Println("Коэффцициент = 0.5")
			}
			scoreMap[code] = 0.5
			childrenBoards := Map(func(mb maybe.Maybe[Board]) Board { return mb.FromJust() }, Filter(func(mb maybe.Maybe[Board]) bool { return mb.HasValue() }, Map(func(p Place) maybe.Maybe[Board] { return Step(ini, ini.NextMove(), p) }, ini.Places())))
			boards = append(boards, childrenBoards...)
		} else {
			if Winner(ini) == maybe.Just[Figure](Gamer) {
				if ll == INFO {
					log.Printf("Мы (%v) победили. Коэффцициент = 1\n", Gamer)
				}
				scoreMap[code] = 1
			}
			if Winner(ini) == maybe.Just[Figure](Next(Gamer)) {
				if ll == INFO {
					log.Printf("Мы (%v) проиграли. Коэффцициент = 0\n", Gamer)
				}
				scoreMap[code] = 0
			}
			if ll == INFO {
				log.Println("Игра окончена. Следующие комбинации не будут рассматриваться")
			}
		}
	} else {
		if ll == INFO {
			log.Printf("Повторная комбинация. Пропуск")
		}
	}
	return RecSteps2(scoreMap, boards[1:])
}

func Steps(b Board, fig Figure) []maybe.Maybe[Board] {
	if ll == DEBUG {
		Duration(Track("Steps"))
	}
	return Map(func(p Place) maybe.Maybe[Board] { b := Step(b, fig, p); return b }, b.Places())
}

func analyzeScoreMap(scoreMap map[string]float64) {
	var win, loose int
	for _, v := range scoreMap {
		if v == 1 {
			win += 1
		}
		if v == 0 {
			loose += 1
		}
	}
	fmt.Printf("Размер мапы : %v\n", len(scoreMap))
	fmt.Printf("Выигрышей   : %v\n", win)
	fmt.Printf("Проигрышей  : %v\n", loose)
}

func showMap(sm map[string]float64) {
	for k, v := range sm {
		b, err := Decode(k)
		if err != nil {
			panic(err)
		} else {
			fmt.Println(b)
			fmt.Println(v)
		}
	}
}

type Agent struct {
	qmap   map[string]float64
	figure Figure
}

func (a Agent) Lookup(b Board) float64 {
	code, err := Encode(b)
	if err != nil {
		panic(err)
	}
	value, ok := a.qmap[code]
	if !ok {
		panic(fmt.Errorf("%v\nКомбинация не найдена в матрице ценностей", b))
	}
	return value
}

func (a Agent) MakeMove(b Board) Board {
	value := a.Lookup(b)
	emptyCells := b.EmptyPlaces()
	_, _ = value, emptyCells
	candidates := Map(func(mb maybe.Maybe[Board]) Board { return mb.FromJust() }, Filter(func(mb maybe.Maybe[Board]) bool { return mb.HasValue() }, Map(func(p Place) maybe.Maybe[Board] { return Step(b, a.figure, p) }, emptyCells)))
	values := Map(a.Lookup, candidates)
	maxV := slices.Max(values)
	for i, v := range values {
		if v == maxV {
			return candidates[i]
		}
	}
	return b
}

func Match() {
	board := NewBoard(3)
	agent := X
	opponent := O

}

// Main
func main() {
	G1 := NewBoard(3)
	bs := RecSteps2(make(map[string]float64, int(math.Pow(3, 9))), []Board{G1})

	fmt.Println(len(bs))
	analyzeScoreMap(bs)
	//showMap(bs)
}
