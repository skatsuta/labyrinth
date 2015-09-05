package commands

import (
	"testing"

	"github.com/skatsuta/labyrinth/mazelib"
)

func TestPrintMaze(t *testing.T) {
	x, y := 15, 10
	z := createMaze(x, y)
	mazelib.PrintMaze(z)
	t.Fail()
}
