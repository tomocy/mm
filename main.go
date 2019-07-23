package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func init() {
	if err := activateCBTerm(); err != nil {
		log.Fatalf("failed to activate cbreak terminal: %s\n", err)
	}
}

func activateCBTerm() error {
	cmd := exec.Command("/bin/stty", "cbreak", "-echo")
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func main() {
	defer cleanUp()
	game := new(game)
	if err := game.load("./maze.txt"); err != nil {
		log.Fatalf("failed for game to load: %s\n", err)
	}

	if err := game.start(); err != nil {
		log.Printf("failed for game to start: %s\n", err)
	}
}

func cleanUp() {
	if err := activateCookedTerm(); err != nil {
		log.Fatalf("failed to activate cooked termainal: %s\n", err)
	}
}

func activateCookedTerm() error {
	cmd := exec.Command("/bin/stty", "-cbreak", "echo")
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

type game struct {
	maze   maze
	player player
}

func (g *game) load(name string) error {
	var err error
	g.maze, err = loadMaze(name)
	if err != nil {
		return err
	}

	g.player = g.maze.findPlayer()

	return nil
}

func loadMaze(name string) (maze, error) {
	src, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	var maze maze
	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		line := scanner.Text()
		maze = append(maze, line)
	}

	return maze, nil
}

func (g *game) start() error {
	for {
		printScreen(g.maze)

		key, err := readInput()
		if err != nil {
			return err
		}
		if key == keyEsc {
			break
		}
	}

	return nil
}

func printScreen(maze maze) {
	cleanScreen()
	fmt.Print(maze)
}

func cleanScreen() {
	fmt.Print("\x1b[2J")
	moveCursor(0, 0)
}

func moveCursor(row, col int) {
	fmt.Printf("\x1b[%d;%df", row, col)
}

type maze []string

func (m maze) String() string {
	buf := new(bytes.Buffer)
	for _, line := range m {
		fmt.Fprintf(buf, "%s\n", line)
	}

	return buf.String()
}

func (m maze) findPlayer() player {
	for row, line := range m {
		for col, level := range line {
			if string(level) == levelPlayer {
				return player{
					position: position{
						row: row,
						col: col,
					},
				}
			}
		}
	}

	return player{}
}

const (
	levelPlayer = "P"
)

func readInput() (string, error) {
	buf := make([]byte, 10)
	cnt, err := os.Stdin.Read(buf)
	if err != nil {
		return "", err
	}

	if cnt == 1 && buf[0] == 0x1b {
		return keyEsc, nil
	}
	if 3 <= cnt && buf[0] == 0x1b && buf[1] == '[' {
		switch buf[2] {
		case 'A':
			return keyUp, nil
		case 'B':
			return keyDown, nil
		case 'C':
			return keyRight, nil
		case 'D':
			return keyLeft, nil
		}
	}

	return string(buf[:cnt]), nil
}

const (
	keyEsc   = "ecs"
	keyUp    = "up"
	keyDown  = "down"
	keyRight = "right"
	keyLeft  = "left"
)

type player struct {
	position position
}

type position struct {
	row, col int
}

func (p *player) move(maze maze, direction string) {
	p.position.row, p.position.col = makeMove(maze, direction, p.position.row, p.position.col)
}

func makeMove(maze maze, direction string, oldRow, oldCol int) (int, int) {
	row, col := oldRow, oldCol
	switch direction {
	case keyUp:
		row--
		if row < 0 {
			row = len(maze) - 1
		}
	case keyDown:
		row++
		if len(maze) <= row {
			row = 0
		}
	case keyRight:
		col++
		if len(maze[0]) <= col {
			col = 0
		}
	case keyLeft:
		col--
		if col < 0 {
			col = len(maze[0]) - 1
		}
	}

	if maze[row][col] == '#' {
		row, col = oldRow, oldCol
	}

	return row, col
}
