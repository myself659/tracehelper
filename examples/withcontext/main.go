package main

import (
	"context"
	"fmt"
	"time"
	"math/rand"
	"github.com/myself659/tracehelper"
)

func main() {
	// Pass a context with a timeout to tell a blocking function that it
	// should abandon its work after the timeout elapses.
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	tracehelper.WithContext(ctx, "ctx")
	defer cancel()
	go func() {
		select {
		case <-time.After(10 * time.Second):
			fmt.Println("g1 waitting for 10s")
		case <-ctx.Done():
			fmt.Println("g1", ctx.Err()) // prints "g1 context deadline exceeded"
		}
	}()
	go func() {
		select {
		case <-time.After(4 * time.Second):
			fmt.Println("g2 waitting for 4s")
		case <-ctx.Done():
			fmt.Println("g2 ", ctx.Err()) // prints "g2 context deadline exceeded"
		}
	}()
	d := time.Duration(rand.Intn(12))
	<-time.After(d * time.Second)
}
