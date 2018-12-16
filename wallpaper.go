// Copyright 2018 Ian Johnson
//
// This file is part of wallpaper.
//
// Wallpaper is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Wallpaper is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Wallpaper. If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"time"
)

var (
	width, height *int
	output        *string
	discrete      *bool
)

var pictures = map[string]picture{
	"gradient":   gradient,
	"mandelbrot": mandelbrot,
}

func init() {
	const (
		defaultWidth  = 1366
		defaultHeight = 738
		defaultOutput = "wallpaper.png"
	)

	width = flag.Int("w", defaultWidth, "set the width of the generated image")
	height = flag.Int("h", defaultHeight, "set the height of the generated image")
	output = flag.String("o", defaultOutput, "set the output file")
	discrete = flag.Bool("d", false, "use only colors from the given list")
}

func main() {
	flag.Parse()
	rand.Seed(time.Now().Unix())

	if len(flag.Args()) == 0 {
		fmt.Fprintln(os.Stderr, "wallpaper: no picture specified")
		os.Exit(2)
	}
	pic, ok := pictures[flag.Args()[0]]
	if !ok {
		fmt.Fprintf(os.Stderr, "wallpaper: unknown picture %v\n", flag.Args()[0])
		os.Exit(2)
	}

	colors, err := readColors(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "wallpaper: could not read colors: %v\n", err)
		os.Exit(1)
	}
	if len(colors) < 2 {
		fmt.Fprintln(os.Stderr, "wallpaper: not enough colors")
		os.Exit(1)
	}

	c1, c2 := chooseTwo(colors)
	var color func(grad float64) color.Color
	if *discrete {
		color = discreteColor(c1, c2, colors)
	} else {
		color = continousColor(c1, c2)
	}
	img := wallpaper{
		w:         *width,
		h:         *height,
		gradation: pic(*width, *height, flag.Args()[1:]),
		color:     color,
	}

	out, err := os.Create(*output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "wallpaper: could not create image file: %v\n", err)
		os.Exit(1)
	}

	if err := png.Encode(out, img); err != nil {
		out.Close()
		fmt.Fprintf(os.Stderr, "wallpaper: could not write image: %v\n", err)
		os.Exit(1)
	}

	if err := out.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "wallpaper: could not close output: %v\n", err)
		os.Exit(1)
	}
}

// picture is a function that, given the width and height of the desired
// wallpaper along with a set of command line arguments, returns a gradation
// function. If the given arguments are invalid for whatever reason, the
// picture should end the program or return some sensible default gradation.
type picture func(w, h int, args []string) func(x, y int) float64

// wallpaper is an image that colors each pixel using a combination of a
// "gradation function", producing a value between 0 and 1, and a "color
// function", turning the gradation value into a color.
type wallpaper struct {
	w, h      int
	gradation func(x, y int) float64
	color     func(grad float64) color.Color
}

func (w wallpaper) ColorModel() color.Model {
	return color.RGBAModel
}

func (w wallpaper) Bounds() image.Rectangle {
	return image.Rect(0, 0, w.w, w.h)
}

func (w wallpaper) At(x, y int) color.Color {
	return w.color(w.gradation(x, y))
}

// continuousColor returns a continuous color function, mapping gradation values
// evenly between the two given colors.
func continousColor(c1, c2 color.Color) func(float64) color.Color {
	return func(grad float64) color.Color {
		return gradate(c1, c2, grad)
	}
}

// discreteColor returns a discrete color function, mapping gradation values as
// in continousColor but only returning colors in the given slice.
func discreteColor(c1, c2 color.Color, colors []color.Color) func(float64) color.Color {
	return func(grad float64) color.Color {
		return closest(gradate(c1, c2, grad), colors)
	}
}

// closest returns the color from the given slice closest to the given color. It
// panics if the slice is empty.
func closest(c color.Color, colors []color.Color) color.Color {
	if len(colors) == 0 {
		panic("no colors to choose from")
	}
	close := colors[0]
	d := distance(c, close)
	for _, color := range colors[1:] {
		nd := distance(c, color)
		if nd < d {
			close = color
			d = nd
		}
	}
	return close
}

// distance returns a measure of how "far away" two colors are.
func distance(c1, c2 color.Color) float64 {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()
	dr := float64(r2) - float64(r1)
	dg := float64(g2) - float64(g1)
	db := float64(b2) - float64(b1)
	return dr*dr + dg*dg + db*db
}

// gradate returns a color "between" c1 and c2, with a value of 0 being c1
// exactly and a value of 1 being c2 exactly.
func gradate(c1, c2 color.Color, value float64) color.Color {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()
	rf1, gf1, bf1 := float64(r1/255), float64(g1/255), float64(b1/255)
	rf2, gf2, bf2 := float64(r2/255), float64(g2/255), float64(b2/255)
	r := uint8((rf2-rf1)*value + rf1)
	g := uint8((gf2-gf1)*value + gf1)
	b := uint8((bf2-bf1)*value + bf1)
	return color.RGBA{r, g, b, 255}
}

// chooseTwo chooses two colors at random from the given slice. It panics if
// there are fewer than two colors provided.
func chooseTwo(colors []color.Color) (color.Color, color.Color) {
	if len(colors) < 2 {
		panic("not enough colors")
	}
	i1 := rand.Intn(len(colors))
	i2 := rand.Intn(len(colors) - 1)
	if i2 == i1 {
		i2++
	}
	return colors[i1], colors[i2]
}

var colorRegexp = regexp.MustCompile("^#([A-Fa-f0-9]{2})([A-Fa-f0-9]{2})([A-Fa-f0-9]{2})$")

// readColors reads a color from each line of the given reader, returning a slice
// of all the colors found (or an error if one or more lines is not a color).
func readColors(r io.Reader) ([]color.Color, error) {
	var colors []color.Color
	s := bufio.NewScanner(r)
	for s.Scan() {
		if match := colorRegexp.FindStringSubmatch(s.Text()); match != nil {
			r, _ := strconv.ParseUint(match[1], 16, 8)
			g, _ := strconv.ParseUint(match[2], 16, 8)
			b, _ := strconv.ParseUint(match[3], 16, 8)
			colors = append(colors, color.RGBA{uint8(r), uint8(g), uint8(b), 255})
		} else {
			return nil, fmt.Errorf("not a color: %v", s.Text())
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return colors, nil
}
