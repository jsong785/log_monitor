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

func main() {
	addr := flag.String("addr", "localhost:8080", "address:port to run server")
	dir := flag.String("dir", "/var/syslog", "default serving directory")
	timeout := flag.Uint64("timeout", 2000, "timeout in milliseconds to serve a request")
	flag.Parse()

	server := CreateLogServer(*addr,
		*timeout,
		serveNLines(*dir),
		serveFilterLines(*dir),
		serveLinesThenFilter(*dir))
	log.Fatal(server.ListenAndServe())
}

func CreateLogServer(addr string, timeout uint64, handleNLines http.HandlerFunc, filterLines http.HandlerFunc, linesThenFilter http.HandlerFunc) http.Server {
	router := mux.NewRouter()
	router.HandleFunc("/{file}", linesThenFilter).Queries("lines", "{lines}").Queries("filter", "{filter}")
	router.HandleFunc("/{file}", handleNLines).Queries("lines", "{lines}")
	router.HandleFunc("/{file}", filterLines).Queries("filter", "{filter}")
	return http.Server{
		Addr:         addr,
		Handler:      router,
		WriteTimeout: time.Duration(timeout) * time.Millisecond,
	}
}

func serveNLines(baseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		file := vars["file"]
		nlines := vars["lines"]

		n, _ := strconv.Atoi(nlines)

		res, err := file_reader.ReadReverseNLines(filepath.Join(baseDir, file), uint64(n))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		io.Copy(w, res)
	}
}

func serveFilterLines(baseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		file := vars["file"]
		filter := vars["filter"]

		res, err := file_reader.ReadReversePassesFilter(filepath.Join(baseDir, file), filter)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		io.Copy(w, res)
	}
}

func serveLinesThenFilter(baseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		file := vars["file"]
		nlines := vars["lines"]
		filter := vars["filter"]
		n, _ := strconv.Atoi(nlines)

		res, err := file_reader.ReadReverseNLines(filepath.Join(baseDir, file), uint64(n))
		res, err = core_utils.LogFuncBind(res, err,
			func(b io.ReadSeeker) (io.ReadSeeker, error) {
				return core.ReadReversePassesFilter(b, filter)
			})
		if err != nil {
			http.NotFound(w, r)
			return
		}
		io.Copy(w, res)
	}
}
