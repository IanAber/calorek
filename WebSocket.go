package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	EnableCompression: true,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var dataSignal *sync.Cond

/**
Start the Web Socket server. This sends out data to all subscribers on a regular schedule so subscribers don't need to poll for updates.
*/
func startDataWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	for {
		dataSignal.L.Lock()   // Get the signal and lock it.
		dataSignal.Wait()     // Wait for it to be signalled again. It is unlocked while we wait then locked again before returning
		dataSignal.L.Unlock() // Unlock it

		w, err := conn.NextWriter(websocket.TextMessage)
		if err != nil {
			log.Println("startDataWebSocket - ", err)
			return
		}

		sJSON, err := Params.getJSON()
		_, err = fmt.Fprint(w, string(sJSON))
		if err != nil {
			log.Println("failed to write the values message to the websocket - ", err)
			return
		}
		if err := w.Close(); err != nil {
			//			log.Println("Failed to close the values websocket writer - ", err)
			return
		}
	}
}
