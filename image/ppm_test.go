package image

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpen(t *testing.T) {
	ppm, err := NewPpmFromFile("./test.ppm")
	assert.Nil(t, err, "opening error")

	err = ppm.Read()
	assert.Nil(t, err, "reading error")

	fmt.Printf("got ppm: %v\n", ppm)

	ppm = ppm.withPath("./test_copy.ppm")
	err = ppm.Write()
	assert.Nil(t, err)
}
