package commands

import (
	"testing"

	"github.com/skatsuta/labyrinth/mazelib"
)

func TestPrintMaze(t *testing.T) {
	x, y := 15, 10
	z := createMaze(x, y)
	mazelib.PrintMaze(z)
}

func TestConfigureRooms(t *testing.T) {
	w, h := 2, 2
	z := emptyMaze(w, h)

	configureRooms(z)

	want := 2
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			room, _ := z.GetRoom(x, y)
			got := len(room.Nbr)
			if got != want {
				t.Errorf("each size of neigbors must be %d, but got %d", want, got)
			}
		}
	}
}

func TestAllRooms(t *testing.T) {
	tests := []struct {
		w, h int
		want int
	}{
		{1, 1, 1},
		{1, 2, 2},
		{2, 2, 4},
	}

	for _, tt := range tests {
		m := emptyMaze(tt.w, tt.h)
		got := len(m.AllRooms())
		if got != tt.want {
			t.Errorf("got %d; want %d", got, tt.want)
		}
	}
}
