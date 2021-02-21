package main

import (
	"flag"
	"github.com/gorilla/mux"
	"io"
	"log"
	"log_monitor/monitor/core"
	"log_monitor/monitor/core_utils"
	"log_monitor/monitor/file_reader"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)

func CreateLogServer(dir string, address string, readTimeout uint, writeTimeout uint) http.Server {
	return http.Server{
		Addr:         address,
		Handler:      getRouter(dir),
		ReadTimeout:  time.Duration(readTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(writeTimeout) * time.Millisecond,
	}
}

func getRouter(dir string) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/{file}", serveLinesThenFilter(dir)).Queries("lines", "{lines}").Queries("filter", "{filter}").Methods("GET")
	router.HandleFunc("/{file}", serveNLines(dir)).Queries("lines", "{lines}").Methods("GET")
	router.HandleFunc("/{file}", serveFilterLines(dir)).Queries("filter", "{filter}").Methods("GET")
	return router
}

func serveNLines(baseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path, n, err := nLinesParse(baseDir, r)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		res, err := file_reader.ReadReverseNLines(path, n)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		io.Copy(w, res)
	}
}

func serveFilterLines(baseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path, filter := filterLinesParse(baseDir, r)
		res, err := file_reader.ReadReversePassesFilter(path, filter)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		io.Copy(w, res)
	}
}

func serveLinesThenFilter(baseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path, n, err := nLinesParse(baseDir, r)
		_, filter := filterLinesParse(baseDir, r)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		res, err := file_reader.ReadReverseNLines(path, n)
		res, err = core_utils.LogFuncBind(res, err, func(buf io.ReadSeeker) (io.ReadSeeker, error) {
			return core.ReadReversePassesFilter(buf, filter)
		})
		if err != nil {
			http.NotFound(w, r)
			return
		}
		io.Copy(w, res)
	}
}

func nLinesParse(baseDir string, r *http.Request) (string, uint64, error) {
	vars := mux.Vars(r)
	nLines, err := strconv.ParseUint(vars["lines"], 10, 64)
	return filepath.Join(baseDir, vars["file"]), nLines, err
}

func filterLinesParse(baseDir string, r *http.Request) (string, string) {
	vars := mux.Vars(r)
	return filepath.Join(baseDir, vars["file"]), vars["filter"]
}

func main() {
	addr := flag.String("addr", "localhost:8080", "address:port to run server")
	dir := flag.String("dir", "/var/log", "default serving directory")
	timeout := flag.Uint("timeout", 2000, "timeout in milliseconds to serve a request")
	flag.Parse()

	server := CreateLogServer(*dir, *addr, 100, *timeout)
	log.Fatal(server.ListenAndServe())
}
