package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var (
	// flags for commandline api
	size = flag.Int("size", 100, "window size")
	// input file where int values separated by \r\n
	fname = flag.String("file", "", "input filename")
)

func main() {
	flag.Parse()
	r, err := os.Open(*fname)
	if err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}
	window := NewWindow(*size)
	reader := bufio.NewReader(r)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Printf("%+v\n", err)
				os.Exit(1)
			}
			// check if there is nothing exit loop
			if line == "" {
				break
			}
		}
		line = strings.TrimRight(line, "\r\n")
		val, err := strconv.Atoi(line)
		if err != nil {
			fmt.Printf("%+v\n", err)
			os.Exit(1)
		}
		window.AddVal(val)
		fmt.Printf("%d\r\n", window.GetMedian())
	}
}

// Window struct to operate sliding window
// safe for concurrent use
type Window struct {
	m          sync.Mutex
	size       int
	buffer     []int
	oldestElm  int
	sortedVals []int
}

// NewWindow contructor
// size of the window
func NewWindow(size int) *Window {
	return &Window{
		size:       size,
		buffer:     make([]int, 0, size),
		oldestElm:  0,
		sortedVals: make([]int, 0, size),
	}
}

// AddVal adds value to sliding window
func (w *Window) AddVal(val int) {
	w.m.Lock()
	oldVal := w.addToBuf(val)
	w.updateSorted(oldVal, val)
	w.m.Unlock()
}

// addToBuf add value to fifo queue with size
// returns oldest element if queue is full
func (w *Window) addToBuf(val int) (oldVal int) {
	if w.size > len(w.buffer) {
		w.buffer = append(w.buffer, val)
		return
	}

	oldVal, w.buffer[w.oldestElm] = w.buffer[w.oldestElm], val
	w.oldestElm++

	if w.oldestElm == len(w.buffer) {
		w.oldestElm = 0
	}

	return
}

// updatedSorted updates sorted values in place
// removes old value if required
// insert new value in sorted order
func (w *Window) updateSorted(oldVal, newVal int) {
	// remove old value
	id := sort.Search(len(w.sortedVals), func(i int) bool { return w.sortedVals[i] >= oldVal })
	if id < len(w.sortedVals) && w.sortedVals[id] == oldVal {
		w.sortedVals = append(w.sortedVals[:id], w.sortedVals[id+1:]...)

	}

	// add new value
	inserted := false
	idn := sort.Search(len(w.sortedVals), func(i int) bool { return w.sortedVals[i] > newVal })
	if idn < len(w.sortedVals) && w.sortedVals[idn] > newVal {
		w.sortedVals = append(w.sortedVals, 0)
		copy(w.sortedVals[idn+1:], w.sortedVals[idn:])
		w.sortedVals[idn] = newVal
		inserted = true
	}

	if len(w.sortedVals) == 0 || !inserted {
		w.sortedVals = append(w.sortedVals, newVal)
	}

}

// GetMedian returns median value from sliding window
// returns -1 if there is only one element
func (w *Window) GetMedian() int {
	w.m.Lock()
	defer w.m.Unlock()
	return findMedianInSortedSlice(w.sortedVals)
}

// returns median
func findMedianInSortedSlice(sl []int) int {
	if len(sl) < 2 {
		return -1
	}

	if len(sl)%2 == 0 {
		n := len(sl)/2 - 1
		return (sl[n] + sl[n+1]) / 2

	}

	return sl[(len(sl)+1)/2-1]
}
