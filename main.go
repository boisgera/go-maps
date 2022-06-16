package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
)

// Empty Struct
// -----------------------------------------------------------------------------
type empty = struct{}

var nothing = empty{}

// Sets
// -----------------------------------------------------------------------------
type Set[E comparable] map[E]empty

// Geometry
// -----------------------------------------------------------------------------
type Geometry [4]int // xmin, ymin, xmax, ymax

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func overlap(first, second [2]int) int {
	low := max(first[0], second[0])
	high := min(first[1], second[1])
	return high - low
}

func (first Geometry) overlap(second Geometry) int {
	first_x := [2]int{first[0], first[2]}
	first_y := [2]int{first[1], first[3]}
	second_x := [2]int{second[0], second[2]}
	second_y := [2]int{second[1], second[3]}
	o1 := overlap(first_x, second_x)
	o2 := overlap(first_y, second_y)
	if o1 >= 0 && o2 >= 0 {
		return o1 + o2
	} else {
		return 0
	}
}

// Country
// -----------------------------------------------------------------------------
type Country struct {
	Name      string
	Geometry  Geometry
	Neighbors []*Country
	Color     int
}

func (c *Country) String() string {
	s := c.Name + ": neighbors = "
	for _, c_ := range c.Neighbors {
		s += c_.Name + ", "
	}
	s += fmt.Sprintf(" color = %v", c.Color)
	return s
}

func (c *Country) neighborsColors() Set[int] {
	colors := Set[int]{}
	for _, n := range c.Neighbors {
		if n.Color != 0 { // 0 means not colored
			colors[n.Color] = nothing
		}
	}
	return colors
}

func (c *Country) saturation() int {
	return len(c.neighborsColors())
}

func (c *Country) setColor(colors_ ...int) (err error) {
	if len(colors_) >= 1 {
		c.Color = colors_[0]
		return
	}

	colors := c.neighborsColors()
	// Find the first available color ("real" one: >=1)
	color := 0
	exists := true
	for exists {
		color += 1
		_, exists = colors[color]
	}
	c.Color = color
	if color > 4 {
		err = fmt.Errorf("invalid color: %v", color)
	}
	return
}

// Map
// -----------------------------------------------------------------------------
type Map []*Country

func (m Map) clearColors() {
	for _, c := range m {
		c.setColor(0)
	}
}

func (m Map) String() (s string) {
	for _, c := range m {
		s += c.String() + "\n"
	}
	return
}

func (m Map) ComputeNeighbors() {
	for i := range m {
		for j := i + 1; j < len(m); j++ {
			c1 := m[i]
			c2 := m[j]
			if c1.Geometry.overlap(c2.Geometry) > 0 {
				c1.Neighbors = append(c1.Neighbors, c2)
				c2.Neighbors = append(c2.Neighbors, c1)
			}
		}
	}
}

func loadMap(path string) Map {
	var Countries Map

	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			words := strings.Fields(line)
			name := words[0]
			geometry := [4]int{}
			for i := 1; i < 5; i++ {
				integer, error := strconv.Atoi(words[i])
				if error != nil {
					panic(error)
				}
				geometry[i-1] = integer
			}
			country := Country{Name: name, Geometry: geometry}
			Countries = append(Countries, &country)
		}
	}
	return Countries
}

// Buggy now that Map is a list of struct values, countries are not modified.
func DSATUR(Countries Map) (order Map, err error) {
	countries := Map{}
	countries = append(countries, Countries...)

	for len(countries) > 0 {
		//fmt.Println("*", len(countries))
		compare := func(i, j int) bool {
			c1 := countries[i]
			c2 := countries[j]
			if c1.saturation() < c2.saturation() {
				return true
			} else if c1.saturation() == c2.saturation() {
				if len(c1.Neighbors) < len(c2.Neighbors) {
					return true
				} else if len(c1.Neighbors) == len(c2.Neighbors) {
					return i < j
				}
			}
			return false
		}

		sort.Slice(countries, compare)
		country := countries[len(countries)-1]
		if country.setColor() != nil {
			err = errors.New("invalid color")
		}
		//fmt.Print("*", country.Color)
		order = append(order, country)
		countries = countries[:len(countries)-1]
	}
	return
}

