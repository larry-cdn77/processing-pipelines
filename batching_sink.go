// a high-performance sink minimises lock-based penalty of select statement
// by performing batches of channel inputs without polling cancellation context
//
// such sink can become blocked if source is cancelled in the middle of sink
// batch
//
// if source also closes its channel afterwards, that terminates the batch in
// progress and shutdown can proceed cleanly
//
// note that a context-based cancellation path remains in the sink can it can
// be shut down in assorted code paths without having to store away what comes
// out of a channel input as a result of a negative close poll
//
// try removing the close statement: sink ends up blocked and waitgroup will
// detect deadlock at runtime
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
					log.Print("sink exit")
					return
				default:
					for i := 0; i < 5; i++ {
						n, ok := <-c
						if !ok {
							log.Print("close seen")
							break
						}
						log.Print(n, "<-")
						time.Sleep(1*time.Second) // work
					}
			}
		}
	}()

	time.Sleep(2*time.Second)

	log.Print("initiate shutdown")
	cancel()

	wg.Wait()
}
