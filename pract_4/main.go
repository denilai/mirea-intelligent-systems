package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"slices"
	"time"

	"github.com/fogleman/gg"
)

// Управляющие параметры
const (
	// количество попыток построить генераций случайного поля
	GEN_TRY_COUNT uint = 10
	// процент заполнения ячеек на поле
	FILLING_THRESHOLD float64 = 0.9
	// размер одной ячейки в пикселях
	CELL_SIZE_PX uint = 10
	// количество итераций
	ITER_COUNT uint = 10000
	// порог счастья
	HAPPY_THRESHOLD uint = 4
	// Размер поля
	FIELD_SIZE uint = 100
	// Отрисовывать каждые n итераций
	DRAW_EVERY uint = 5
	DEBUG      bool = false
)

// Empty: color.RGBA{252, 255, 231, 255}}
var COLOR_MODEL = map[Color]color.RGBA{Blue: color.RGBA{43, 51, 103, 255}, Red: color.RGBA{235, 69, 95, 255}, White: color.RGBA{252, 255, 231, 255}, Black: color.RGBA{0, 0, 0, 0}, Green: color.RGBA{47, 211, 30, 170}, Grey: color.RGBA{25, 29, 25, 200}}

// Тип для цвета ячейки
type Color uint8

func (c Color) String() string {
	if c == Red {
		return "Red"
	} else if c == Blue {
		return "Blue"
	} else if c == White {
		return "White"
	} else if c == Black {
		return "Black"
	} else if c == Green {
		return "Green"
	} else if c == Grey {
		return "Grey"
	} else {
		return fmt.Sprintf("%v", c)
	}
}

const (
	Red Color = iota
	Blue
	White
	Black
	Green
	Grey
	ColorLen
)

type Addr struct {
	Row, Col uint
}

func (a Addr) String() string {
	return fmt.Sprintf("{Row: %v, Col: %v}", a.Row, a.Col)
}

