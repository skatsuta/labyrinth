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
