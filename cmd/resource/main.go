package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	runTestServer("")
}

func runTestServer(ip string) {
	http.HandleFunc("/hi", func(writer http.ResponseWriter, request *http.Request) {
		log.Println(request.RemoteAddr)
		writer.Write([]byte(fmt.Sprintf("hi %s Mesage: %s \n ", request.RemoteAddr, request.URL.String())))
		return
	})
	err := http.ListenAndServe(fmt.Sprintf("%s:8080", ip), nil)
	if err != nil {
		log.Println(err)
	}
}
