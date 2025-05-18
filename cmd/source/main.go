package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dispatcherURL := flag.String("dispatcherURL", "", "url where webhook-dispatcher is running")
	flag.Parse()

	start(ctx, *dispatcherURL)
}

func start(ctx context.Context, dispatcherURL string) {
	type body struct {
		Time string `json:"time"`
	}

	c := http.Client{}
	for i := range time.Tick(time.Second) {
		b := body{Time: i.String()}

		out, err := json.Marshal(&b)
		if err != nil {
			log.Fatal(err)
		}

		url := fmt.Sprintf("%s/events/odd", dispatcherURL)
		if i.Second()%2 == 0 {
			url = fmt.Sprintf("%s/events/even", dispatcherURL)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(out))
		if err != nil {
			log.Fatal(err)
		}

		resp, err := c.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("source: resp = %s", resp.Status)
	}
}
