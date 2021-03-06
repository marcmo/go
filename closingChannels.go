package main

import (
	"fmt"
	"sync"
	"time"
)

const WORKERS int = 5

func crunchNumber(queue <-chan int, out chan<- int, quit <-chan bool, id int, wg *sync.WaitGroup, workerWg *sync.WaitGroup) {
	defer workerWg.Done()
	for {
		select {
		case task, ok := <-queue:
			if !ok {
				return
			}
			fmt.Printf("[%d]:crunching %d\n", id, task)
			if task > 1 {
				res := task / 2
				rest := task - res
				out <- res
				out <- rest
				wg.Add(2)
			} else {
				fmt.Printf("[%d]:done processing\n", id)
			}
			time.Sleep(50 * time.Millisecond)
			wg.Done()
		case <-quit:
			fmt.Println("QUITTING")
			return
		}
	}
}

func main() {
	queue := make(chan int, 100)
	unfiltered := make(chan int, 100)
	quit := make(chan bool)
	var wg sync.WaitGroup
	var workerWg sync.WaitGroup

	wg.Add(1)
	queue <- 10

	workerWg.Add(WORKERS)
	for i := 0; i < WORKERS; i++ {
		go crunchNumber(queue, unfiltered, quit, i, &wg, &workerWg)
	}
	// setup filtering
	go filterDuplicates(unfiltered, queue, &wg)

	fmt.Printf("waiting for queue...\n")
	wg.Wait()
	fmt.Printf("queue done!!\n")
	for i := 0; i < WORKERS; i++ {
		quit <- true
	}
	fmt.Printf("waiting for crunchers...\n")
	workerWg.Wait()
	fmt.Printf("crunchers done!!\n")
}

func filterDuplicates(in chan int, out chan int, wg *sync.WaitGroup) {
	var seen = make(map[int]bool)
	for v := range in {
		if !seen[v] {
			seen[v] = true
			out <- v
		} else {
			fmt.Printf("discarding %d!!\n", v)
			wg.Done() //already counted for, but we don't use it
		}
	}
}
