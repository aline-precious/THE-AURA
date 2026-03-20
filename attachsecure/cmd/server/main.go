package main

import (
	"log"
	"net/http"
	"os"

	handler "attachsecure/api"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("AttachSecure running at http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, http.HandlerFunc(handler.Handler)); err != nil {
		log.Fatal(err)
	}
}
