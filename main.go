package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"
	"webhook-dispatcher/api"
	"webhook-dispatcher/broadcaster"
	"webhook-dispatcher/db"
	"webhook-dispatcher/dispatcher"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	port := flag.Int("port", 8080, "port to listen on")
	flag.Parse()

	err := start(ctx, stop, *port)
	if err != nil {
		log.Fatal(err)
	}
}

func start(ctx context.Context, stop func(), port int) error {
	database, err := db.NewDB()
	if err != nil {
		return err
	}

	bServer := broadcaster.NewBroadcastServer[dispatcher.RequestInfo](ctx)

	all, err := database.FetchAll()
	if err != nil {
		return err
	}

	for _, d := range all {
		ch := bServer.Subscribe(d.ID)
		go func() {
			for req := range ch {
				d.Send(req)
			}
		}()
	}

	mux := http.NewServeMux()
	mux.Handle("/events/", api.BuildIngressEndpointHandler(bServer))
	mux.Handle("POST /subscribers", api.BuildRegisterSubscriberEndpointHandler(database, bServer))
	mux.Handle("DELETE /subscribers/{id}", api.BuildRemoveSubscriberEndpointHandler(database, bServer))

	ongoingCtx, stopOngoingGracefully := context.WithCancel(context.Background())
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
		BaseContext: func(_ net.Listener) context.Context {
			return ongoingCtx
		},
	}

	go func() {
		log.Printf("dispatcher starting on port %d", port)
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
		log.Println("Stopped serving new connections.")
	}()

	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	defer stopOngoingGracefully()

	return httpServer.Shutdown(shutdownCtx)
}
