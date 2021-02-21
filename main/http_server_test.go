package main

import (
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestNonExistentFile(t *testing.T) {
	res, err := http.NewRequest("GET", "/non_existent_file", nil)
	assert.Nil(t, err)

	response := executeRequest(res, getRouter("../files/"))
	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestExistentFile_NoQuery(t *testing.T) {
	res, err := http.NewRequest("GET", "/syslog_ex", nil)
	assert.Nil(t, err)

	response := executeRequest(res, getRouter("../files/"))
	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestExistentFile_ValidQueryNLines(t *testing.T) {
	res, err := http.NewRequest("GET", "/syslog_ex?lines=2", nil)
	assert.Nil(t, err)

	response := executeRequest(res, getRouter("../files/"))
	assert.Equal(t, "jkl\nghi\n", response.Body.String())
}

func TestExistentFile_ValidQueryFilterLines(t *testing.T) {
	res, err := http.NewRequest("GET", "/syslog_ex?filter=l", nil)
	assert.Nil(t, err)

	response := executeRequest(res, getRouter("../files/"))
	assert.Equal(t, "jkl\n_world\n_hello\n", response.Body.String())
}

func TestExistentFile_ValidQueryNLinesThenFilterLines(t *testing.T) {
	res, err := http.NewRequest("GET", "/syslog_ex?lines=3&filter=l", nil)
	assert.Nil(t, err)

	response := executeRequest(res, getRouter("../files/"))
	assert.Equal(t, "jkl\n", response.Body.String())
}

// documenting order does not matter here (not a great feature)
func TestExistentFile_ValidQueryNLinesThenFilterLines_DifferntOrdering(t *testing.T) {
	res, err := http.NewRequest("GET", "/syslog_ex?filter=l&lines=3", nil)
	assert.Nil(t, err)

	response := executeRequest(res, getRouter("../files/"))
	assert.Equal(t, "jkl\n", response.Body.String())
}

func TestExistentFile_ValidQuery_InvalidMethod(t *testing.T) {
	testMethod := func(method string) {
		res, err := http.NewRequest(method, "/syslog_ex?lines=3", nil)
		assert.Nil(t, err)

		response := executeRequest(res, getRouter("../files/"))
		assert.Equal(t, http.StatusMethodNotAllowed, response.Code)
	}
	// post
	testMethod("POST")
	testMethod("PUT")
	testMethod("HEAD")
	testMethod("DELETE")
	testMethod("PATCH")
	testMethod("OPTIONS")
}

func BenchmarkLargeFileRead_SingleRequest(b *testing.B) {
	res, err := http.NewRequest("GET", "/syslog_large?lines=1000", nil)
	assert.Nil(b, err)
	router := getRouter("../files/")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executeRequest(res, router)
	}
}

func BenchmarkLargeFileRead_ManyRequests(b *testing.B) {
	res, err := http.NewRequest("GET", "/syslog_large?lines=50", nil)
	assert.Nil(b, err)
	router := getRouter("../files/")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for j := 0; j < 1000; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				executeRequest(res, router)
				assert.Nil(b, err)
			}()
		}
		wg.Wait()
	}
}

func executeRequest(request *http.Request, router *mux.Router) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}
