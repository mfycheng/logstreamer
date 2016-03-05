package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "net/http/pprof"

	"github.com/mfycheng/logstreamer"
	"golang.org/x/net/websocket"
)

var (
	stream = logstreamer.NewBufferedLogStream(10)
)

func webSocketStream(ws *websocket.Conn) {
	obs := stream.NewObserver()
	defer obs.Close()
	defer ws.Close()

	var lastNumber int64
	for entry := range obs.Chan() {
		if entry.Number > lastNumber+1 {
			fmt.Fprintf(ws, "Skipping %v lines...\n", entry.Number-(lastNumber+1))
		}

		lastNumber = entry.Number

		_, err := fmt.Fprintf(ws, "%v - %v\n", time.Now().Format(time.UnixDate), entry.Line)
		if err != nil {
			log.Println("Error writing:", err)
			return
		}
	}
}

func logPage(w http.ResponseWriter, req *http.Request) {
	f, err := os.Open("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer f.Close()
	io.Copy(w, f)
}

func main() {
	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			stream.WriteLine(scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			log.Fatal("reading stdin:", err)
		}
	}()

	http.HandleFunc("/log", logPage)
	http.Handle("/log_stream", websocket.Handler(webSocketStream))
	if err := http.ListenAndServe(":9998", nil); err != nil {
		log.Fatal(err)
	}
}
