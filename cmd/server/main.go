// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/webdeveloppro/golang-websocket-client/pkg/server"
)

var addr = flag.String("addr", ":8000", "http service address")

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}

func main() {
	flag.Parse()

	hub := server.NewHub()
	go hub.Run()
	http.HandleFunc("/frontend", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("got new connection")
		server.ServeWs(hub, w, r)
	})

	fmt.Println("server started ... ")
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		panic(err)
	}

}
