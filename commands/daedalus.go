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

package commands

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/skatsuta/labyrinth/log"
	"github.com/skatsuta/labyrinth/mazelib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Maze is a maze.
type Maze struct {
	rooms      [][]mazelib.Room
	start      mazelib.Coordinate
	end        mazelib.Coordinate
	icarus     mazelib.Coordinate
	StepsTaken int
}

// Tracking the current maze being solved

// WARNING: This approach is not safe for concurrent use
// This server is only intended to have a single client at a time
// We would need a different and more complex approach if we wanted
// concurrent connections than these simple package variables
var currentMaze *Maze
var scores []int
var debug bool

// Defining the daedalus command.
// This will be called as 'laybrinth daedalus'
var daedalusCmd = &cobra.Command{
	Use:     "daedalus",
	Aliases: []string{"deadalus", "server"},
	Short:   "Start the laybrinth creator",
	Long: `Daedalus's job is to create a challenging Labyrinth for his opponent
  Icarus to solve.

  Daedalus runs a server which Icarus clients can connect to to solve laybrinths.`,
	Run: func(cmd *cobra.Command, args []string) {
		RunServer()
	},
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano()) // need to initialize the seed
	gin.SetMode(gin.ReleaseMode)

	RootCmd.AddCommand(daedalusCmd)
}

// RunServer runs the web server.
func RunServer() {
	// Adding handling so that even when ctrl+c is pressed we still print
	// out the results prior to exiting.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		printResults()
		os.Exit(1)
	}()

	// Using gin-gonic/gin to handle our routing
	r := gin.Default()
	v1 := r.Group("/")
	{
		v1.GET("/awake", GetStartingPoint)
		v1.GET("/move/:direction", MoveDirection)
		v1.GET("/done", End)
	}

	if e := r.Run(":" + viper.GetString("port")); e != nil {
		panic(e)
	}
}

// End ends a session and prints the results.
// Called by Icarus when he has reached
//   the number of times he wants to solve the laybrinth.
func End(c *gin.Context) {
	printResults()
	os.Exit(1)
}

// GetStartingPoint initializes a new maze and places Icarus in his awakening location
func GetStartingPoint(c *gin.Context) {
	ySize := viper.GetInt("height")
	xSize := viper.GetInt("width")
	initializeMaze(xSize, ySize)
	startRoom, err := currentMaze.Discover(currentMaze.Icarus())
	if err != nil {
		fmt.Println("Icarus is outside of the maze. This shouldn't ever happen")
		fmt.Println(err)
		os.Exit(-1)
	}
	mazelib.PrintMaze(currentMaze)

	c.JSON(http.StatusOK, mazelib.Reply{Survey: startRoom})
}

// MoveDirection returns the API response to the /move/:direction address
func MoveDirection(c *gin.Context) {
	var err error

	switch c.Param("direction") {
	case "left":
		err = currentMaze.MoveLeft()
	case "right":
		err = currentMaze.MoveRight()
	case "down":
		err = currentMaze.MoveDown()
	case "up":
		err = currentMaze.MoveUp()
	}

	var r mazelib.Reply

	if err != nil {
		r.Error = true
		r.Message = err.Error()
		c.JSON(409, r)
		return
	}

	s, e := currentMaze.LookAround()

	if e != nil {
		if e == mazelib.ErrVictory {
			scores = append(scores, currentMaze.StepsTaken)
			r.Victory = true
			r.Message = fmt.Sprintf("Victory achieved in %d steps \n", currentMaze.StepsTaken)
		} else {
			r.Error = true
			r.Message = err.Error()
		}
	}

	r.Survey = s

	c.JSON(http.StatusOK, r)

	if viper.GetBool("debug") {
		mazelib.PrintMaze(currentMaze)
	}
}

func initializeMaze(x, y int) {
	currentMaze = createMaze(x, y)
}

// Print to the terminal the average steps to solution for the current session
func printResults() {
	fmt.Printf("Labyrinth solved %d times with an avg of %d steps\n", len(scores), mazelib.AvgScores(scores))
}

// GetRoom returns a room from the maze
func (m *Maze) GetRoom(x, y int) (*mazelib.Room, error) {
	if x < 0 || y < 0 || x >= m.Width() || y >= m.Height() {
		return &mazelib.Room{}, errors.New("room outside of maze boundaries")
	}

	return &m.rooms[y][x], nil
}

