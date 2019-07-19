package main

import (
	"bufio"
	"bytes"
	"fmt"
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
	buf := new(bytes.Buffer)
	for _, line := range m {
		fmt.Fprintf(buf, "%s\n", line)
	}

	return buf.String()
}