type House interface {
	IsEmpty() bool
	GetAddr() Addr
	SetAddr(Addr)
	String() string
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

func (c Citizen) IsEmpty() bool      { return false }
func (c *Citizen) SetAddr(addr Addr) { c.Addr = addr }
func (c Citizen) GetAddr() Addr      { return c.Addr }
func (c Citizen) String() string {
	return fmt.Sprintf("Citizen {Tag: %v, Happy: %v, Addr: %v}", c.Tag, c.Happy, c.Addr)
}

func (c Empty) IsEmpty() bool      { return true }
func (e Empty) GetAddr() Addr      { return e.Addr }
func (e *Empty) SetAddr(addr Addr) { e.Addr = addr }
func (e Empty) String() string {
	return fmt.Sprintf("Empty {Addr: %v}", e.Addr)
}

// квадратное поле для проведения эксперимента Шеллинга
type SField struct {
	Size uint
	Grid []Street
}

type SFieldStats struct {
	All         uint
	Filled      uint
	Empty       uint
	Happy       uint
	Unhappy     uint
	FillFactor  float64
	HappyFactor float64
	TagMap      map[Color]uint
}

func (sf *SField) Clean() {
	for row := range sf.Grid {
		for col := range sf.Grid[row] {
			uCol, uRow := uint(col), uint(row)
			sf.Grid[uRow][uCol] = &Empty{Addr{Col: uCol, Row: uRow}}
		}
	}

}

// Inhabit (англ.) -- населить
func (sf *SField) Inhabit(hs []House) {
	if DEBUG {
		defer Duration(Track("Inhabit"))
	}
	sf.Clean()
	for i := range hs {
		addr := hs[i].GetAddr()
		sf.Grid[addr.Row][addr.Col] = hs[i]
	}
}

func (sf *SField) HappinesAssessment() {
	allHouses := sf.FlattenHouses()
	for _, house := range allHouses {
		if count, err := sf.CountOfNeighbours(house); err != nil {
			continue
		} else {
			if c, ok := house.(*Citizen); ok {
				if count < int(HAPPY_THRESHOLD) {
					if DEBUG {
						log.Printf("%v is't happy(", c)
					}
					c.Happy = false
				} else {
					if DEBUG {
						log.Printf("%v is happy!", c)
					}
					c.Happy = true
				}
			} else {
				log.Panic("Неожиданный тип. Ожидался *Citizen, получен %T", house)
			}
		}
	}
}

func NewSField(size uint) SField {
	if DEBUG {
		defer Duration(Track("NewSField"))
	}
	grid := make([]Street, int(size))
	for row := range grid {
		grid[row] = make([]House, int(size))
		for col := range grid[row] {
			uCol, uRow := uint(col), uint(row)
			grid[row][col] = &Empty{Addr{Row: uRow, Col: uCol}}
		}
	}
	return SField{size, grid}
}

func (sf SField) Stats() SFieldStats {
	allHouses := sf.FlattenHouses()
	emptyCount := uint(len(Filter(func(h House) bool { return h.IsEmpty() }, allHouses)))
	happyCount := uint(0)
	unhappyCount := uint(0)
	allCount := uint(len(allHouses))
	tagMap := make(map[Color]uint)
	for _, h := range allHouses {
		if c, ok := h.(*Citizen); ok {
			if c.Happy {
				happyCount++
			} else {
				unhappyCount++
			}
			if _, ok := tagMap[c.Tag]; !ok {
				tagMap[c.Tag] = 1
			} else {
				tagMap[c.Tag]++
			}
		}
	}
	return SFieldStats{All: allCount, Filled: allCount - emptyCount, Empty: emptyCount, FillFactor: float64(allCount-emptyCount) / float64(allCount), TagMap: tagMap, Happy: happyCount, Unhappy: unhappyCount, HappyFactor: float64(happyCount) / float64(allCount-emptyCount)}
}

func (s SFieldStats) String() string {
	res := "================\n"
	res += fmt.Sprintf("Статистика SField:\n")
	res += fmt.Sprintf("Кол-во ячеек: %v\n", s.All)
	res += fmt.Sprint("Из них:\n")
	res += fmt.Sprintf("  Пустых: %v\n", s.Empty)
	res += fmt.Sprintf("  Заполненных: %v\n", s.Filled)
	res += fmt.Sprint("  Из них (по цветам):\n")
	res += fmt.Sprintf("    %v\n", s.TagMap)
	res += fmt.Sprint("  Из них (по счастью):\n")
	res += fmt.Sprintf("    Несчастливых: %v\n", s.Unhappy)
	res += fmt.Sprintf("    Счастливых: %v\n", s.Happy)
	res += fmt.Sprintf("  Коэффициент заполнения: %.2f\n", s.FillFactor)
	res += fmt.Sprintf("  Коэффициент счастья: %.2f\n", s.HappyFactor)
	res += "================\n"
	return res
}

func CreateHouses(fsize uint) []House {
	if DEBUG {
		defer Duration(Track("CreateHouses"))
	}
	c := int(math.Pow(float64(fsize), 2))
	hs := make([]House, c)
	redCount := int(float64(c) * float64(FILLING_THRESHOLD) / 2)
	if DEBUG {
		fmt.Println("Количество ячеек =", c)
		fmt.Println("Количество красных и синих ячеек =", redCount)
	}
	for i := 0; i < redCount; i++ {
		hs[i] = &Citizen{Tag: Red, Happy: false, Addr: Addr{}}
		hs[i+redCount] = &Citizen{Tag: Blue, Happy: false, Addr: Addr{}}
	}
	for i := redCount * 2; i < c; i++ {
		hs[i] = &Empty{Addr{}}
	}
	return hs
}

// seed = -1 --  значение по умолчанию (сгенерировать зерно)
func MyShuffle[T any](list []T, seed int64) int64 {
	if seed == -1 {
		seed = time.Now().UnixNano()
	}
	ns := rand.New(rand.NewSource(seed))
	ns.Shuffle(len(list), func(i, j int) { list[i], list[j] = list[j], list[i] })
	return seed
}

func ShuffleHouses(hs []House) {
	MyShuffle(hs, -1)
	if DEBUG {
		log.Printf("Горожане перемешаны")
	}
}

// Выдать жильцам адреса
func RegisterCitizens(citizens []Citizen, size uint) []Citizen {
	if DEBUG {
		defer Duration(Track("RegisterCitizen"))
	}
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
	if DEBUG {
		log.Printf("Количество уникальных адресов = %v\n", len(addresses))
	}
	return citizens
}

func TestCTX() {

	ctx := gg.NewContext(100, 100)
	ctx.Push()
	ctx.DrawRectangle(10, 10, 10, 10)
	ctx.SetRGB(100, 100, 100)
	ctx.Fill()
	ctx.Pop()
	ctx.DrawCircle(10, 10, 10)
	ctx.SetRGB(200, 100, 100)
	ctx.Fill()
	ctx.SavePNG("ctx_test.png")
}

func (sf SField) Render(filename string, drawHappines bool) {
	if DEBUG {
		defer Duration(Track("Render"))
	}
	ctx := gg.NewContext(int(sf.Size*CELL_SIZE_PX), int(sf.Size*CELL_SIZE_PX))
	iCELL_SIZE, fCELL_SIZE := int(CELL_SIZE_PX), float64(CELL_SIZE_PX)

	for row, street := range sf.Grid {
		for col, house := range street {
			var color color.RGBA
			var tag Color
			ctx.DrawRectangle(float64(iCELL_SIZE*col), float64(iCELL_SIZE*row), fCELL_SIZE, fCELL_SIZE)
			c, isCitizen := house.(*Citizen)
			if isCitizen {
				tag = c.Tag
			} else if _, ok := house.(*Empty); ok {
				tag = White
			} else {
				log.Panic("Неожиданный тип. Ожидался *Citizen или *Empty")
			}
			color, ok := COLOR_MODEL[tag]
			if !ok {
				log.Panicf("Несуществующий цвет под тэгом: %v", tag)
			}
			ctx.SetColor(color)
			ctx.Fill()
			if isCitizen && drawHappines {
				if c.Happy {
					if DEBUG {
						log.Printf("%v отмечен как счастливый", c)
					}
					ctx.DrawCircle(float64(fCELL_SIZE*float64(col)+0.5*fCELL_SIZE), float64(fCELL_SIZE*float64(row)+0.5*fCELL_SIZE), fCELL_SIZE*0.2)
					ctx.SetColor(COLOR_MODEL[Green])
				} else {
					if DEBUG {
						log.Printf("%v отмечен как НЕсчастливый", c)
					}
					ctx.DrawCircle(float64(fCELL_SIZE*float64(col)+0.5*fCELL_SIZE), float64(fCELL_SIZE*float64(row)+0.5*fCELL_SIZE), fCELL_SIZE*0.2)
					ctx.SetColor(COLOR_MODEL[Grey])
				}
				ctx.Fill()
			}
		}
	}
	fmt.Printf("Создаем изображение %v ...\n", filename)
	if er := ctx.SavePNG(filename); er != nil {
		log.Println("Error in SavePNG")
	}
}

func (sf SField) CountOfNeighbours(h House) (int, error) {
	if DEBUG {
		defer Duration(Track("CountOfNeighbours"))
	}
	count := 0
	addr := h.GetAddr()
	c, ok := sf.Grid[addr.Row][addr.Col].(*Citizen)
	if !ok {
		return 0, fmt.Errorf("House (cell) with addres {%v} is empty", addr)
	}
	myTag := c.Tag
	intCol, intRow := int(addr.Col), int(addr.Row)
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if intRow+dy < 0 || intCol+dx < 0 || intCol+dx >= int(sf.Size) || intRow+dy >= int(sf.Size) || (dx == 0 && dy == 0) {
				continue
			}
			//log.Printf("dx = %v, dy = %v", dx, dy)
			if c1, ok := sf.Grid[intRow+dy][intCol+dx].(*Citizen); ok && c1.Tag == myTag {
				count += 1
			}
		}
	}
	if DEBUG {
		log.Printf("У %v %v близких соседей", h, count)
	}
	return count, nil
}

