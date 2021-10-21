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

const loopIteration = 1

const max = 20
const chanSize = 2

const feedMilis = 200
const tickerMilis = 500
const block = 5

type SingleRequest struct {
	request      int
	responseChan chan SingleResponse
}

type SingleResponse struct {
	response int
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func group(chanIn chan SingleRequest, chanOut chan []int) {

	var chanAux = make(chan SingleRequest, chanSize)

	go func() {
		ticker := time.NewTicker(tickerMilis * time.Millisecond)
		for {
			select {
			case <-ticker.C:

				var arr []SingleRequest
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

				if len(arr) > 0 {
					//fmt.Println("Out -> tiempo o por longitud (", length >= block, ")")
					var grouped []int
					for _, simple := range arr {
						simple.responseChan <- SingleResponse{response: simple.request}
						grouped = append(grouped, simple.request)
					}
					chanOut <- grouped
				}
			}
		}
	}()

	go func() {
		for item := range chanIn {
			//fmt.Println("Reading from chanIn -> ", item.request)

			chanAux <- item

			var arr []SingleRequest
			length := len(chanAux)
			if length >= block {
				for i := 0; i < length; i++ {
					item := <-chanAux
					arr = append(arr, item)
				}
				//fmt.Println("Out -> por longitud")
				var grouped []int
				for simple := range arr {
					item.responseChan <- SingleResponse{response: simple}
					grouped = append(grouped, simple)
				}
				chanOut <- grouped
			}

		}
		close(chanAux)
	}()

}

func loop(iteration int, wait *sync.WaitGroup, done chan bool) {
	var chanIn = make(chan SingleRequest, max)

	var chanOut = make(chan []int, chanSize)

	go func() {
		// Seed should be set once, better spot is func init()
		rand.Seed(time.Now().UTC().UnixNano())
		for i := 0; i < max; i++ {

			go func() {
				// TODO Imprimir tu respuesta... tras hacer batch y recibir la respuesta y separarlas...
				// metiendo el envio y recpcion en gorutine extra...

				responseChannel := make(chan SingleResponse, 1)
				wait.Add(1)
				req := SingleRequest{
					request:      i,
					responseChan: responseChannel,
				}
				fmt.Println("In <- ", req.request)
				chanIn <- req

				//fmt.Println("Waiting response in controller")
				respFromChannel := <-responseChannel
				close(responseChannel)

				fmt.Println(respFromChannel, " -> End Controller")
			}()

			time.Sleep(time.Duration(randInt(1, feedMilis)) * time.Millisecond)
		}
		close(chanIn)

		wait.Wait()
		close(chanOut)
	}()

	group(chanIn, chanOut)

	go func() {
		for item := range chanOut {
			for _, _ = range item {
				wait.Done()
			}
			fmt.Println(item, " -> Out (", iteration, ")")
		}
		done <- true
	}()
}

func main() {

	var done = make(chan bool, loopIteration)
	var wait = &sync.WaitGroup{}

	fmt.Println("Waiting")

	for i := 0; i < loopIteration; i++ {
		loop(i, wait, done)
		<-done
	}

	fmt.Println("Finish")
}
