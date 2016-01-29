package gcfilter

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"runtime"

	syscall "golang.org/x/sys/unix"
)

var current *os.File

// GCRun holds data about a GC run
type GCRun struct {
	N     int     // the GC number, incremented at each GC
	Time  float32 //time in seconds since program start
	Perc  int     // percentage of time spent in GC since program start
	Clock Timing
	CPU   CPU
	Heap  Heap
	Procs int
}

// Timing is the base struct for clock or cpu timing information
type Timing struct {
	SweepTerm, Scan, Sync, Mark, MarkTerm float64
}

type Mark struct {
	Assist, Background, Idle float64
}

type CPU struct {
	Timing
	Mark Mark
}

// Heap tracks the size of the heap during a GC run
type Heap struct {
	Start int
	End   int
	Live  int
	Goal  int
}

// Filter ..
func Filter() error {

	/*if GCOn() == false {
		return fmt.Errorf("GC Debug flag not set")
	}*/

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

	// Listen on os.Stderr for messages and pipe them to the new files below
	go func(fwd *os.File, rr *os.File) {
		fmt.Println("running the goroutine - os.Stderr listener")
		// Just copy everything from the new os.Stderr into the older one
		io.Copy(w, rr)
	}(w, rr)

	// move os.Stderr
	os.Stderr = ww

	go func() {
		fmt.Println("running the goroutine")
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()

			if line[0:2] == "gc" {
				var wall, cpu, hs string

				run := &GCRun{
					Clock: Timing{},
					CPU:   CPU{},
					Heap:  Heap{},
				}

				// gc 2 @1.289s 0%: 0.10+0.61+0.018+1.3+0.20 ms clock, 0.20+0.61+0+0.13/1.1/2.5+0.41 ms cpu, 4->4->1 MB, 4 MB goal, 4 P
				_, err := fmt.Sscanf(line, "gc %d @%fs %d%%: %s ms clock, %s ms cpu, %s MB, %d MB goal, %d P", &run.N, &run.Time, &run.Perc, &wall, &cpu, &hs, &run.Heap.Goal, &run.Procs)

				if err != nil {
					fmt.Print(err)
				}

				fmt.Sscanf(hs, "%d->%d->%d", &run.Heap.Start, &run.Heap.End, &run.Heap.Live)
				fmt.Sscanf(wall, "%f+%f+%f+%f+%f", &run.Clock.SweepTerm, &run.Clock.Scan, &run.Clock.Sync, &run.Clock.Mark, &run.Clock.MarkTerm)
				fmt.Sscanf(cpu, "%f+%f+%f+%f/%f/%f+%f", &run.CPU.SweepTerm, &run.CPU.Scan, &run.CPU.Sync, &run.CPU.Mark.Assist, &run.CPU.Mark.Background, &run.CPU.Mark.Idle, &run.CPU.MarkTerm)

				fmt.Println(run)

			} else {

				fmt.Println(line)

			}

		}
	}()

	return nil
}

func GCOn() bool {
	ms := &runtime.MemStats{}
	runtime.ReadMemStats(ms)

	return ms.DebugGC
}

// Reset returns Stderr to the original location
func Reset() {
	os.Stderr = current
}
