package main

import (
	"fmt"
	"sync"
	"time"
)

//this program represents concurrency «pipeline and fan in/fan out» patterns
type item struct {
	category string
	price    int
}

func main() {
	test := producer(
		item{"belt", 50},
		item{"pant", 100},
		item{"shoe", 150},
		item{"coat", 200},
	)
	c1 := disscount(test)
	c2 := disscount(test)
	out := fanIn(c1, c2)
	for processed := range out { //reading data from a channel
		fmt.Println("category:", processed.category, "price:", processed.price)
	}

}

//this function takes all elements and writes/converting them to the channel
func producer(items ...item) <-chan item {
	out := make(chan item, len(items))
	for _, i := range items {
		out <- i
	}
	close(out)
	return out
}

//this function proccesing our data and then giving back a channel to read out to see what the resualt is
func disscount(items <-chan item) <-chan item {
	out := make(chan item)
	go func() {
		defer close(out)
		for i := range items {
			time.Sleep(time.Second / 2) //simulates server load
			if i.category == "shoe" {
				i.price /= 2
			}
			if i.category == "belt" {
				i.price /= 2
			}
			out <- i
		}
	}()
	return out
}

//this function splitting two copies of the discount goroutines in order to add some parallelism and unloads server work
func fanIn(channels ...<-chan item) <-chan item {
	var wg sync.WaitGroup
	out := make(chan item) //creating output channel that will have all data from the discount goroutines

	output := func(c <-chan item) { //stores anonymous function in a variable, to use it on for each of the channels being read discount goroutines

		defer wg.Done() //this defer guarantees us that all data has been read before we close our single combine channel
		for i := range c {
			out <- i
		}
	}
	wg.Add(len(channels))
	for _, c := range channels {
		go output(c) //calls anon-function in goroutine to read outcomes data of discount func in the background and putting them into single outbound channel
	}
	go func() { //waits until all inbound data has been read
		wg.Wait()
		close(out)
	}()
	return out
}
