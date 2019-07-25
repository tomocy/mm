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

	if err := game.run(); err != nil {
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

func (g *game) run() error {
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
	g.flushDebugMessage()
	g.flushPlayer()
}

func cleanScreen() {
	fmt.Print("\x1b[2J")
	moveCursor(point{0, 0})
}

func (g *game) flushMaze() {
	fmt.Println(g.maze)
}

func (g *game) flushDebugMessage() {
	fmt.Println(g.player.position)
}

func (g *game) flushPlayer() {
	moveCursor(g.player.position)
	fmt.Print(levelPlayer)
	moveCursor(g.player.position)
}

func moveCursor(point point) {
	fmt.Printf("\x1b[%d;%df", point.y+1, point.x+1)
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
	p.position = move(maze, key, p.position)
}

func move(maze maze, key key, oldPos point) point {
	pos := point{
		x: oldPos.x,
		y: oldPos.y,
	}
	switch key {
	case keyUp:
		pos.y--
		if pos.y < 0 {
			pos.y = len(maze) - 1
		}
	case keyDown:
		pos.y++
		if len(maze) <= pos.y {
			pos.y = 0
		}
	case keyRight:
		pos.x++
		if len(maze[0]) <= pos.x {
			pos.x = 0
		}
	case keyLeft:
		pos.x--
		if pos.x < 0 {
			pos.x = len(maze[0]) - 1
		}
	}

	if string(maze[pos.y][pos.x]) == levelBlock {
		pos = oldPos
	}

	return pos
}

const (
	levelPlayer = "P"
	levelBlock  = "#"
)

type ghost struct {
	position point
}
