package main

import (
	"bufio"
	"bytes"
	"errors"
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

	cleanScreen()
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
	if err := g.loadMaze(name); err != nil {
		return err
	}

	if err := g.loadPlayer(); err != nil {
		return err
	}

	return nil
}

func (g *game) loadMaze(name string) error {
	maze, err := loadMaze(name)
	if err != nil {
		return err
	}

	g.maze = maze

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

func (g *game) loadPlayer() error {
	player, err := loadPlayer(g.maze)
	if err != nil {
		return err
	}

	g.player = player

	return nil
}

func loadPlayer(maze maze) (player, error) {
	for y, line := range maze {
		for x, level := range line {
			if string(level) == levelPlayer {
				return player{
					position: point{
						y: y,
						x: x,
					},
				}, nil
			}
		}
	}

	return player{}, errors.New("no player in given maze")
}

func (g *game) start() error {
	for {
		g.flush()

		key, err := readKey()
		if err != nil {
			return err
		}
		if key == keyEsc {
			break
		}

		g.movePlayer(key)
	}

	return nil
}

func (g *game) flush() {
	cleanScreen()

	g.flushMaze()
	fmt.Println(g.player.position)
	g.flushPlayer()
}

func cleanScreen() {
	fmt.Print("\x1b[2J")
	moveCursor(0, 0)
}

func (g *game) flushMaze() {
	fmt.Println(g.maze)
}

func (g *game) flushPlayer() {
	moveCursor(g.player.position.y+1, g.player.position.x+1)
	fmt.Print(levelPlayer)
	moveCursor(g.player.position.y+1, g.player.position.x+1)
}

func moveCursor(row, col int) {
	fmt.Printf("\x1b[%d;%df", row, col)
}

func (g *game) movePlayer(key key) {
	g.player.move(g.maze, key)
}

type maze []string

func (m maze) String() string {
	buf := new(bytes.Buffer)
	for _, line := range m {
		for _, char := range line {
			target := ' '
			if string(char) == levelBlock {
				target = char
			}

			buf.WriteRune(target)
		}

		buf.WriteByte('\n')
	}

	return buf.String()
}

const (
	levelPlayer = "P"
	levelBlock  = "#"
)

func readKey() (key, error) {
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

	return "", nil
}

const (
	keyEsc   = "ecs"
	keyUp    = "up"
	keyDown  = "down"
	keyRight = "right"
	keyLeft  = "left"
)

type key string

type player struct {
	position point
}

type point struct {
	x, y int
}

func (p *player) move(maze maze, key key) {
	p.position.y, p.position.x = move(maze, key, p.position.y, p.position.x)
}

func move(maze maze, key key, oldY, oldX int) (int, int) {
	y, x := oldY, oldX
	switch key {
	case keyUp:
		y--
		if y < 0 {
			y = len(maze) - 1
		}
	case keyDown:
		y++
		if len(maze) <= y {
			y = 0
		}
	case keyRight:
		x++
		if len(maze[0]) <= x {
			x = 0
		}
	case keyLeft:
		x--
		if x < 0 {
			x = len(maze[0]) - 1
		}
	}

	if maze[y][x] == '#' {
		y, x = oldY, oldX
	}

	return y, x
}
