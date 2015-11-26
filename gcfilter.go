package gcfilter

import (
	"bufio"
	"log"
	"os"
)

var current *os.File

// Filter ..
func Filter() error {

	current = os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		return err
	}

	os.Stdout = w

	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()

			if line[0:2] == "gc" {

				log.Println("Found a GC - %v", line)

				continue
			}
			current.WriteString(line)
		}
	}()

	return nil
}

// Reset returns Stderr to the original location
func Reset() {
	os.Stderr = current
}
