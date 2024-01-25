package main

import (
	"log"
	"net/http"

	metadata "github.com/ugurcancaykara/odd-service/metadata/internal/controller"
	httphandler "github.com/ugurcancaykara/odd-service/metadata/internal/handler/http"
	"github.com/ugurcancaykara/odd-service/metadata/internal/repository/memory"
)

func main() {
	log.Println("Starting the movie metadata service")
	repo := memory.New()
	ctrl := metadata.New(repo)
	h := httphandler.New(ctrl)
	http.Handle("/metadata", http.HandlerFunc(h.GetMetadata))
	if err := http.ListenAndServe(":8081", nil); err != nil {
		panic(err)
	}
}
