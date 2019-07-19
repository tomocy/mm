package main

import (
	"bufio"
	"os"
)

func main() {}

func loadMaze(name string) ([]string, error) {
	src, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	var maze []string
	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		line := scanner.Text()
		maze = append(maze, line)
	}

	return maze, nil
}
