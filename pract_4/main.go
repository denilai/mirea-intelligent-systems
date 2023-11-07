package main

import (
	"fmt"
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
	// процент заполнения ячеек на поле
	FILLING_THRESHOLD float64 = 0.7
	// размер одной ячейки в пикселях
	CELL_SIZE_PX uint = 10
	// количество итераций
	ITER_COUNT uint = 1000
	// порог счастья
	HAPPY_THRESHOLD uint = 3
	FIELD_SIZE      uint = 100
)

// Empty: color.RGBA{252, 255, 231, 255}}
var COLOR_MODEL = map[Color]color.RGBA{Blue: color.RGBA{43, 51, 103, 255}, Red: color.RGBA{235, 69, 95, 255}, White: color.RGBA{252, 255, 231, 255}}

// Тип для цвета ячейки
type Color uint8

const (
	Red Color = iota
	Blue
	White
	ColorLen
)

type Addr struct {
	Row, Col uint
}

type House interface {
	isFree() bool
	getAddr() Addr
}

type Street = []House

type Empty struct {
	Addr Addr
}

type Citizen struct {
	Tag   Color
	Happy bool
	Addr  Addr
}

func (c *Citizen) isFree() bool  { return false }
func (c Empty) isFree() bool     { return true }
func (c *Citizen) getAddr() Addr { return c.Addr }
func (e Empty) getAddr() Addr    { return e.Addr }

// квадратное поле для проведения эксперимента Шеллинга
type SField struct {
	Size uint
	Grid []Street
}

func (sf SField) Show() {
	for _, street := range sf.Grid {
		fmt.Println(street)
		//for x := range sf.Grid {
		//	fmt.Printf("Grid [%v][%v] = %v\n", y, x, sf.Grid[y][x])
		//}
	}
}

func (sf *SField) Clean() {
	for y := range sf.Grid {
		for x := range sf.Grid[y] {
			ux, uy := uint(x), uint(y)
			sf.Grid[y][x] = Empty{Addr{ux, uy}}
		}
	}

}

// Inhabit (англ.) -- населить
func (sf *SField) Inhabit(cs []Citizen) {
	defer Duration(Track("Inhabit"))
	sf.Clean()
	for i := range cs {
		sf.Grid[cs[i].Addr.Row][cs[i].Addr.Col] = &cs[i]
	}
}

func (sf SField) GetHappyCitizens() []Citizen {
	happy := make([]Citizen, 0, sf.Size)
	for _, street := range sf.Grid {
		for _, house := range street {
			c, ok := house.(*Citizen)
			if !ok {
				continue
			}
			if c.Happy {
				happy = append(happy, *c)
			}
		}
	}
	return happy
}

func (sf *SField) HappinesAssessment() {
	allHouses := sf.FlattenHouses()
	for _, house := range allHouses {
		if c, ok := house.(*Citizen); ok {
			if count, _ := sf.CountOfNeighbours(c); count >= int(HAPPY_THRESHOLD) {
				log.Printf("%v is happy!", c)
				c.Happy = true
			}
		}
	}
	//	for _, street := range sf.Grid {
	//		for _, house := range street {
	//			c, ok := house.(*Citizen)
	//			if !ok {
	//				continue
	//			}
	//			if neighbC, err := sf.CountOfNeighbours(c); err == nil && neighbC >= int(HAPPY_THRESHOLD) {
	//				log.Printf("%v is happy!", c)
	//				c.Happy = true
	//			}
	//		}
	//	}
}

func NewSField(size uint) SField {
	defer Duration(Track("NewSField"))
	grid := make([]Street, int(size))
	for y := range grid {
		grid[y] = make([]House, int(size))
		for x := range grid[y] {
			ux, uy := uint(x), uint(y)
			grid[y][x] = Empty{Addr{ux, uy}}
		}
	}
	return SField{size, grid}
}

// Создать жильцов
func CreateCitizens(size uint) []Citizen {
	defer Duration(Track("CreateCitizens"))
	numberOfCells := int(math.Pow(float64(size), 2))
	log.Printf("Количество ячеек = %v", numberOfCells)
	numberOfFreeCells := int(float64(numberOfCells) * FILLING_THRESHOLD)
	log.Printf("Количество пригодных для заселения ячеек = %v, что составляет %0.2f%%", numberOfFreeCells, float64(numberOfFreeCells)/float64(numberOfCells)*100)
	citizens := make([]Citizen, numberOfFreeCells)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range citizens {
		color := Color(r.Intn(2))
		//log.Printf("Create citizen with color: %v", color)
		citizens[i] = Citizen{color, false, Addr{}}
	}
	return citizens
}

