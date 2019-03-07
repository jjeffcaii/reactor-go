package rs

import (
	"context"
	"log"
	"math/rand"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestQueue_Poll(t *testing.T) {
	done := make(chan struct{})
	qu := NewQueue(1)
	go func(ctx context.Context) {
		defer close(done)
		for {
			v, ok := qu.Poll(ctx, func(ctx context.Context, q *Queue) {
				time.Sleep(1 * time.Second)
				n := (rand.Uint32() % 4) + 1
				log.Println("requestN:", n)
				q.RequestN(int32(n))
			})
			if !ok {
				break
			}
			log.Println("next:", v)
		}
	}(context.Background())

	//go func() {
	//	tk := time.NewTicker(10 * time.Millisecond)
	//	n := 0
	//	for {
	//		if n > 10 {
	//			break
	//		}
	//		select {
	//		case v := <-tk.C:
	//			qu.Add(v)
	//			n++
	//		}
	//	}
	//	tk.Stop()
	//}()
	<-done
}

func TestCh(t *testing.T) {
	done := make(chan struct{})

	<-done
}
