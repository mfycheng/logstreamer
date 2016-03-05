package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"

	"golang.org/x/net/websocket"
)

var (
	host = flag.String("host", "localhost", "Host of the streamer")
	port = flag.Int("port", 8080, "Port of the streamer")
)

func main() {
	flag.Parse()
	origin := fmt.Sprintf("http://%v/", *host)
	url := fmt.Sprintf("ws://%v:%v/log_stream", *host, *port)

	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	log.Println("Scanning...")
	scanner := bufio.NewScanner(ws)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}
