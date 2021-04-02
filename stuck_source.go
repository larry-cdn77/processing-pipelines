// cancellation context can be seen as a convenience without which a data sink
// routine often ends up having to store away what coomes out of a channel
// input as a result of a negative close poll
//
// after introducing cancellation context it becomes possible for source to
// block on a buffered channel if sink is cancelled prematurely
//
// the situation is easily demonstrated with an infinitely fast source and
// a slow blocking sink, meant to represent a real setting where source channel
// depth in insufficient to spread burstiness of incoming data evenly
//
// one solution is close & drain to complete the shutdown cleanly
//
// try removing the drain loop: source is likely to end up blocked and
// waitgroup will detect deadlock at runtime
//
package main

import (
	"context"
	"log"
	"sync"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	wg := new(sync.WaitGroup)

	c := make(chan int, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		n := 0
		for {
			select {
				case <-ctx.Done():
					close(c)
					log.Print("source exit")
					return
				default:
					c <- n
					log.Print("<-", n)
					n += 1
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
				case <-ctx.Done():
					for range c {
						// drain
					}
					log.Print("sink exit")
					return
				case n := <-c:
					log.Print(n, "<-")
					time.Sleep(1*time.Second) // work
			}
		}
	}()

	time.Sleep(1*time.Second)

	log.Print("initiate shutdown")
	cancel()

	wg.Wait()
}
