package image

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Pixel struct {
	r, g, b uint8
}

// Ppm represents an image file in ppm format
type Ppm struct {
	path string
	file *os.File

	width, height uint

	// TODO(wittrock): support this, currently all images are assumed to have a maxval <= UINT8MAX
	maxVal uint

	data [][]Pixel
}

func NewPpmFromFile(path string) (Ppm, error) {
	return Ppm{
		path: path,
	}, nil
}

func NewPpm(path string, data [][]Pixel) Ppm {
	// TODO(wittrock): validate that all the lines are the same length
	return Ppm{
		path:   path,
		file:   nil,
		width:  uint(len(data[0])),
		height: uint(len(data)),
		maxVal: uint(255), // TODO(wittrock): support maxVal
		data:   data,
	}
}

func (ppm Ppm) withPath(path string) Ppm {
	ppmCopy := ppm
	ppmCopy.path = path
	return ppmCopy
}

func (ppm *Ppm) Write() error {
	var err error
	fmt.Printf("opening %s\n", ppm.path)
	ppm.file, err = os.OpenFile(ppm.path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	ppm.file.WriteString("P3\n")
	ppm.file.WriteString(fmt.Sprintf("%d %d\n", ppm.width, ppm.height))
	ppm.file.WriteString("255\n") // TODO(wittrock): support maxval

	for r := range ppm.data {
		for c := range ppm.data[r] {
			pixel := ppm.data[r][c]
			ppm.file.WriteString(fmt.Sprintf("%d %d %d\t", pixel.r, pixel.g, pixel.b))
		}
		ppm.file.WriteString("\n")
	}

	ppm.file = nil
	ppm.file.Close()

	return nil
}

func (ppm *Ppm) Read() error {
	var err error
	ppm.file, err = os.Open(ppm.path)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(ppm.file)

	// Get first line, check file magic
	scanner.Scan()
	magic := scanner.Text()
	if magic != "P3" {
		return errors.New("invalid magic")
	}

	// Get the second line, set width and height
	scanner.Scan()
	whLine := scanner.Text()
	widthHeight := strings.Fields(whLine)
	if len(widthHeight) != 2 {
		return errors.New("invalid width/height line")
	}
	width, err := strconv.ParseUint(widthHeight[0], 10, 32)
	if err != nil {
		return errors.New("invalid width")
	}
	ppm.width = uint(width)

	height, err := strconv.ParseUint(widthHeight[1], 10, 32)
	if err != nil {
		return errors.New("invalid height")
	}
	ppm.height = uint(height)

	// initialize the data array
	ppm.data = make([][]Pixel, ppm.height)
	for i := range ppm.data {
		ppm.data[i] = make([]Pixel, ppm.width)
	}

	// Get the third line, set the max val
	scanner.Scan()
	maxValLine := scanner.Text()
	maxVal, err := strconv.ParseUint(maxValLine, 10, 32)
	if err != nil {
		return errors.New("invalid maxVal")
	}
	ppm.maxVal = uint(maxVal)

	// get the rest of the lines as image data
line:
	for row := uint(0); scanner.Scan() && row < ppm.height; row++ {
		text := scanner.Text()
		lineScanner := bufio.NewScanner(strings.NewReader(text))
		lineScanner.Split(bufio.ScanWords)

		pixel := Pixel{}
		for col := uint(0); lineScanner.Scan() && col/3 < ppm.width; col++ {
			fragmentVal := lineScanner.Text()
			if fragmentVal == "#" {
				// Skip the comment
				row--
				continue line
			}

			pixelVal, err := strconv.ParseUint(fragmentVal, 10, 8)
			if err != nil {
				return errors.New("invalid pixel val")
			}

			switch col % 3 {
			case 0:
				pixel.r = uint8(pixelVal)
			case 1:
				pixel.g = uint8(pixelVal)
			case 2:
				pixel.b = uint8(pixelVal)
			}

			if col%3 == 2 {
				ppm.data[row][col/3] = pixel
			}
		}
		fmt.Printf("next line\n")
	}

	return nil
}
