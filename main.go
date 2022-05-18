package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/danicat/simpleansi"
)

// Config holds the emoji configuration
type Config struct {
	Alien     string        `json:"alien"`
	Wall      string        `json:"wall"`
	Pill      string        `json:"pill"`
	Death     string        `json:"death"`
	Space     string        `json:"space"`
	Fire      string        `json:"fire"`
	Ash       string        `json:"ash"`
	River     string        `json:"river"`
	Bridge    string        `json:"bridge"`
	Tree      string        `json:"tree"`
	Clock     string        `json:"clock"`
	Time      time.Duration `json:"time"`
	Duration  int           `json:"duration"`
	DebugMode bool          `json:"debug_mode"`
	Map       string        `json:"map"`
}

type Sprite struct {
	row int
	col int
	num int
}

type Flag struct {
	Emoji       string `json:"emoji"`
	Description string `json:"description"`
}

var cfg Config
var flags []Flag
var grid []string
var maps []string
var logBook []string
var aliens []*Sprite
var space int
var numAliens int
var nu int
var mapName string

var duration = cfg.Duration

func initialise() {
	cbTerm := exec.Command("stty", "cbreak", "-echo")
	cbTerm.Stdin = os.Stdin

	err := cbTerm.Run()
	if err != nil {
		log.Fatalln("unable to activate cbreak mode:", err)
	}
}

func cleanup() {
	cookedTerm := exec.Command("stty", "-cbreak", "echo")
	cookedTerm.Stdin = os.Stdin

	err := cookedTerm.Run()
	if err != nil {
		log.Fatalln("unable to activate cooked mode:", err)
	}
}

func loadFile(file string, store []string) ([]string, error) {
	l, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer l.Close()

	scanner := bufio.NewScanner(l)
	for scanner.Scan() {
		line := scanner.Text()
		store = append(store, line)
	}

	return store, nil
}

func airdropAliens() {
	var i = 0

	// for len(aliens) != numAliens {
	for numAliens > 0 {
		for row, line := range grid {
			for col, char := range line {
				switch char {
				case ' ':
					if numAliens != 0 && rand.Intn(2) == 1 {
						if row > 1 && col > 1 && row < len(grid)-2 && col < len(line)-2 {
							i++
							numAliens--
							aliens = append(aliens, &Sprite{row, col, i})
						}
					}
				}
			}
		}
	}
}

func countAvailableSpace() {
	for _, line := range grid {
		for _, char := range line {
			switch char {
			case ' ':
				space++
			}
		}
	}
	space = space / 2
}

func printScreen() {
	simpleansi.ClearScreen()
	for _, line := range grid {
		for _, chr := range line {
			switch chr {
			case '#':
				fmt.Print(simpleansi.WithBlueBackground(cfg.Wall))
			case '@':
				fmt.Print(cfg.Fire)
			case '!':
				fmt.Print(cfg.River)
			case '$':
				fmt.Print(cfg.Clock)
			case 'T':
				fmt.Print(cfg.Tree)
			case 'B':
				fmt.Print(cfg.Bridge)
			case 'A':
				fmt.Print(cfg.Ash)
			default:
				fmt.Print(cfg.Space)
			}
		}
		fmt.Println()
	}

	for _, g := range aliens {
		moveCursor(g.row, g.col)
		fmt.Print(cfg.Alien)
	}

	// Move info outside of simulation area
	moveCursor(len(grid)+1, 0)
	fmt.Println("\tMap:", mapName)
	fmt.Println("\tDuration:", duration)
	fmt.Println("\tNumber of aliens alive:", len(aliens))
}

func readInput() (string, error) {
	buffer := make([]byte, 100)

	cnt, err := os.Stdin.Read(buffer)
	if err != nil {
		return "", err
	}

	if cnt == 1 && buffer[0] == 0x1b {
		return "ESC", nil
	} else if cnt >= 3 {
		if buffer[0] == 0x1b && buffer[1] == '[' {
			switch buffer[2] {
			case 'A':
				return "UP", nil
			case 'B':
				return "DOWN", nil
			case 'C':
				return "RIGHT", nil
			case 'D':
				return "LEFT", nil
			}
		}
	}

	return "", nil
}

func makeMove(oldRow, oldCol int, dir string) (newRow, newCol int) {
	newRow, newCol = oldRow, oldCol

	switch dir {
	case "UP":
		newRow = newRow - 1
		if newRow < 0 {
			newRow = len(grid) - 1
		}
	case "DOWN":
		newRow = newRow + 1
		if newRow == len(grid) {
			newRow = 0
		}
	case "RIGHT":
		newCol = newCol + 1
		if newCol == len(grid[0]) {
			newCol = 0
		}
	case "LEFT":
		newCol = newCol - 1
		if newCol < 0 {
			newCol = len(grid[0]) - 1
		}
	}

	if grid[newRow][newCol] == '#' || grid[newRow][newCol] == '@' || grid[newRow][newCol] == '!' || grid[newRow][newCol] == '$' || grid[newRow][newCol] == 'T' || grid[newRow][newCol] == 'A' {
		newRow = oldRow
		newCol = oldCol
	}

	return
}

func drawDirection() string {
	dir := rand.Intn(4)
	move := map[int]string{
		0: "UP",
		1: "DOWN",
		2: "RIGHT",
		3: "LEFT",
	}
	return move[dir]
}