// Width returns the width of a maze.
func (m *Maze) Width() int { return len(m.rooms[0]) }

// Height returns the height of a maze.
func (m *Maze) Height() int { return len(m.rooms) }

// Icarus returns Icarus's current position
func (m *Maze) Icarus() (x, y int) {
	return m.icarus.X, m.icarus.Y
}

// SetStartPoint sets the location where Icarus will awake
func (m *Maze) SetStartPoint(x, y int) error {
	r, err := m.GetRoom(x, y)

	if err != nil {
		return err
	}

	if r.Treasure {
		return errors.New("can't start in the treasure")
	}

	r.Start = true
	m.icarus = mazelib.Coordinate{x, y}
	return nil
}

// SetTreasure sets the location of the treasure for a given maze
func (m *Maze) SetTreasure(x, y int) error {
	r, err := m.GetRoom(x, y)

	if err != nil {
		return err
	}

	if r.Start {
		return errors.New("can't have the treasure at the start")
	}

	r.Treasure = true
	m.end = mazelib.Coordinate{x, y}
	return nil
}

// LookAround discovers that room when given Icarus's current location.
// It will return ErrVictory if Icarus is at the treasure.
func (m *Maze) LookAround() (mazelib.Survey, error) {
	if m.end.X == m.icarus.X && m.end.Y == m.icarus.Y {
		fmt.Printf("Victory achieved in %d steps \n", m.StepsTaken)
		return mazelib.Survey{}, mazelib.ErrVictory
	}

	return m.Discover(m.icarus.X, m.icarus.Y)
}

// Discover survey the room when given two points.
// It will return error if two points are outside of the maze
func (m *Maze) Discover(x, y int) (mazelib.Survey, error) {
	r, err := m.GetRoom(x, y)
	if err != nil {
		return mazelib.Survey{}, nil
	}
	return r.Walls, nil
}

// MoveLeft moves Icarus's position left one step
// Will not permit moving through walls or out of the maze
func (m *Maze) MoveLeft() error {
	s, e := m.LookAround()
	if e != nil {
		return e
	}
	if s.Left {
		return errors.New("Can't walk through walls")
	}

	x, y := m.Icarus()
	if _, err := m.GetRoom(x-1, y); err != nil {
		return err
	}

	m.icarus = mazelib.Coordinate{x - 1, y}
	m.StepsTaken++
	return nil
}

// MoveRight moves Icarus's position right one step
// Will not permit moving through walls or out of the maze
func (m *Maze) MoveRight() error {
	s, e := m.LookAround()
	if e != nil {
		return e
	}
	if s.Right {
		return errors.New("Can't walk through walls")
	}

	x, y := m.Icarus()
	if _, err := m.GetRoom(x+1, y); err != nil {
		return err
	}

	m.icarus = mazelib.Coordinate{x + 1, y}
	m.StepsTaken++
	return nil
}

// MoveUp moves Icarus's position up one step
// Will not permit moving through walls or out of the maze
func (m *Maze) MoveUp() error {
	s, e := m.LookAround()
	if e != nil {
		return e
	}
	if s.Top {
		return errors.New("Can't walk through walls")
	}

	x, y := m.Icarus()
	if _, err := m.GetRoom(x, y-1); err != nil {
		return err
	}

	m.icarus = mazelib.Coordinate{x, y - 1}
	m.StepsTaken++
	return nil
}

// MoveDown moves Icarus's position down one step
// Will not permit moving through walls or out of the maze
func (m *Maze) MoveDown() error {
	s, e := m.LookAround()
	if e != nil {
		return e
	}
	if s.Bottom {
		return errors.New("Can't walk through walls")
	}

	x, y := m.Icarus()
	if _, err := m.GetRoom(x, y+1); err != nil {
		return err
	}

	m.icarus = mazelib.Coordinate{x, y + 1}
	m.StepsTaken++
	return nil
}

// AllRooms returns all the Rooms in the Maze.
func (m *Maze) AllRooms() []*mazelib.Room {
	size := m.Width() * m.Height()
	rooms := make([]*mazelib.Room, 0, size)
	for y, row := range m.rooms {
		for x := range row {
			rooms = append(rooms, &m.rooms[y][x])
		}
	}
	return rooms
}

