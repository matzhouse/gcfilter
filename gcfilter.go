package gcfilter

import (
	"bufio"
	"fmt"
	"os"

	syscall "golang.org/x/sys/unix"
)

var current *os.File

type Wall struct {
	SweepTerm, Scan, Sync, Mark, MarkTerm float64
}

// Filter ..
func Filter() error {

	fmt.Println("filtering GC statements")

	current = os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		return err
	}

	rr, ww, err := os.Pipe()
	if err != nil {
		return err
	}

	// dup the Stderr File descriptor
	syscall.Dup2(int(w.Fd()), int(os.Stderr.Fd()))

	// move os.Stderr
	os.Stderr = ww

	// Listen on os.Stderr for messages and pipe them to the new files below
	go func(fwd *os.File, rr *os.File) {
		fmt.Println("running the goroutine")
		scanner := bufio.NewScanner(rr)
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) != 0 {
				w.Write(line)
			}
		}
	}(w, rr)

	go func() {
		fmt.Println("running the goroutine")
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()

			var time, perc, wall, cpu string
			var number int

			if line[0:2] == "gc" {
				w := &Wall{}
				// gc 1 @0.248s 0%: 0.14+0.43+0.33+0.54+0.43 ms clock, 0.29+0.43+0+0.13/0.44/1.4+0.86 ms cpu, 4->4->1 MB, 4 MB goal, 4 P
				fmt.Sscanf(line, "gc %d %s %s %s ms clock, %s ms cpu,", &number, &time, &perc, &wall, &cpu)
				fmt.Sscanf(wall, "%f+%f+%f+%f+%f", &w.SweepTerm, &w.Scan, &w.Sync, &w.Mark, &w.MarkTerm)
				fmt.Println(w)
			} else {
				fmt.Println(line)
			}

		}
	}()

	return nil
}

// Reset returns Stderr to the original location
func Reset() {
	os.Stderr = current
}
