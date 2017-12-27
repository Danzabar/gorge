package main

import (
    "net/http"
    "text/template"

    "github.com/Danzabar/gorge/engine"
    "github.com/Danzabar/gorge/utils"
    "github.com/gorilla/mux"
    "github.com/teris-io/shortid"
)

var (
    // Instance of the Game Manager
    // this is needed to connect
    GM *engine.GameManager

    // A test script to see the events streaming from the
    // server
    T = `<script>
            const ws = new WebSocket("ws://localhost:8080/server");

            ws.onopen = function() {
                console.log('connected');
            }

            ws.onmessage = function(e) {
                console.log(e);
            }

            ws.onclose = function() {
                console.log('disconnected');
            }
        </script>`
)

func main() {
    GM = engine.NewGame()

    utils.LoadDefaultComponents(GM)

    go GM.Run()

    router := mux.NewRouter()
    router.HandleFunc("/", TestHandler)
    router.HandleFunc("/server", WsHandler)

    http.Handle("/", router)
    http.ListenAndServe(":8080", nil)
}

func TestHandler(w http.ResponseWriter, r *http.Request) {
    t, _ := template.New("out").Parse(T)
    t.Execute(w, nil)
}

// A websocket handler
func WsHandler(w http.ResponseWriter, r *http.Request) {
    id, _ := shortid.Generate()
    if err := utils.ConnectToServer(GM, id, w, r); err != nil {
        panic(err)
    }
}
