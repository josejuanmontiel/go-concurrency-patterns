// In the previous example we saw how to manage simple
// counter state using [atomic operations](atomic-counters).
// For more complex state we can use a [_mutex_](http://en.wikipedia.org/wiki/Mutual_exclusion)
// to safely access data across multiple goroutines.
package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func main() {

	const max = 50
	const chanSize = 50

	const feedMilis = 5
	const tickerMilis = 500
	const block = 25

	var done = make(chan bool)

	var chanIn = make(chan int, max)
	var chanAux = make(chan int, chanSize)
	var chanOut = make(chan []int, chanSize)

	var wait = &sync.WaitGroup{}

	fmt.Println("Waiting")

	go func() {
		// Seed should be set once, better spot is func init()
		rand.Seed(time.Now().UTC().UnixNano())
		for i := 0; i < max; i++ {
			wait.Add(1)
			fmt.Println("In <-", i)
			chanIn <- i
			time.Sleep(time.Duration(randInt(1, feedMilis)) * time.Millisecond)
		}

		close(chanIn)

		wait.Wait()
		close(chanOut)
	}()

	go func() {
		ticker := time.NewTicker(tickerMilis * time.Millisecond)
		for {
			select {
			case <-ticker.C:

				var arr []int
				var howmany int

				length := len(chanAux)
				if length >= block {
					howmany = block
				} else {
					howmany = length
				}

				for i := 0; i < howmany; i++ {
					item := <-chanAux
					arr = append(arr, item)
				}
				chanOut <- arr
			}

		}
	}()

	go func() {
		for item := range chanOut {
			for _, _ = range item {
				wait.Done()
			}
			fmt.Println(item, " -> Out")
		}
		done <- true
	}()

	go func() {
		for item := range chanIn {
			chanAux <- item

			var arr []int
			length := len(chanAux)
			if length >= block {
				for i := 0; i < length; i++ {
					item := <-chanAux
					arr = append(arr, item)
				}
				chanOut <- arr
			}

		}
		close(chanAux)
	}()

	<-done
	fmt.Println("Finish")
}
