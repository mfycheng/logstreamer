package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	_ "net/http/pprof"

	"github.com/mfycheng/logstreamer"
	"golang.org/x/net/websocket"
)

var (
	stream        = logstreamer.NewBufferedLogStream(10)
	listenAddress = flag.String("addr", ":8080", "Listen address")
	webSocketHTML = `
	<html>
		<head>
			<script type="text/javascript">
				function websocket()
				{
					if ("WebSocket" in window) {
						// Let us open a web socket
						var ws = new WebSocket("ws://" + window.location.hostname + ":" + window.location.port + "/log_stream");

						ws.onmessage = function (evt) {
							div = document.getElementById("out");
							div.innerHTML += evt.data + "<br/>";
						};

						ws.onerror = function() {
							ws.close();
						}
					} else {
						// Redirect to non-websocket stream.
						window.location("/log?ws=false");
					}
				}

				window.onload = websocket();
			</script>
		</head>
		<body><div id="out" style="font-family: monospace;"/></body>
	</html>
	`
)

type writerFlusher struct {
	writer io.Writer
}

func (w *writerFlusher) Write(p []byte) (n int, err error) {
	n, err = w.writer.Write(p)
	if err != nil {
		return n, err
	}

	if f, ok := w.writer.(http.Flusher); ok {
		f.Flush()
	}

	return n, err
}

func handler(w http.ResponseWriter, req *http.Request) {
	useWebSockets := true

	// User explicitly doesn't want websockets
	if req.URL.Query().Get("ws") == "false" {
		useWebSockets = false
	}

	// cURL doesn't support websockets.
	if strings.Contains(req.UserAgent(), "curl") {
		useWebSockets = false
	}

	if useWebSockets {
		if _, err := fmt.Fprint(w, webSocketHTML); err != nil {
			log.Println(err)
		}
	} else {
		// This really should be done using proper long polling, especially
		// if the logs have very low output rates. However, the only clients
		// that don't support websockets is Opera Mini 8 (as of 5/3/2016),
		// and cURL. Since a user can simply use logstream, or not Opera Mini 8,
		// it's probably not worth the effort to implement long polling.
		err := logstreamer.StreamToWriter(stream, &writerFlusher{w})
		if err != nil {
			log.Println(err)
		}
	}
}

func webSocketStream(ws *websocket.Conn) {
	defer ws.Close()

	logstreamer.StreamToWriter(stream, ws)
}

func main() {
	flag.Parse()

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			stream.WriteLine(scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			log.Fatal("reading stdin:", err)
		}
	}()

	http.HandleFunc("/log", handler)
	http.Handle("/log_stream", websocket.Handler(webSocketStream))
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		log.Fatal(err)
	}
}
