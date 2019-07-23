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
	maze, err := loadMaze("./maze.txt")
	if err != nil {
		log.Fatalf("failed to load maze: %s\n", err)
	}

	for {
		printScreen(maze)

		key, err := readInput()
		if err != nil {
			log.Printf("failed to read input: %s\n", err)
			break
		}
		if key == keyEsc {
			break
		}
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

	return string(buf[:cnt]), nil
}

const (
	keyEsc = "ecs"
)

type player struct {
	position position
}

type position struct {
	row, col int
}
