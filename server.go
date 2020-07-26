package main

import (
	"fmt"
	"net/http"
	Controllers "./controllerClasses"
)

func handleRequests() {
	fmt.Println("Server started on: http://localhost:8080")
	http.HandleFunc("/hash/", Controllers.GetHashedValue)
	http.HandleFunc("/hash", Controllers.SetHashedValue)
	http.HandleFunc("/stats", Controllers.ReadStats)
	http.HandleFunc("/shutdown", Controllers.PrepShutdown)
	http.ListenAndServe(":8080", nil)
}

func main() {
	handleRequests()
}