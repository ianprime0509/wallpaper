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
	"flag"
	"math/cmplx"
)

// gradient returns a gradation function for a simple horizontal gradient given
// the width and height of the output image.
func gradient(w, h int, args []string) func(x, y int) float64 {
	return func(x, y int) float64 {
		return float64(x) / float64(w)
	}
}

// mandelbrot returns a gradation function for the Mandelbrot set, scaled to fit
// within the given width and height without changing its proportions.
func mandelbrot(w, h int, args []string) func(x, y int) float64 {
	flags := flag.NewFlagSet("mandelbrot", flag.ExitOnError)
	iterations := flags.Int("i", 50, "set the number of iterations")
	flags.Parse(args)

	cx := float64(w / 2)
	cy := float64(h / 2)
	var r float64 // the radius of the containing disk around the origin (in px)
	if h < w {
		r = float64(h / 2)
	} else {
		r = float64(w / 2)
	}

	return func(x, y int) float64 {
		c := complex(2*(float64(x)-cx)/r, 2*(float64(y)-cy)/r)
		z := complex128(0)
		var i int
		for i = 0; i < *iterations; i++ {
			if cmplx.Abs(z) > 2 {
				break
			}
			z = z*z + c
		}
		return float64(i) / float64(*iterations)
	}
}
