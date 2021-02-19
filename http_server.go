package log_monitor

import (
    "github.com/gorilla/mux"
    "log"
    "net/http"
)

func CreateLogServer(addr string, handleNLines http.HandlerFunc, filterLines http.HandlerFunc) http.Server {
    router := mux.NewRouter()
    router.HandleFunc("/{file}/nlines={arg}", handleNLines)
    router.HandleFunc("/{file}/filter={arg}", filterLines)
    return http.Server {
        Addr: addr,
        Handler: router,
    }
}

func main() {
    server := CreateLogServer(":10000",
        func(w http.ResponseWriter, r *http.Request) {
        },
        func(w http.ResponseWriter, r *http.Request) {
        })
    log.Fatal(server.ListenAndServe())
}

