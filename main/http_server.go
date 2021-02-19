package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"log_monitor/monitor/reader_file"
	"net/http"
	"strconv"
)

func main() {
	server := CreateLogServer(":10000",
		serveNLines("/home/jsong/projects/log_monitor/reader_file/"),
		serveFilterLines("/home/jsong/projects/log_monitor/reader_file/"))
	log.Fatal(server.ListenAndServe())
}

func CreateLogServer(addr string, handleNLines http.HandlerFunc, filterLines http.HandlerFunc) http.Server {
	router := mux.NewRouter()
	router.HandleFunc("/{file}", handleNLines).Queries("lines", "{lines}")
	router.HandleFunc("/{file}", filterLines).Queries("filter", "{filter}")
	return http.Server{
		Addr:    addr,
		Handler: router,
	}
}

func serveNLines(baseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		file := vars["file"]
		nlines := vars["lines"]

		n, _ := strconv.Atoi(nlines)
		lines, err := reader_file.ReadLastNLinesFromFile(baseDir+file, uint64(n))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		for _, l := range lines {
			fmt.Fprintf(w, "%s\n", l)
		}
	}
}

func serveFilterLines(baseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		file := vars["file"]
		filter := vars["filter"]

		lines, err := reader_file.ReadLastLinesContainsStringFromFile(baseDir+file, filter)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		for _, l := range lines {
			fmt.Fprintf(w, "%s\n", l)
		}
	}
}
