// generated by stringer -type=Direction; DO NOT EDIT

package mazelib

import "fmt"

const _Direction_name = "NSEW"

var _Direction_index = [...]uint8{0, 1, 2, 3, 4}

func (i Direction) String() string {
	i -= 1
	if i < 0 || i >= Direction(len(_Direction_index)-1) {
		return fmt.Sprintf("Direction(%d)", i+1)
	}
	return _Direction_name[_Direction_index[i]:_Direction_index[i+1]]
}