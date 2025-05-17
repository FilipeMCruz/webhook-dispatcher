package main

import (
	"context"
	"log"
	"net/http"
	"webhook-dispatcher/api"
	"webhook-dispatcher/broadcaster"
	"webhook-dispatcher/db"
)

func main() {
	ctx := context.Background()
	defer ctx.Done()

	database, err := db.NewDB()
	if err != nil {
		log.Fatal(err)
	}

	bServer := broadcaster.NewBroadcastServer[*http.Request](ctx)

	all, err := database.FetchAll()
	if err != nil {
		log.Fatal(err)
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
	mux.Handle("/events", api.BuildIngressEndpointHandler(bServer))
	mux.Handle("POST /subscribers", api.BuildRegisterSubscriberEndpointHandler(database, bServer))
	mux.Handle("DELETE /subscribers/{id}", api.BuildRemoveSubscriberEndpointHandler(database, bServer))

	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
