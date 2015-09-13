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

func TestDeadEnds(t *testing.T) {
	m := emptyMaze(2, 2)
	m.rooms[0][0].Link(&m.rooms[0][1])

	tests := []struct {
		name string
		m    *Maze
		want int
	}{
		{"2 by 2 empty maze", emptyMaze(2, 2), 0},
		{"2 by 2 full maze", fullMaze(2, 2), 0}, // isolated room is not a dead end
		{"2 by 2 maze where (0, 0) is linked to (0, 1)", m, 2},
	}

	for _, tt := range tests {
		got := len(tt.m.DeadEnds())
		if got != tt.want {
			t.Errorf("%s: got %d; want %d", tt.name, got, tt.want)
		}
	}
}
