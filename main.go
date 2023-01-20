package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

func main() {

	http.HandleFunc("/ws", wsEndpoint)

	log.Println("http server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)
var clients = make(map[*websocket.Conn]bool)
var lock sync.RWMutex

func broadcast(ch chan string) {
	for {
		msg := <-ch
		lock.RLock()
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				log.Println("Error writing message to client:", err)
				client.Close()
				delete(clients, client)
			}
		}
		lock.RUnlock()
	}
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	defer ws.Close()

	lock.Lock()
	clients[ws] = true
	lock.Unlock()

	ch := make(chan string)
	go broadcast(ch)

	for {

		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Println(err)
			delete(clients, ws)
			break
		}
		ch <- string(msg)
	}

	lock.Lock()
	delete(clients, ws)
	lock.Unlock()

}