func Map[T any, M any](f func(T) M, data []T) []M {
	if DEBUG {
		defer Duration(Track("Map"))
	}
	n := make([]M, len(data))
	for i, e := range data {
		n[i] = f(e)
	}
	return n
}

func Filter[T any](f func(T) bool, data []T) []T {
	if DEBUG {
		defer Duration(Track("Filter"))
	}
	fltd := make([]T, 0, len(data))
	for _, e := range data {
		if f(e) {
			fltd = append(fltd, e)
		}
	}
	return fltd
}

func (sf SField) FlattenHouses() []House {
	allHouses := []House{}
	for i := range sf.Grid {
		allHouses = append(allHouses, sf.Grid[i]...)
	}
	return allHouses
}

func (sf SField) GetUnhappyCitizens() []*Citizen {
	return Map(func(h House) *Citizen {
		if c, ok := h.(*Citizen); ok {
			return c
		} else {
			log.Panicf("Неожиданный тип. Ожидался *Citizen, получен %T", c)
			return nil
		}
	}, Filter(func(h House) bool { c, ok := h.(*Citizen); return ok && !c.Happy }, sf.FlattenHouses()))
}

func EraseElem[T any](data []T, i int) []T {
	data[i] = data[len(data)-1]
	data = data[:len(data)-1]
	return data
}

