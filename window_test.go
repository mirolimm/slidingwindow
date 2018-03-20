package main

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

func TestAddVal(t *testing.T) {
	windowSize := 3
	window := NewWindow(windowSize)
	cases := []struct {
		val            int
		expectedValues []int
	}{
		{10, []int{10}},
		{11, []int{10, 11}},
		{101, []int{10, 11, 101}},
		{201, []int{201, 11, 101}},
	}

	for i, tc := range cases {
		window.AddVal(tc.val)
		if !reflect.DeepEqual(tc.expectedValues, window.buffer) {
			t.Errorf("case %d, expected %+v, got %+v", i, tc.expectedValues, window.buffer)
		}
	}
}

func TestGetMedian(t *testing.T) {
	windowSize := 10
	window := NewWindow(windowSize)
	cases := []struct {
		val          int
		sortedValues []int
	}{
		{10, []int{10}},
		{11, []int{10, 11}},
		{101, []int{10, 11, 101}},
		{201, []int{10, 11, 101, 201}},
		{50, []int{10, 11, 50, 101, 201}},
		{60, []int{10, 11, 50, 60, 101, 201}},
		{210, []int{10, 11, 50, 60, 101, 201, 210}},
		{110, []int{10, 11, 50, 60, 101, 110, 201, 210}},
		{20, []int{10, 11, 20, 50, 60, 101, 110, 201, 210}},
		{20, []int{10, 11, 20, 20, 50, 60, 101, 110, 201, 210}},
		{1000, []int{11, 20, 20, 50, 60, 101, 110, 201, 210, 1000}},
		{70, []int{20, 20, 50, 60, 70, 101, 110, 201, 210, 1000}},
	}

	for i, tc := range cases {
		window.AddVal(tc.val)
		expectedVal := findMedianInSortedSlice(tc.sortedValues)
		median := window.GetMedian()
		if median != expectedVal {
			t.Errorf("case %d, expected %+v, got %+v", i, expectedVal, median)
		}
	}
}

// test with -race flag
func TestConcurrentAccess(t *testing.T) {
	start := make(chan struct{})
	finish := make(chan struct{})
	windowSize := 1000
	window := NewWindow(windowSize)
	values := genRandomValuesInRange(1000, 10000)
	go func() {
		<-start
		for _, v := range values {
			window.AddVal(v)
		}
		// signal finish
		close(finish)
	}()

	fnChecker := func(id int) {
		<-start
		val := -2
	LOOP:
		for {
			val = window.GetMedian()
			select {
			case <-finish:
				break LOOP
			default:
			}

		}
		fmt.Printf("g%d last val %d\n", id, val)

	}

	go fnChecker(1)
	go fnChecker(2)

	close(start) // signal start
	<-finish     // wait finish
}

func BenchmarkAddVal1000Size(b *testing.B) {
	windowSize := 1000
	window := NewWindow(windowSize)
	values := genRandomValuesInRange(1000, 2000)
	j := 0
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		window.AddVal(values[j])
		if j < len(values)-1 {
			j++
		} else {
			j = 0
		}

	}
	fmt.Printf("window first elm %d\n", window.buffer[0]) // output for debug
}

func BenchmarkAddToBuf1000Size(b *testing.B) {
	oldVal := 0
	val := 2018
	windowSize := 1000
	window := NewWindow(windowSize)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		oldVal = window.addToBuf(val)
	}
	fmt.Printf("old val %d\n", oldVal) // output for debug
}

func BenchmarkUpdateSorted10000Size(b *testing.B) {
	windowSize := 10000
	window := NewWindow(windowSize)
	values := genRandomValuesInRange(1000, 10000)
	for _, v := range values {
		window.AddVal(v)
	}
	j := 0
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		window.updateSorted(values[j], values[j])
		if j < len(values)-1 {
			j++
		} else {
			j = 0
		}

	}
	fmt.Printf("window first elm %d\n", window.buffer[0]) // output for debug
}

func genRandomValuesInRange(max, length int) []int {
	// make stable random
	r := rand.New(rand.NewSource(1000))
	sl := make([]int, length)
	for i := 0; i < length; i++ {
		sl[i] = r.Intn(max)
	}

	return sl
}
