package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

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

func (c Country) String() string {
	s := c.Name + " ["
	for _, c_ := range c.Neighbors {
		s += c_.Name + ","
	}
	s += "]"
	s += fmt.Sprintf(" color: %v", c.Color)
	return s
}

func (c Country) saturation() int {
	colors := map[int]bool{}
	for _, n := range c.Neighbors {
		if n.Color != 0 {
			colors[n.Color] = true
		}
	}
	return len(colors)
}

func (c *Country) setColor() {
	colors := map[int]bool{}
	for _, n := range c.Neighbors {
		if n.Color != 0 { // 0 means not colored
			colors[n.Color] = true
		}
	}
	// Find the first available color
	color := 1
	for colors[color] {
		color += 1
	}
	c.Color = color
}

// Map
// -----------------------------------------------------------------------------
type Map []*Country

func (m Map) String() (s string) {
	for _, c := range m {
		s += c.String() + "\n"
	}
	return
}

func (m Map) ComputeNeighbors() {
	for i, c1 := range m {
		for j := i + 1; j < len(m); j++ {
			c2 := m[j]
			if c1.Geometry.overlap(c2.Geometry) > 0 {
				c1.Neighbors = append(c1.Neighbors, c2)
				c2.Neighbors = append(c2.Neighbors, c1)
			}
		}
	}
}

func DSATUR(Countries Map) {
	countries := []*Country{}
	countries = append(countries, []*Country(Countries)...)

	for len(countries) > 0 {
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
		country.setColor()
		countries = countries[:len(countries)-1]
	}
}

// SVG Export
// -----------------------------------------------------------------------------
var Colormap map[int]string = map[int]string{
	0: "#e9ecef", // light grey
	1: "green",
	2: "yellow",
	3: "orange",
	4: "red",
	5: "magenta",
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
			s := fmt.Sprintf("too many colors: color #%v", country.Color)
			panic(s)
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
type empty struct{}

func main() {
	paths := os.Args[1:]
	N := len(paths)
	sem := make(chan empty, N)

	for _, path := range paths {
		go func(path string) {
			file, err := os.Open(path)
			if err != nil {
				panic(err)
			}
			defer file.Close()

			var Countries Map
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
					country := &Country{Name: name, Geometry: geometry}
					Countries = append(Countries, country)
				}
			}

			Countries.ComputeNeighbors()
			DSATUR(Countries)

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
			sem <- empty{} // signal the go routine end
		}(path)
	}
	for i := 0; i < N; i++ { // wait for the end of all goroutines
		<-sem
	}
}