func (sf *SField) ShuffleUnhappyCells() error {
	emptyAddrs := sf.GetEmptyAddreses()
	unhappyCitizens := sf.GetUnhappyCitizens()
	var moveTo []Addr
	if len(unhappyCitizens) == 0 {
		return fmt.Errorf("Все счастливы")
	}
	if DEBUG {
		log.Printf("Несчаcтливые жители: %v, len = %v", unhappyCitizens, len(unhappyCitizens))
	}
	unhappyAddrs := Map(func(c *Citizen) Addr { return c.GetAddr() }, unhappyCitizens)

	if len(unhappyCitizens) > len(emptyAddrs) {
		moveTo = append(emptyAddrs, unhappyAddrs...)
	} else {
		moveTo = emptyAddrs
	}
	if len(unhappyCitizens) > len(moveTo) {
		return fmt.Errorf("Не хватает свободных мест для переселения")
	}
	if DEBUG {
		fmt.Printf("Места для переселения: %v, len = %v\n", moveTo, len(moveTo))
	}
	MyShuffle(moveTo, -1)
	if DEBUG {
		log.Printf("Места перемешаны")
	}
	for _, c := range unhappyCitizens {
		j := rand.Intn(len(moveTo))
		for c.GetAddr() == moveTo[j] {
			j = rand.Intn(len(moveTo))
		}
		c.SetAddr(moveTo[j])
		if DEBUG {
			log.Printf("%v переселили в %v", c, moveTo[j])
		}
		moveTo = EraseElem(moveTo, j)
	}
	return nil
}

func (sf SField) GetEmptyAddreses() []Addr {
	return Map(func(h House) Addr { return h.GetAddr() }, Filter(func(h House) bool { return h.IsEmpty() }, sf.FlattenHouses()))
}

