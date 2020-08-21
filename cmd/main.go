package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"net/http"

	api "api"
)

func main(){
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		_ = <-sigs
		done <- true
	}()

	httpHandler := api.StartHTTP()

	srv := &http.Server{
		Handler: httpHandler,
		Addr: ":8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout: 15	* time.Second,
	}

	go func() {
		log.Fatal(srv.ListenAndServe())
	}()

	log.Println("api started")
	<-done
	srv.Shutdown(context.Background())

	log.Println("api stopped")
}