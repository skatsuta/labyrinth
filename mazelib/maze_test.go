package mazelib

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestIsLinked(t *testing.T) {
	r := []Room{NewRoom(), NewRoom(), NewRoom()}
	r[0].Link(&r[1])

	tests := []struct {
		i1, i2 int
		want   bool
	}{
		{0, 1, true},
		{0, 2, false},
		{1, 2, false},
	}

	for _, tt := range tests {
		got := r[tt.i1].IsLinked(&r[tt.i2])
		if got != tt.want {
			t.Errorf("%d linked with %d?: got %t, but want %t", tt.i1, tt.i2, got, tt.want)
		}
	}
}

func TestLink(t *testing.T) {
	r1, r2 := NewRoom(), NewRoom()
	r1.Link(&r2)

	got := r1.IsLinked(&r2)
	if !got {
		t.Errorf("%v should be linked with %v; but %t", r1, r2, got)
	}
}

func TestLinks(t *testing.T) {
	r1, r2 := NewRoom(), NewRoom()
	r1.Link(&r2)

	got := len(r1.Links())
	if got != 1 {
		t.Errorf("# of linked rooms of %v should be 1; but got %d", r1, got)
	}
}

func TestShuffle(t *testing.T) {
	// seed by which rand.Perm() returns indices of reverse order
	rand.Seed(20)

	tests := []struct {
		rooms []*Room
		want  bool
	}{
		{[]*Room{}, true},
		{[]*Room{&Room{Walls: Survey{}}}, true},
		{[]*Room{&Room{Walls: Survey{Top: true}}, &Room{Walls: Survey{Bottom: true}}}, false},
	}

	for _, tt := range tests {
		shfl := Shuffle(tt.rooms)
		if len(shfl) != len(tt.rooms) {
			t.Errorf("length: got %d but want %d", len(shfl), len(tt.rooms))
		}
		if reflect.DeepEqual(shfl, tt.rooms) != tt.want {
			t.Errorf("equality: equality of %v and %v should not be %t", shfl, tt.rooms, tt.want)
		}
	}
}

func TestRandom(t *testing.T) {
	rand.Seed(3)

	rooms := []*Room{&Room{}, &Room{}, &Room{}}
	tests := []struct {
		rooms []*Room
		want  *Room
	}{
		{nil, nil},
		{[]*Room{}, nil},
		{rooms[:1], rooms[0]},
		{rooms, rooms[2]},
	}

	for _, tt := range tests {
		got := Random(tt.rooms)
		// compare with respect to pointer address
		if got != tt.want {
			t.Errorf("%v: got %p but want %p", tt.rooms, got, tt.want)
		}
	}
}
