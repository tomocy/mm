package main

import (
	"bufio"
	"os"
)

func main() {}

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
	bs := make([]byte, 0)
	for _, line := range m {
		bs = append(bs, line...)
	}

	return string(bs)
}
