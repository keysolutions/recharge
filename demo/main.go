// demo is a sample application used to experiment with the recharger example.
package main

import (
	"log"
	"net/http"
)

func main() {
	err := http.ListenAndServe(":3001", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	}))
	log.Fatal(err)
}