// Выдать жильцам адреса
func RegisterCitizens(citizens []Citizen, size uint) []Citizen {
	defer Duration(Track("RegisterCitizen"))
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	addresses := make([]Addr, 0, len(citizens))
	for i := 0; i < cap(addresses); i++ {
		var randAddr Addr
		for {
			randAddr = Addr{uint(r.Intn(int(size))), uint(r.Intn(int(size)))}
			if !slices.Contains(addresses, randAddr) {
				break
			}
		}
		addresses = append(addresses, randAddr)
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(addresses), func(i, j int) { addresses[i], addresses[j] = addresses[j], addresses[i] })
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(addresses), func(i, j int) { addresses[i], addresses[j] = addresses[j], addresses[i] })
	for i := range citizens {
		citizens[i].Addr = addresses[i]
	}
	log.Printf("Количество уникальных адресов = %v\n", len(addresses))
	return citizens
}

func (sf SField) Render(filename string) {
	defer Duration(Track("Render"))
	ctx := gg.NewContext(int(sf.Size*CELL_SIZE_PX), int(sf.Size*CELL_SIZE_PX))
	for row, street := range sf.Grid {
		for col, house := range street {
			var color color.RGBA
			var tag Color
			ctx.DrawRectangle(float64(int(CELL_SIZE_PX)*col), float64(int(CELL_SIZE_PX)*row), float64(CELL_SIZE_PX), float64(CELL_SIZE_PX))
			switch ht := house.(type) {
			case *Citizen:
				tag = ht.Tag
				//log.Printf("Cell with color %v", tag)
			case Empty:
				tag = White
			default:
				log.Fatalf("Неожиданный тип: %T", ht)
			}
			color, ok := COLOR_MODEL[tag]
			if !ok {
				log.Panicf("Несуществующий цвет под тэгом: %v", tag)
			}
			//log.Printf("Выбран цвет: %v", tag)
			ctx.SetColor(color)
			ctx.Fill()
		}
	}
	log.Printf("Создаем изображение %v...", filename)
	if er := ctx.SavePNG(filename); er != nil {
		log.Println("Error in SavePNG")
	}
}

func (sf SField) CountOfNeighbours(h House) (int, error) {
	defer Duration(Track("CountOfNeighbours"))
	count := 0
	addr := h.getAddr()
	c, ok := sf.Grid[addr.Row][addr.Col].(*Citizen)
	if !ok {
		return 0, fmt.Errorf("House (cell) with addres {%v} is empty", c.Addr)
	}
	myTag := c.Tag
	intCol, intRow := int(c.Addr.Col), int(c.Addr.Row)
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if intRow+dy < 0 || intCol+dx < 0 || intCol+dx >= int(sf.Size) || intRow+dy >= int(sf.Size) || (dx == 0 && dy == 0) {
				continue
			}
			//log.Printf("dx = %v, dy = %v", dx, dy)
			if c1, ok := sf.Grid[intRow+dy][intCol+dx].(*Citizen); ok && c1.Tag == myTag {
				//log.Printf("%v -- близкий друг %v", c1, c)
				count += 1
			}
		}
	}
	//log.Printf("У жителя дома %v %v близких соседей", h, count)
	return count, nil
}

func Map[T any, M any](f func(T) M, data []T) []M {
	defer Duration(Track("Map"))
	n := make([]M, len(data))
	for i, e := range data {
		n[i] = f(e)
	}
	return n
}

func Filter[T any](f func(T) bool, data []T) []T {
	defer Duration(Track("Filter"))
	fltd := make([]T, 0, len(data))
	for _, e := range data {
		if f(e) {
			fltd = append(fltd, e)
		}
	}
	return fltd
}

func (sf SField) CountOfHouses() int {
	return int(math.Pow(float64(sf.Size), 2))
}

func (sf SField) CountOfCitizens() int {
	allHouses := sf.FlattenHouses()
	citizens := Filter(func(h House) bool { _, ok := h.(*Citizen); return ok }, allHouses)
	return len(citizens)
}

//func (sf SField) MapCitizens (f func(House) Citzen) []Citizen {
//	res := []House{}
//	for _, street := range sf.Grid {
//		for _, house := range street {
//			if c,ok:=house.(*Citizen); ok {
//				res = append(res, f(house))
//			}
//		}
//	}
//	return res
//
//}

func (sf SField) FilterHouses(f func(House) bool) []House {
	res := []House{}
	for _, street := range sf.Grid {
		for _, house := range street {
			if f(house) {
				res = append(res, house)
			}
		}
	}
	return res
}

