package commands

import (
	"reflect"
	"testing"

	"github.com/skatsuta/labyrinth/mazelib"
)

func TestPop(t *testing.T) {
	r := []record{
		{dirs: []mazelib.Direction{mazelib.N}},
		{dirs: []mazelib.Direction{mazelib.S}},
	}
	stk := newStack(r...)

	got1 := stk.pop()
	if stk.size() != 1 {
		t.Errorf("1 pop: expected size = %d, but got %d", 1, stk.size())
	}
	if !reflect.DeepEqual(got1, r[1]) {
		t.Errorf("1 pop: %v != %v", got1, r[1])
	}

	got2 := stk.pop()
	if stk.size() != 0 {
		t.Errorf("2 pop: expected size = %d, but got %d", 0, stk.size())
	}
	if !reflect.DeepEqual(got2, r[0]) {
		t.Errorf("1 pop: %v != %v", got2, r[0])
	}
}

func TestLast(t *testing.T) {
	r := []record{{dirs: []mazelib.Direction{mazelib.E}}}

	tests := []struct {
		r    []record
		want *record
	}{
		{[]record{}, nil},
		{r, &r[0]},
	}

	for _, tt := range tests {
		stk := newStack(tt.r...)
		got := stk.last()
		if got != tt.want {
			t.Errorf("got %v but want %v", got, tt.want)
		}
	}
}
