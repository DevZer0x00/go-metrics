package main

import (
	"go-metrics/internal/routes"
	"net/http"
)

func main() {
	mux := routes.NewRouter()

	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}
