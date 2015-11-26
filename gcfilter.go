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

	os.Stderr = w

	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()

			log.Printf("LINE - %v \n", line)
			log.Printf("LINE START - %v \n", line[0:3])

			current.WriteString(line + "\n")
		}
	}()

	return nil
}

// Reset returns Stderr to the original location
func Reset() {
	os.Stderr = current
}
