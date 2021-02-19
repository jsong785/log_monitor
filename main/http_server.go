package main

import (
	"github.com/gorilla/mux"
        "io"
	"log"
	"log_monitor/monitor/reader"
	"log_monitor/monitor/reader_file"
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
		lines, err := reader_file.ReadLastNLinesFromFile(baseDir+file, uint64(n))
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

		lines, err := reader_file.ReadLastLinesContainsStringFromFile(baseDir+file, filter)
		if err != nil {
			http.NotFound(w, r)
			return
		}
                io.Copy(w, lines)
	}
}

type LogFuncMonad func(io.ReadSeeker) (io.ReadSeeker, error)
func LogFuncBind(buffer io.ReadSeeker, err error, f ...LogFuncMonad) (io.ReadSeeker, error) {
    if err != nil {
        return nil, err
    }

    if _, err := buffer.Seek(0, io.SeekEnd); err != nil {
        return nil, err
    }
    if len(f) == 1 {
        return f[0](buffer)
    }
    res, err := f[0](buffer)
    return LogFuncBind(res, err, f[1:]...)
}

func serveLinesThenFilter(baseDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		file := vars["file"]
		nlines := vars["lines"]
		filter := vars["filter"]
		n, _ := strconv.Atoi(nlines)

                res, err := reader_file.ReadLastNLinesFromFile(baseDir+file, uint64(n))
                res, err = LogFuncBind(res, err,
                    func(buffer io.ReadSeeker) (io.ReadSeeker, error) {
                        return reader.ReadLastLinesContainsStringHelper(buffer, filter)
                    })
		if err != nil {
			http.NotFound(w, r)
			return
		}
                io.Copy(w, res)
	}
}