func displayOrder(order []*Country) {
	l := []string{}
	for i := 0; i < len(order); i++ {
		l = append(l, order[i].Name)
	}
	fmt.Printf("%v", l)
}

// func SHUFFLE(Countries Map, order []*Country) (order_out []*Country) {
// 	// try everything in order. If some setColor fails, up it in the list
// 	// and try again.
// 	for iter := 0; iter < 1000; iter++ {
// 		fmt.Printf("iter: %v ", iter)
// 		Countries.clearColors()
// 		for i, c := range order {
// 			err := c.setColor()
// 			if err != nil {
// 				fmt.Printf("%v\n", i)
// 				displayOrder(order)
// 				a := order[0]
// 				order[0] = c
// 				order[i] = a
// 				displayOrder(order)
// 				break
// 			}
// 			if i == len(order)-1 {
// 				order_out = order
// 				return
// 			}
// 		}
// 	}
// 	return

// }

// SVG Export
// -----------------------------------------------------------------------------
var Colormap map[int]string = map[int]string{
	-1: "magenta",
	0:  "#e9ecef", // light grey
	1:  "green",
	2:  "yellow",
	3:  "orange",
	4:  "red",
}

func (m Map) SVG() (svg string) {
	minX := 100000
	maxX := -100000
	minY := 100000
	maxY := -100000
	for _, country := range m {
		geometry := country.Geometry
		minX = min(minX, geometry[0])
		maxX = max(maxX, geometry[2])
		minY = min(minY, geometry[1])
		maxY = max(maxY, geometry[3])
	}
	width := maxX - minX
	height := maxY - minY

	svg = fmt.Sprintf(`<svg version="1.1" viewBox="%v %v %v %v" width="%v" height="%v" xmlns="http://www.w3.org/2000/svg">`, minX, minY, width, height, width, height)
	svg += "\n"
	for _, country := range m {
		geometry := country.Geometry
		color, ok := Colormap[country.Color]
		if !ok {
			color = Colormap[-1]
		}
		x := geometry[0]
		y := geometry[1]
		width := geometry[2] - x
		height := geometry[3] - y
		rect := fmt.Sprintf(`<rect x="%v" y="%v" width="%v" height="%v" fill="%v" stroke="black" />`, x, y, width, height, color)
		svg += rect + "\n"
		text := fmt.Sprintf(`<text x="%v" y="%v">%v</text>`, x+1, y+10, country.Name)
		svg += text + "\n"
	}
	svg += "</svg>\n"
	return
}

// Main Entry Point
// -----------------------------------------------------------------------------
var profile bool = true

func main() {
	if profile {
		f, err := os.Create("main.prof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	paths := os.Args[1:]
	N := len(paths)
	sem := make(chan struct{}, N)

	for _, path := range paths {
		go func(path string) {
			Countries := loadMap(path)

			Countries.ComputeNeighbors()
			// fmt.Println(Countries)

			_, err := DSATUR(Countries)
			if err != nil {
				fmt.Printf("Map %v needs more than 4 colors\n", path)
			}
			// if err != nil {
			// 	order = SHUFFLE(Countries, order)
			// }

			out := path + ".svg"
			f, err := os.Create(out)
			if err != nil {
				panic(err)
			}
			_, err = f.WriteString(Countries.SVG())
			if err != nil {
				f.Close()
				panic(err)
			}
			sem <- struct{}{} // signal the go routine end
		}(path)
	}
	for i := 0; i < N; i++ { // wait for the end of all goroutines
		<-sem
	}
}
