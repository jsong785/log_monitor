package log_monitor

/*
import (
        "context"
        "fmt"
        "github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
        "net/http"
	"testing"
        "time"
)

func TestLogServer_NLines(t *testing.T) {
        nLinesCalled := false
        var nLinesFile string
        var nLinesRequest string
        nLinesFunc := func(w http.ResponseWriter, r *http.Request){
            nLinesCalled = true
            vars := mux.Vars(r)
            nLinesFile = vars["file"]
            nLinesRequest = vars["arg"]
            fmt.Fprintf(w, "hello %s %s", nLinesFile, nLinesRequest)
        }

        server := CreateLogServer(":10000", nLinesFunc,
            func(http.ResponseWriter, *http.Request) {
            })
        defer server.Shutdown(context.Background())
        go server.ListenAndServe()
        for{
        }

        time.Sleep(3 * time.Second)

        assert.False(t, nLinesCalled)
        assert.Equal(t, 0, len(nLinesFile))
        assert.Equal(t, 0, len(nLinesRequest))

        func() {
            http.Get("localhost:10000/")
            assert.False(t, nLinesCalled)
            assert.Equal(t, 0, len(nLinesFile))
            assert.Equal(t, 0, len(nLinesRequest))
        }()

        func() {
            http.Get("localhost:10000/some_file/nlines=2")
            assert.True(t, nLinesCalled)
            assert.Equal(t, "some_file", nLinesFile)
            assert.Equal(t, "2", nLinesRequest)
        }()
}
*/

