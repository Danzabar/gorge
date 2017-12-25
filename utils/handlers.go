package utils

import (
    "net/http"

    "github.com/Danzabar/gorge/engine"
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

func ConnectToServer(GM *engine.GameManager, id string, w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, w.Header())

    if err != nil {
        http.Error(w, "An error has occured", http.StatusBadRequest)
        return
    }

    GM.Connect(conn, id)
}
