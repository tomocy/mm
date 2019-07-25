package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"math/rand"
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
	ghosts []ghost
	dots   int
}

func (g *game) load(name string) error {
	if err := g.loadMaze(name); err != nil {
		return err
	}
	if err := g.loadPlayer(); err != nil {
		return err
	}
	if err := g.loadGhosts(); err != nil {
		return err
	}

	g.countDots()

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
	player, err := g.maze.findPlayer()
	if err != nil {
		return err
	}

	g.player = player

	return nil
}

func (g *game) loadGhosts() error {
	ghosts, err := g.maze.findGhosts()
	if err != nil {
		return err
	}

	g.ghosts = ghosts

	return nil
}

func (g *game) countDots() {
	dots := g.maze.find(levelDot)

	g.dots = len(dots)
}

func (g *game) run() error {
	for {
		g.flush()

		key, err := readKey()
		if err != nil {
			return err
		}
		if key == keyEsc || !g.canContinue() {
			break
		}

		g.moveGhosts()
		g.movePlayer(key)
	}

	return nil
}

func (g *game) flush() {
	cleanScreen()

	g.flushMaze()
	g.flushMeta()
	g.flushGhosts()
	g.flushPlayer()
}

func cleanScreen() {
	fmt.Print("\x1b[2J")
	moveCursor(point{0, 0})
}

func (g *game) flushMaze() {
	fmt.Println(g.maze)
}

func (g *game) flushMeta() {
	moveCursor(point{x: 0, y: len(g.maze) + 1})
	fmt.Println("Score; ", g.player.score)
	fmt.Println("Lives: ", g.player.lives)
}

func (g *game) flushDebugMessage() {
	fmt.Println(g.player.position)
	fmt.Println(g.ghosts)
}

func (g *game) flushGhosts() {
	for _, ghost := range g.ghosts {
		moveCursor(ghost.position)
		fmt.Print(levelGhost)
		moveCursor(ghost.position)
	}
}

func (g *game) flushPlayer() {
	moveCursor(g.player.position)
	fmt.Print(levelPlayer)
	moveCursor(g.player.position)
}

func moveCursor(point point) {
	fmt.Printf("\x1b[%d;%df", point.y+1, point.x+1)
}

func (g *game) canContinue() bool {
	return 0 < g.player.lives && 0 < g.dots
}

func (g *game) moveGhosts() {
	for i := range g.ghosts {
		g.ghosts[i].moveRandomly(g.maze)
	}
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
			if string(char) == levelBlock || string(char) == levelDot {
				target = char
			}

			buf.WriteRune(target)
		}

		buf.WriteByte('\n')
	}

	return buf.String()
}

func (m maze) findPlayer() (player, error) {
	poss := m.find(levelPlayer)
	if len(poss) <= 0 {
		return player{}, errors.New("no player")
	}

	return player{
		position: poss[0],
	}, nil
}

func (m maze) findGhosts() ([]ghost, error) {
	poss := m.find(levelGhost)
	if len(poss) <= 0 {
		return nil, errors.New("no ghost")
	}

	ghosts := make([]ghost, len(poss))
	for i, pos := range poss {
		ghosts[i] = ghost{
			position: pos,
		}
	}

	return ghosts, nil
}

func (m maze) find(level string) []point {
	var founds []point
	for y, line := range m {
		for x, char := range line {
			if string(char) != level {
				continue
			}

			found := point{
				x: x,
				y: y,
			}

			founds = append(founds, found)
		}
	}

	return founds
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
	lives    int
	score    int
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
	levelGhost  = "G"
	levelBlock  = "#"
	levelDot    = "."
)

type ghost struct {
	position point
}

func (g *ghost) moveRandomly(maze maze) {
	dir := randomDirection()
	g.position = move(maze, dir, g.position)
}

func randomDirection() key {
	x := rand.Intn(4)
	dirs := [...]key{keyUp, keyDown, keyRight, keyLeft}

	return dirs[x]
}
