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
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math/cmplx"
	"os"
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

// graphic returns a gradation function based on the image whose filepath is
// given as an argument. The gradation is based on the grayscale conversion of
// the image.
func graphic(w, h int, args []string) func(x, y int) float64 {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: graphic <filepath>")
		os.Exit(2)
	}

	img, err := loadGrayImage(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "graphic: %v\n", err)
		os.Exit(2)
	}

	// Calculate the x and y scales for the image.
	sx := float64(w) / float64(img.Bounds().Dx())
	sy := float64(h) / float64(img.Bounds().Dy())
	// Preserve aspect ratio by taking the max and setting both scales to it.
	if sx > sy {
		sy = sx
	} else {
		sx = sy
	}

	return func(x, y int) float64 {
		// Calculate the "projected" x and y onto the original image.
		px := float64(x)/sx + float64(img.Bounds().Min.X)
		py := float64(y)/sy + float64(img.Bounds().Min.Y)
		// TODO: use some sort of interpolation instead of "nearest
		// neighbor".
		c := img.GrayAt(int(px), int(py))
		return float64(c.Y) / 255
	}
}

// loadGrayImage loads the image at the given filepath as a grayscale image.
func loadGrayImage(filepath string) (*image.Gray, error) {
	in, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("opening image: %v", err)
	}
	defer in.Close()

	img, _, err := image.Decode(in)
	if err != nil {
		return nil, fmt.Errorf("decoding image: %v", err)
	}

	gray := image.NewGray(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()))
	for x := 0; x < img.Bounds().Dx(); x++ {
		for y := 0; y < img.Bounds().Dy(); y++ {
			gray.Set(x, y, img.At(img.Bounds().Min.X+x, img.Bounds().Min.Y+y))
		}
	}
	return gray, nil
}
