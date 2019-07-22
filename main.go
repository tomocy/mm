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

func main() {}

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
