// Copyright © 2015 Steve Francia <spf@spf13.com>.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.
//

// This is a small set of interfaces and utilities designed to help
// with the Go Challenge 6: Daedalus & Icarus

package mazelib

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Coordinate describes a location in the maze
type Coordinate struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Reply from the server to a request
type Reply struct {
	Survey  Survey `json:"survey"`
	Victory bool   `json:"victory"`
	Message string `json:"message"`
	Error   bool   `json:"error"`
}

// Survey Given a location, survey surrounding locations
// True indicates a wall is present.
type Survey struct {
	Top    bool `json:"top"`
	Right  bool `json:"right"`
	Bottom bool `json:"bottom"`
	Left   bool `json:"left"`
}

// Direction is a direction.
type Direction int

// N, S, E, W are directions to north, south, east and west.
const (
	N Direction = 1
	S           = 2
	E           = 3
	W           = 4
)

// ErrVictory is an error representing the victory of Icarus.
var ErrVictory = errors.New("Victory")

// Room contains the minimum informaion about a room in the maze.
type Room struct {
	Treasure bool
	Start    bool
	Visited  bool
	Walls    Survey
	Nbr      map[*Room]Direction
	links    map[*Room]bool
}

// NewRoom creates a new Room.
func NewRoom() Room {
	return Room{
		Nbr:   make(map[*Room]Direction),
		links: make(map[*Room]bool),
	}
}

// AddWall adds a wall in the `dir` direction.
func (r *Room) AddWall(dir Direction) {
	switch dir {
	case N:
		r.Walls.Top = true
	case S:
		r.Walls.Bottom = true
	case E:
		r.Walls.Right = true
	case W:
		r.Walls.Left = true
	}
}

// RmWall removes a wall in the `dir` direction.
func (r *Room) RmWall(dir Direction) {
	switch dir {
	case N:
		r.Walls.Top = false
	case S:
		r.Walls.Bottom = false
	case E:
		r.Walls.Right = false
	case W:
		r.Walls.Left = false
	}
}

// Link links r with room, e.g. removes face-to-face walls.
func (r *Room) Link(room *Room) {
	r.link(room, true)
}

func (r *Room) link(room *Room, bidi bool) {
	r.RmWall(r.Nbr[room])
	r.links[room] = true
	if bidi {
		room.RmWall(room.Nbr[r])
		room.link(r, false)
	}
}

// Links returns all the rooms linked with r.
func (r *Room) Links() []*Room {
	l := make([]*Room, 0, len(r.links))
	for k := range r.links {
		l = append(l, k)
	}
	return l
}

// Neighbors returns all the neighbors around `r`.
func (r *Room) Neighbors() []*Room {
	nbrs := make([]*Room, 0, len(r.Nbr))
	for nbr := range r.Nbr {
		nbrs = append(nbrs, nbr)
	}
	return nbrs
}

// MazeI Interface
type MazeI interface {
	GetRoom(x, y int) (*Room, error)
	Width() int
	Height() int
	SetStartPoint(x, y int) error
	SetTreasure(x, y int) error
	LookAround() (Survey, error)
	Discover(x, y int) (Survey, error)
	Icarus() (x, y int)
	MoveLeft() error
	MoveRight() error
	MoveUp() error
	MoveDown() error
}

// AvgScores calculates the avarage of `in`.
func AvgScores(in []int) int {
	if len(in) == 0 {
		return 0
	}

	total := 0

	for _, x := range in {
		total += x
	}
	return total / (len(in))
}

// PrintMaze : Function to Print Maze to Console
func PrintMaze(m MazeI) {
	ix, iy := m.Icarus()

	fmt.Println("_" + strings.Repeat("___", m.Width()))
	for y := 0; y < m.Height(); y++ {
		str := ""
		for x := 0; x < m.Width(); x++ {
			if x == 0 {
				str += "|"
			}
			r, err := m.GetRoom(x, y)
			if err != nil {
				fmt.Println(err)
				os.Exit(-1)
			}
			s, err := m.Discover(x, y)
			if err != nil {
				fmt.Println(err)
				os.Exit(-1)
			}
			if s.Bottom {
				if r.Treasure {
					str += "⏅_"
				} else if r.Start {
					str += "⏂_"
				} else if ix == x && iy == y {
					str += "⏈ "
				} else {
					str += "__"
				}
			} else {
				if r.Treasure {
					str += "⏃ "
				} else if r.Start {
					str += "⏀ "
				} else if ix == x && iy == y {
					str += "⏆ "
				} else {
					str += "  "
				}
			}

			if s.Right {
				str += "|"
			} else {
				str += "_"
			}

		}
		fmt.Println(str)
	}
}
