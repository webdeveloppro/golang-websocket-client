package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/webdeveloppro/golang-websocket-client/pkg/client"
)

var addr = flag.String("addr", ":8000", "http service address")

func main() {
	flag.Parse()

	client, err := client.NewWebSocketClient(*addr, "frontend")
	if err != nil {
		panic(err)
	}
	fmt.Println("Connecting")

	go func() {
		// write down data every 100 ms
		ticker := time.NewTicker(time.Millisecond * 1500)
		i := 0
		for range ticker.C {
			err := client.Write(i)
			if err != nil {
				fmt.Printf("error: %v, writing error\n", err)
			}
			i++
		}
	}()

	// Close connection correctly on exit
	sigs := make(chan os.Signal, 1)

	// `signal.Notify` registers the given channel to
	// receive notifications of the specified signals.
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// The program will wait here until it gets the
	<-sigs
	client.Stop()
	fmt.Println("Goodbye")
}
