package mazelib

import (
	"math/rand"
	"reflect"
	"testing"
)

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