func moveAliens() {
	for _, g := range aliens {
		dir := drawDirection()
		g.row, g.col = makeMove(g.row, g.col, dir)
	}
}

func loadConfig(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		return err
	}

	return nil
}

func loadFlags(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	err = decoder.Decode(&flags)
	if err != nil {
		return err
	}

	return nil
}

func moveCursor(row, col int) {
	simpleansi.MoveCursor(row, col*2)
}

func RemoveAlien(s []*Sprite, index int) []*Sprite {
	return append(s[:index], s[index+1:]...)
}

func findDuplicates() {
	stack := []*Sprite{}
	for i := 0; i < len(aliens); i++ {
		if len(stack) != 0 && (stack[len(stack)-1].row == aliens[i].row && stack[len(stack)-1].col == aliens[i].col) {
			//remove the duplicated aliens and burn destroy the city
			var j = rand.Intn(len(flags))
			log.Println("[", flags[j].Emoji, flags[j].Description, "]", "is destroyed by Alien", stack[len(stack)-1].num, " and Alien", aliens[i].num)
			r := aliens[i].row
			c := aliens[i].col
			stack = stack[:len(stack)-1]
			grid[r] = grid[r][0:c] + "@" + grid[r][c+1:]

			// make trees and bridges turn into ashes when next to fire
			if grid[r+1][c] == 84 || grid[r+1][c] == 66 {
				grid[r+1] = grid[r+1][0:c] + "A" + grid[r+1][c+1:]
			}
			if grid[r-1][c] == 84 || grid[r-1][c] == 66 {
				grid[r-1] = grid[r-1][0:c] + "A" + grid[r-1][c+1:]
			}
			if grid[r][c+1] == 84 || grid[r][c+1] == 66 {
				grid[r] = grid[r][0:c+1] + "A" + grid[r][c+2:]
			}
			if grid[r][c-1] == 84 || grid[r][c-1] == 66 {
				grid[r] = grid[r][0:c-1] + "A" + grid[r][c:]
			}
		} else {
			//temporarily save the location of the alien
			stack = append(stack, aliens[i])
		}
	}
	aliens = stack
}

func main() {
	// load the game settings
	err := loadConfig("configs/config.json")
	if err != nil {
		log.Fatal("failed to load configuration:", err)
		return
	}

	duration = cfg.Duration

	// checking config.json's debug mode boolean value
	var debugMode bool
	flag.BoolVar(&debugMode, "debug", cfg.DebugMode, "Runs with debug mode.")
	flag.Parse()

	if debugMode == false {
		// load maps
		files, err := ioutil.ReadDir("./map")
		if err != nil {
			log.Fatal(err)
		}

		for _, file := range files {
			name := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			maps = append(maps, name)
		}

		// select map
		mapSelected := false
		for mapSelected == false {
			for i, _ := range maps {
				fmt.Println(i, maps[i])
			}
			fmt.Print("choose the map : ")
			_, err := fmt.Scanf("%d", &nu)
			if err != nil || nu < 0 || nu > len(maps)-1 {
				fmt.Println("please try again")
			} else {
				mapName = maps[nu]
				mapSelected = true
			}
		}

		// load the map
		grid, err = loadFile("map/"+mapName+".txt", grid)
		if err != nil {
			log.Fatal("failed to load map:", err)
		}

		// count available spaces to airdrop aliens
		countAvailableSpace()

		// initialise cbreak mode
		initialise()
		defer cleanup()

		// checking user input
		for numAliens == 0 {
			fmt.Printf("number of aliens (number should be less than %v) : ", space)
			_, err := fmt.Scanf("%d", &numAliens)
			if err != nil || numAliens >= space {
				fmt.Println("please try again")
				numAliens = 0
			}
		}
	} else {
		// load the map
		grid, err = loadFile("map/"+cfg.Map+".txt", grid)
		if err != nil {
			log.Fatal("failed to load map:", err)
		}

		countAvailableSpace()

		numAliens = space - 1
	}

	// checks if history file exists and deletes it
	if _, err := os.Stat("history/logBook.txt"); err == nil {
		os.Remove("history/logBook.txt")
	}

	// get region information
	err = loadFlags("configs/flags.json")
	if err != nil {
		log.Fatal("failed to load flags:", err)
	}

	// creates a new history file to record the logs
	f, err := os.OpenFile("history/logBook.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	// deploying aliens randomly
	airdropAliens()

	// process input (async)
	input := make(chan string)
	go func(ch chan<- string) {
		for {
			input, err := readInput()
			if err != nil {
				log.Println("error reading input:", err)
				ch <- "ESC"
			}
			ch <- input
		}
	}(input)

	for {
		// decrement time duration
		duration--

		// ESC to kill the simulation
		select {
		case inp := <-input:
			if inp == "ESC" {
				duration = 0
			}
		default:
		}

		// randomly move aliens
		moveAliens()

		// process collisions
		findDuplicates()

		// update screen
		printScreen()

		// repeat
		time.Sleep(cfg.Time * time.Millisecond)

		// check game over
		if duration <= 0 || len(aliens) <= 0 {
			logBook, err = loadFile("history/logBook.txt", logBook)
			for _, line := range logBook {
				fmt.Println(line)
			}
			fmt.Println("\tGame Over")
			break
		}
	}

}
