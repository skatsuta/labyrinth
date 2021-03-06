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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/skatsuta/labyrinth/log"
	"github.com/skatsuta/labyrinth/mazelib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Defining the icarus command.
// This will be called as 'laybrinth icarus'
var icarusCmd = &cobra.Command{
	Use:     "icarus",
	Aliases: []string{"client"},
	Short:   "Start the laybrinth solver",
	Long: `Icarus wakes up to find himself in the middle of a Labyrinth.
  Due to the darkness of the Labyrinth he can only see his immediate cell and if
  there is a wall or not to the top, right, bottom and left. He takes one step
  and then can discover if his new cell has walls on each of the four sides.

  Icarus can connect to a Daedalus and solve many laybrinths at a time.`,
	Run: func(cmd *cobra.Command, args []string) {
		RunIcarus()
	},
}

func init() {
	RootCmd.AddCommand(icarusCmd)
}

// RunIcarus runs the solver as many times as the user desires.
func RunIcarus() {
	// Run the solver as many times as the user desires.
	fmt.Println("Solving", viper.GetInt("times"), "times")
	for x := 0; x < viper.GetInt("times"); x++ {

		solveMaze()
	}

	// Once we have solved the maze the required times, tell daedalus we are done
	_, _ = makeRequest("http://127.0.0.1:" + viper.GetString("port") + "/done")
}

// Make a call to the laybrinth server (daedalus) that icarus is ready to wake up
func awake() mazelib.Survey {
	contents, err := makeRequest("http://127.0.0.1:" + viper.GetString("port") + "/awake")
	if err != nil {
		fmt.Println(err)
	}
	r := ToReply(contents)
	return r.Survey
}

// Move makes a call to the laybrinth server (daedalus)
// to move Icarus a given direction
// Will be used heavily by solveMaze
func Move(direction string) (mazelib.Survey, error) {
	if direction == "left" || direction == "right" || direction == "up" || direction == "down" {

		contents, err := makeRequest("http://127.0.0.1:" + viper.GetString("port") + "/move/" + direction)
		if err != nil {
			return mazelib.Survey{}, err
		}

		rep := ToReply(contents)
		if rep.Victory {
			fmt.Println(rep.Message)
			// os.Exit(1)
			return rep.Survey, mazelib.ErrVictory
		}
		return rep.Survey, errors.New(rep.Message)
	}

	return mazelib.Survey{}, errors.New("invalid direction")
}

// utility function to wrap making requests to the daedalus server
func makeRequest(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response.Body.Close()
	}()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return contents, nil
}

// ToReply handles a JSON response and unmarshalling it into a reply struct
func ToReply(in []byte) mazelib.Reply {
	res := &mazelib.Reply{}
	_ = json.Unmarshal(in, &res)
	return *res
}

func solveMaze() {
	var (
		sv          mazelib.Survey
		dir         mazelib.Direction
		err         error
		s           = awake()
		stack       = newStack(record{survey: s})
		popped      bool
		count       int
		interactive = viper.GetBool("interactive")
	)

	for stack.size() > 0 {
		if interactive {
			var input string
			fmt.Print("Press Enter to move forward...")
			fmt.Scanln(&input)
		}

		count++
		log.Debugf("count: %d\n", count)

		popped = false
		current := stack.last()
		log.Debugf("current: %+v\n", current)

		// init
		cand := make(map[mazelib.Direction]bool)
		if !current.survey.Top {
			cand[mazelib.N] = true
		}
		if !current.survey.Bottom {
			cand[mazelib.S] = true
		}
		if !current.survey.Right {
			cand[mazelib.E] = true
		}
		if !current.survey.Left {
			cand[mazelib.W] = true
		}
		log.Debugf("direction candidates are %v\n", cand)

		// delete the directions Icarus has already moved to unless it's a dead end
		for _, d := range current.dirs {
			if cand[d] {
				log.Debugf("direction %s has been already moved to. deleting...\n", d.String())
				delete(cand, d)
			}
		}

		if len(cand) == 0 {
			switch len(current.dirs) {
			case 0:
				log.Warnf("no direction to move on! giving up...\n")
				return
			default: // move to the oldest direction
				cand[current.dirs[0]] = true
			}

			stack.pop()
			log.Debugf("popping from the stack: size = %d\n", stack.size())
			popped = true
		}

		// sampling
		for dir = range cand {
			sv, err = Move(dir.String())
			break
		}
		log.Debugf("next: %+v\n", sv)
		if err == mazelib.ErrVictory {
			log.Infof("Yay! Treasure discovered!\n")
			return
		}
		if err.Error() != "" {
			log.Debugf("error: %#v\n", err)
			return
		}

		if popped {
			continue
		}

		// record the direction Icarus moved to
		current.dirs = append(current.dirs, dir)

		dirs := []mazelib.Direction{dir.Opposite()}
		// push to stack
		next := record{survey: sv, dirs: dirs}
		stack.push(next)
	}

	log.Warnf("stack is now empty... maybe something wrong?\n")
}

// record is a record of directions Icarus moved to.
type record struct {
	survey mazelib.Survey
	dirs   []mazelib.Direction
}

// stack is a stack of records.
type stack struct {
	stk []record
}

// newStack returns a new empty stack.
func newStack(rec ...record) *stack {
	return &stack{
		stk: rec,
	}
}

func (s *stack) isEmpty() bool {
	return len(s.stk) == 0
}

func (s *stack) size() int {
	return len(s.stk)
}

func (s *stack) push(r record) {
	s.stk = append(s.stk, r)
}

func (s *stack) pop() record {
	if len(s.stk) == 0 {
		return record{}
	}
	r := s.stk[len(s.stk)-1]
	s.stk = s.stk[:len(s.stk)-1]
	return r
}

func (s *stack) last() *record {
	if len(s.stk) == 0 {
		return nil
	}
	return &s.stk[len(s.stk)-1]
}
