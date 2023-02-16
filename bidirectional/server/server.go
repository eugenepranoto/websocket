package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// peraturan dari gorilla websocket. write message hanya tidak bisa
// dipanggil bersamaan di go routine yg berbeda
// maka dari itu disini pake 2 go routine
// 1 untuk handle read dan 1 untuk handle write
// sehingga pasti tidak dobel

var upgrader = websocket.Upgrader{}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}

	defer c.Close()

	done := make(chan struct{})

	go receiver(c, done)
	go sender(c, done)

	<-done

	log.Println("websocker handler is done")
}

func receiver(ws *websocket.Conn, done chan struct{}) {
	defer func() {
		ws.Close()
	}()

	ws.SetPongHandler(func(string) error {
		log.Println("received pong")
		return nil
	})

	for {
		mt, message, err := ws.ReadMessage()
		if err != nil {
			log.Println("error: ", err)
			break
		}
		log.Printf("receive:%s", message)
		err = ws.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
		}
	}

	close(done)
}

func sender(ws *websocket.Conn, done chan struct{}) {
	defer func() {
		ws.Close()
	}()

	messageTicker := time.NewTicker(1 * time.Second)
	defer messageTicker.Stop()

	pingTicker := time.NewTicker(5 * time.Second)

	counter := 0

breakLoop:
	for {
		select {
		case <-messageTicker.C:
			data := "hello world"
			if err := ws.WriteMessage(websocket.TextMessage, []byte(data)); err != nil {
				log.Println("write", err)
				return
			}
			if counter > 20 {
				break breakLoop
			}
			counter++
		case <-pingTicker.C:
			if err := ws.WriteControl(websocket.PingMessage, []byte(`ping message`), time.Time{}); err != nil {
				log.Println("print err", err)
				return
			}
		case <-done:
			return
		}
	}
	err := ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		fmt.Println("error close", err)
	}
	close(done)
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/", echo)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
