package main

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	defer log.Println("wsHandler: Exiting from wsHandler")

	// Upgrade HTTP connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorln("wsHandler: Upgrade error:", err)
		return
	}
	defer func() { _ = conn.Close() }()

	log.Println("Conectado!")

	done := make(chan struct{})
	go func() {
		if err := listenToDbus(conn, done); err != nil {
			log.Errorln("listenToDbus:", err)
			return
		}
	}()

	_, _, _ = conn.ReadMessage()
	//log.Infoln("CloseHandler> Type:", msgType, "Msg:", msg, "Err:", err)
	close(done)
}
