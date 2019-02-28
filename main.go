package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dkrisa/jumpcloud/server"
)

func main() {
	fmt.Println("Hello, Jumpcloud.")
	defer fmt.Println("Goodbye, Jumpcloud.")

	mux := server.NewHashServeMux()
	s := &http.Server{
		Addr:           ":8000",
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	inFlightComplete := mux.SetUp(s)

	log.Println("INFO: Listening on port 8000")
	err := s.ListenAndServe()

	if err == http.ErrServerClosed {
		log.Println("INFO: Waiting for in-flight requests to complete.")
		<-inFlightComplete
		log.Println("INFO: All in-flight requests completed.  Exiting program.")
	} else {
		log.Panicln("PANIC:", err)
	}

}