func (sf SField) FlattenHouses() []House {
	allHouses := []House{}
	for i := range sf.Grid {
		allHouses = append(allHouses, sf.Grid[i]...)
	}
	return allHouses
}

func (sf SField) GetUnhappyCitizens() []*Citizen {
	//allHouses := []House{}
	//for i := range sf.Grid {
	//	allHouses = append(allHouses, sf.Grid[i]...)
	//}
	//fmt.Println(allHouses)
	unhappyHouses := sf.FilterHouses(func(h House) bool { c, ok := h.(*Citizen); return ok && !c.Happy })
	res := Map(func(h House) *Citizen {
		if c, ok := h.(*Citizen); ok {
			return c
		} else {
			return nil
		}
	}, unhappyHouses)
	return res
}

func (sf *SField) ShuffleUnhappyCells() {
	emptyAddrs := sf.GetEmptyAddreses()
	unhappyCitizens := sf.GetUnhappyCitizens()
	if len(unhappyCitizens) == 0 {
		log.Panic("Все счастливы")
	}
	fmt.Println("Несчаcтливые жители: %v, len = %v", unhappyCitizens, len(unhappyCitizens))
	unhappyAddrs := Map(func(c *Citizen) Addr { return c.Addr }, unhappyCitizens)

	//moveTo := emptyAddrs
	moveTo := append(emptyAddrs, unhappyAddrs...)
	fmt.Println("Места для переселения: %v, len = %v", moveTo, len(moveTo))
	if len(unhappyCitizens) > len(moveTo) {
		log.Panic("Не хватает свободных мест для переселения")
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(moveTo), func(i, j int) { moveTo[i], moveTo[j] = moveTo[j], moveTo[i] })
	for i, c := range unhappyCitizens {
		j := i
		for c.Addr == moveTo[j] {
			j = rand.Intn(len(moveTo))
		}
		c.Addr = moveTo[j]
	}

	//citizens := append(unhappyCells, Map(func (h House) {}}Filter(func(h House) bool { h.(type) == Citizen }, street))
	//unhappyCells = append(unhappyCells, Filter(func(h House) bool { h.(type) == Citizen }, street))
	//for _, row := range sf.Grid {

}

func (sf SField) GetEmptyAddreses() []Addr {
	//emptyAddrs := make([]Addr, 0, int(math.Pow(float64(FIELD_SIZE), 2)))
	emptyHouses := sf.FilterHouses(func(h House) bool { _, ok := h.(Empty); return ok })
	emptyAddrs := Map(func(h House) Addr { return h.getAddr() }, emptyHouses)
	return emptyAddrs
}

func test1() {
	filename := "test1_"
	fieldSize := uint(5)
	testCitizens := []Citizen{Citizen{Blue, false, Addr{0, 1}}, Citizen{Red, false, Addr{0, 3}}, Citizen{Red, false, Addr{1, 1}}, Citizen{Red, false, Addr{1, 2}}, Citizen{Blue, false, Addr{1, 3}}, Citizen{Blue, false, Addr{2, 0}}, Citizen{Red, false, Addr{2, 2}}, Citizen{Red, false, Addr{2, 4}}, Citizen{Blue, false, Addr{3, 1}}, Citizen{Red, false, Addr{3, 3}}, Citizen{Red, false, Addr{4, 2}}, Citizen{Blue, false, Addr{4, 4}}}
	testSF := NewSField(fieldSize)
	testSF.Inhabit(testCitizens)
	testSF.Render(filename + ".png")
	for i := 0; i < int(3); i++ {
		testSF.HappinesAssessment()
		testSF.ShuffleUnhappyCells()
		testSF.Inhabit(testCitizens)
		testSF.Render(filename + "_" + string(i) + ".png")
	}
}

func main() {
	var filename string
	filename = time.Now().String()[:19]
	fieldSize := FIELD_SIZE
	citizens := CreateCitizens(fieldSize)
	citizens = RegisterCitizens(citizens, fieldSize)
	sf := NewSField(fieldSize)
	sf.Inhabit(citizens)
	sf.HappinesAssessment()
	unhappyCoeff := float64(len(sf.GetUnhappyCitizens())) / float64(sf.CountOfCitizens())
	fmt.Printf("Коэффициент несчастливых жителей = %.3f", unhappyCoeff)
	//os.Exit(1)
	sf.Render(filename + ".png")
	for i := 0; i < int(ITER_COUNT); i++ {
		sf.HappinesAssessment()
		sf.ShuffleUnhappyCells()
		sf.Inhabit(citizens)
		if i%1 == 0 {
			f := fmt.Sprintf("%v_%v.png", filename, i)
			sf.Render(f)
		}
	}
}
