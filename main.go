package main

import (
    "net/http"

    "github.com/Danzabar/gorge/engine"
    "github.com/Danzabar/gorge/utils"
    "github.com/gorilla/mux"
)

var (
    GM *engine.GameManager
)

func main() {
    GM = engine.NewGame()

    router := mux.NewRouter()
    router.HandleFunc("/server", WsHandler)

    http.Handle("/", router)
    http.ListenAndServe(":8080", nil)
}

func TestHandler(w http.ResponseWriter, r *http.Request) {

}

// A websocket handler
func WsHandler(w http.ResponseWriter, r *http.Request) {
    if err := utils.ConnectToServer(GM, "test", w, r); err != nil {
        panic(err)
    }
}
