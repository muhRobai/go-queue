package main

import (
	"api"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	worker, err := api.CreateWorker()
	if err != nil {
		log.Println(err)
		return
	}

	workercontext, workercancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		_ = <-sigs
		done <- true
	}()

	go worker.DispatchQueue(workercontext)

	log.Println("Batch worker started")
	<-done
	defer workercancel()
	log.Println("Batch worker stopped")
}
