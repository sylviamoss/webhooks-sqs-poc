package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func handlePost(w http.ResponseWriter, req *http.Request) {
	buf := new(strings.Builder)
	_, _ = io.Copy(buf, req.Body)
	fmt.Println("=== Message Received ===")
	fmt.Println(buf.String())

	fmt.Println("->> Returning bad request")
	w.WriteHeader(http.StatusBadRequest)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", handlePost).Methods("POST")

	srv := &http.Server{
		Addr:    ":3000",
		Handler: r,
	}
	srv.ListenAndServe()
}
