package main

import (
	"log"
	"net/http"
)

// test command is `curl http://127.0.0.1:55555/hello`
func main() {
	http.HandleFunc("/hello", hello)
	log.Println("hello server is started ...")
	log.Fatal(http.ListenAndServe(":55555", nil))
}

func hello(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("Hello World"))
}
