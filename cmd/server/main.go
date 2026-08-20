package main

import (
	"go-metrics/internal/handler"
	"net/http"
)

func main() {
	mux := handler.NewRouter()

	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}
