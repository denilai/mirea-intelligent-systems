package main

import (
	"image/color"
	"log"
	"math"
	"math/rand"
	"slices"
	"time"

	"github.com/fogleman/gg"
)

// Управляющие параметры
const (
	// величина измерение поля
	FIELD_SIZE uint = 300
	// процент заполнения ячеек на поле
	FILLING_THRESHOLD float64 = 0.9
	// размер одной ячейки в пикселях
	CELL_SIZE_PX uint = 10
)

// Empty: color.RGBA{252, 255, 231, 255}}
var COLOR_MODEL = map[Color]color.RGBA{Blue: color.RGBA{43, 51, 103, 255}, Red: color.RGBA{235, 69, 95, 255}}

// Тип для цвета ячейки
type Color uint8

const (
	Empty Color = iota
	Red
	Blue
	ColorLen
)

type Addr struct {
	X, Y uint
}

type Tenant struct {
	Tag   Color
	Happy bool
	Addr
}

// квадратное поле для проведения эксперимента Шеллинга
type SField struct {
	Size uint
	Grid [][]Tenant
}

// Inhabit (англ.) -- населить
func (sf *SField) Inhabit(tns []Tenant) {
	for _, t := range tns {
		sf.Grid[t.Y][t.X] = t
	}
}

func NewSField() SField {
	grid := make([][]Tenant, FIELD_SIZE)
	for i := range grid {
		grid[i] = make([]Tenant, FIELD_SIZE)
	}
	return SField{FIELD_SIZE, grid}
}

// Создать жильцов и выдать им адреса (Addr)
func RegisterTenats() []Tenant {
	numberOfCells := int(math.Pow(float64(FIELD_SIZE), 2))
	log.Printf("Количество ячеек = %v", numberOfCells)
	numberOfFreeCells := int(float64(numberOfCells) * FILLING_THRESHOLD)
	log.Printf("Количество пригодных для заселения ячеек = %v, что составляет %f%", numberOfFreeCells, float64(numberOfFreeCells)/float64(numberOfCells)*100)
	residends := make([]Tenant, numberOfFreeCells)
	addresses := make([]Addr, numberOfFreeCells)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < numberOfFreeCells; i++ {
		randX, randY := uint(r.Intn(int(FIELD_SIZE))), uint(r.Intn(int(FIELD_SIZE)))
		addr := Addr{randX, randY}
		if !slices.Contains(addresses, addr) {
			addresses = append(addresses, addr)
			residends = append(residends, Tenant{Color(r.Intn(int(FIELD_SIZE))), false, addr})
		}
	}
	return residends

}

func (sf SField) Render(filename string) {
	ctx := gg.NewContext(int(FIELD_SIZE*CELL_SIZE_PX), int(FIELD_SIZE*CELL_SIZE_PX))
	for y, row := range sf.Grid {
		for x, tnt := range row {
			ctx.DrawRectangle(float64(int(CELL_SIZE_PX)*x), float64(int(CELL_SIZE_PX)*y), float64(CELL_SIZE_PX), float64(CELL_SIZE_PX))
			color := COLOR_MODEL[tnt.Tag]
			ctx.SetColor(color)
			ctx.Fill()
		}
	}
	if er := ctx.SavePNG(filename); er != nil {
		log.Println("Error in SavePNG")
	}
}

/*
func (sf SField) CountOfNeighbours(x, y uint) int {
	count := 0
	myColor := sf.Grid[x][y].Color
	intX, intY := int(x), int(y)
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if intY+dy < 0 || intX+dx < 0 || intX+dx >= len(sf.Grid[y]) || intY+dy >= len(sf.Grid) {
				continue
			}
			if sf.Grid[intY+dy][intX+dx].Color == myColor {
				count += 1
			}
		}
	}
	return count
}
*/

func filter[T any](data []T, f func(T) bool) []T {

	fltd := make([]T, 0, len(data))

	for _, e := range data {
		if f(e) {
			fltd = append(fltd, e)
		}
	}

	return fltd
}

//func ShuffleUnhappyCells(sf SField) {
//	unhappyCells := make([]Cell, 0)
//	for _, row := range sf.Grid {
//		unhappyCells = append(unhappyCells, filter(row, func(cell Cell) bool { return !cell.Happy })...)
//	}
//	unhappyPos
//}

func main() {
	filename := "grid.png"
	tenants := RegisterTenats()
	sf := NewSField()
	sf.Inhabit(tenants)
	sf.Render(filename)

	// RenderSF(sf, "field.png")
}