//func test1() {
//	filename := "test1_"
//	fieldSize := uint(5)
//	testCitizens := []Citizen{Citizen{Blue, false, Addr{0, 1}}, Citizen{Red, false, Addr{0, 3}}, Citizen{Red, false, Addr{1, 1}}, Citizen{Red, false, Addr{1, 2}}, Citizen{Blue, false, Addr{1, 3}}, Citizen{Blue, false, Addr{2, 0}}, Citizen{Red, false, Addr{2, 2}}, Citizen{Red, false, Addr{2, 4}}, Citizen{Blue, false, Addr{3, 1}}, Citizen{Red, false, Addr{3, 3}}, Citizen{Red, false, Addr{4, 2}}, Citizen{Blue, false, Addr{4, 4}}}
//	testSF := NewSField(fieldSize)
//	testSF.Inhabit(testCitizens)
//	testSF.Render(filename + ".png")
//	for i := 0; i < int(3); i++ {
//		testSF.HappinesAssessment()
//		testSF.ShuffleUnhappyCells()
//		testSF.Inhabit(testCitizens)
//		testSF.Render(filename + "_" + string(i) + ".png")
//	}
//}

func RandomRegistration(hs []House, seed int64) int64 {
	seed = MyShuffle(hs, seed)
	if DEBUG {
		log.Printf("Горожане перемешаны")
	}
	d := math.Sqrt(float64(len(hs)))
	for i := range hs {
		row := uint(float64(i) / d)
		col := uint(i % int(d))
		(hs)[i].SetAddr(Addr{Col: col, Row: row})
	}
	return seed
}

func main() {
	var filename, foldername string
	foldername = "results/" + time.Now().String()[:19]
	fsize := FIELD_SIZE
	// Создаем каталог с результатами (если не существует)
	if err := os.MkdirAll(foldername, os.ModePerm); err != nil {
		log.Panic(err)
	}
	// Переходим в каталог
	if err := os.Chdir(foldername); err != nil {
		log.Panic(err)
	}
	//if _, err := os.Stat(foldername); errors.Is(err, os.ErrNotExist) {
	//	err := os.Mkdir(foldername, os.ModePerm)
	//	if err != nil {
	//		log.Panic(err)
	//	}
	//}
	// Создаем житилей согласно коэффициенту заполнения и пропорциям
	houses := CreateHouses(fsize)
	// Случайно выдаем адреса
	var minHappyFactor float64 = 1
	var seed int64
	sf := NewSField(fsize)
	for i := 0; i < int(GEN_TRY_COUNT); i++ {
		s := RandomRegistration(houses, -1)
		citizens := Filter(func(h House) bool { return !h.IsEmpty() }, houses)
		// Населяем поле жителями
		sf.Inhabit(citizens)
		curHF := sf.Stats().HappyFactor
		log.Printf("Фактор cчастья = %v", curHF)
		log.Printf("Зерно = %v", s)
		if curHF < minHappyFactor {
			minHappyFactor = curHF
			seed = s
		}
	}
	if DEBUG {
		log.Printf("Минимальный фактор счастья = %v с зерном %v", minHappyFactor, seed)
	}
	RandomRegistration(houses, seed)
	citizens := Filter(func(h House) bool { return !h.IsEmpty() }, houses)
	// Населяем поле жителями
	sf.Inhabit(citizens)

	sf.Stats()
	// Отрисовываем изначальное распределение (без обозачения счастья)
	//sf.Render("init.png", false)
	for i := 0; i < int(ITER_COUNT); i++ {
		// Оценка счастья
		sf.HappinesAssessment()
		// Определить новые адреса для несчастливых жителей
		// Заселить несчастливых по новым адресам
		if err := sf.ShuffleUnhappyCells(); err != nil {
			fmt.Println(err)
			fmt.Println(sf.Stats())
			sf.Render("final.png", true)
			os.Exit(1)
		}
		sf.Inhabit(citizens)
		if i%int(DRAW_EVERY) == 0 {
			filename = fmt.Sprintf("%v.png", i)
			//filenameM := fmt.Sprintf("%vm.png", i)
			sf.Render(filename, false)
			//sf.Render(filenameM, true)
		}
	}
}