// DeadEnds returns all the dead-end Rooms in the Maze.
func (m *Maze) DeadEnds() []*mazelib.Room {
	var list []*mazelib.Room

	for _, room := range m.AllRooms() {
		if len(room.Links()) == 1 {
			list = append(list, room)
		}
	}

	return list
}

// Braid rearranges the Maze to "braid" one, that is, a maze without dead ends.
// p is the probability for the occurrence of braids. If p <= 0.0, it does nothing.
func (m *Maze) Braid(p float64) {
	for _, room := range mazelib.Shuffle(m.AllRooms()) {
		if len(room.Links()) != 1 || rand.Float64() > p {
			continue
		}

		var nbs, best []*mazelib.Room

		for _, nb := range room.Neighbors() {
			if !nb.IsLinked(room) {
				nbs = append(nbs, nb)
			}
		}

		for _, nb := range nbs {
			if len(nb.Links()) == 1 {
				best = append(best, nb)
			}
		}

		if len(best) == 0 {
			best = nbs
		}

		room.Link(mazelib.Random(best))
	}
}

// Creates a maze without any walls
// Good starting point for additive algorithms
func emptyMaze(xSize, ySize int) *Maze {
	z := Maze{}

	z.rooms = make([][]mazelib.Room, ySize)
	for y := 0; y < ySize; y++ {
		z.rooms[y] = make([]mazelib.Room, xSize)
		for x := 0; x < xSize; x++ {
			z.rooms[y][x] = mazelib.NewRoom()
		}
	}

	configureRooms(&z)

	return &z
}

func configureRooms(z *Maze) {
	dirs := []mazelib.Direction{mazelib.N, mazelib.E, mazelib.S, mazelib.W}

	w, h := z.Width(), z.Height()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			room, err := z.GetRoom(x, y)
			if err != nil {
				continue
			}

			// init Nbr field
			room.Nbr = make(map[*mazelib.Room]mazelib.Direction)

			// north, east, south, west
			coords := [][]int{{x, y - 1}, {x + 1, y}, {x, y + 1}, {x - 1, y}}
			for i, coord := range coords {
				if nbr, err := z.GetRoom(coord[0], coord[1]); err == nil {
					room.Nbr[nbr] = dirs[i]
				}
			}
		}
	}
}

// Creates a maze with all walls
// Good starting point for subtractive algorithms
func fullMaze(xSize, ySize int) *Maze {
	z := emptyMaze(xSize, ySize)

	for y := 0; y < ySize; y++ {
		for x := 0; x < xSize; x++ {
			z.rooms[y][x].Walls = mazelib.Survey{true, true, true, true}
		}
	}

	return z
}

func createMaze(xSize, ySize int) *Maze {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	z := recursiveBacktracker(xSize, ySize)

	z.Braid(viper.GetFloat64("braid"))

	// set the starting point and goal randomly
	w, h := z.Width(), z.Height()
	sx, sy := r.Intn(w), r.Intn(h)
	if e := z.SetStartPoint(sx, sy); e != nil {
		log.Errorf("error setting start point: %v\n", e)
		return emptyMaze(xSize, ySize)
	}

	tx, ty := r.Intn(w), r.Intn(h)
	for sx == tx && sy == ty {
		// retry
		tx, ty = r.Intn(w), r.Intn(h)
	}
	if e := z.SetTreasure(tx, ty); e != nil {
		log.Errorf("error setting treasure: %v\n", e)
		return emptyMaze(xSize, ySize)
	}

	return z
}

// recursiveBacktracker creates a maze by using recursive backtracker algorithm.
func recursiveBacktracker(xSize, ySize int) *Maze {
	z := fullMaze(xSize, ySize)

	// pick a starting Room randomly
	w, h := z.Width(), z.Height()
	start, err := z.GetRoom(rand.Intn(w), rand.Intn(h))
	if err != nil {
		start, _ = z.GetRoom(0, 0)
	}

	stack := []*mazelib.Room{start}

	for len(stack) > 0 {
		current := stack[len(stack)-1]

		var nbs []*mazelib.Room
		for _, nb := range current.Neighbors() {
			if len(nb.Links()) == 0 {
				nbs = append(nbs, nb)
			}
		}

		if len(nbs) == 0 {
			stack = stack[:len(stack)-1]
			continue
		}

		nb := mazelib.Random(nbs)
		current.Link(nb)
		stack = append(stack, nb)
	}

	return z
}
