package main

import (
	"github.com/gorilla/mux"
        "io"
	"log"
	"log_monitor/monitor/core_utils"
	"log_monitor/monitor/core"
	"log_monitor/monitor/file_reader"
	"net/http"
	"strconv"
)

func main() {
	server := CreateLogServer(":10000",
		serveNLines("/home/jsong/projects/log_monitor/reader_file/"),
		serveFilterLines("/home/jsong/projects/log_monitor/reader_file/"),
		serveLinesThenFilter("/home/jsong/projects/log_monitor/reader_file/"))
	log.Fatal(server.ListenAndServe())
}

func CreateLogServer(addr string, handleNLines http.HandlerFunc, filterLines http.HandlerFunc, linesThenFilter http.HandlerFunc) http.Server {
	router := mux.NewRouter()
	router.HandleFunc("/{file}", linesThenFilter).Queries("lines", "{lines}").Queries("filter", "{filter}")
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
		lines, err := file_reader.ReadLastNLinesFromFile(baseDir+file, uint64(n))
		if err != nil {
			http.NotFound(w, r)
			return
		}
                io.Copy(w, lines)
	}
}

func serveFilterLines(baseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		file := vars["file"]
		filter := vars["filter"]

		lines, err := file_reader.ReadLastLinesContainsStringFromFile(baseDir+file, filter)
		if err != nil {
			http.NotFound(w, r)
			return
		}
                io.Copy(w, lines)
	}
}

func serveLinesThenFilter(baseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		file := vars["file"]
		nlines := vars["lines"]
		filter := vars["filter"]
		n, _ := strconv.Atoi(nlines)

                res, err := file_reader.ReadLastNLinesFromFile(baseDir+file, uint64(n))
                res, err = core_utils.LogFuncBind(res, err,
                    func(buffer io.ReadSeeker) (io.ReadSeeker, error) {
                        return core.ReadLastLinesContainsStringHelper(buffer, filter)
                    })
		if err != nil {
			http.NotFound(w, r)
			return
		}
                io.Copy(w, res)
	}
}
