package api

import (
	"encoding/json"
	"github.com/google/uuid"
	"io"
	"log"
	"net/http"
	"webhook-dispatcher/broadcaster"
	"webhook-dispatcher/db"
	"webhook-dispatcher/dispatcher"
)

func BuildIngressEndpointHandler(server broadcaster.BroadcastServer[dispatcher.RequestInfo]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Print(err)
		}
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(r.Body)

		server.Publish(dispatcher.RequestInfo{
			Method: r.Method,
			URL:    r.URL.String(),
			Body:   body,
			Header: r.Header,
		})

		log.Printf("dispatcher: req = %s", r.URL.Path)

		w.WriteHeader(http.StatusOK)
	})
}

func BuildRegisterSubscriberEndpointHandler(bufferSize int, database db.DB, server broadcaster.BroadcastServer[dispatcher.RequestInfo]) http.Handler {
	type requestBody struct {
		Url           string `json:"url,omitempty"`
		Token         string `json:"token,omitempty"`
		MatchingRules struct {
			URLPath string            `json:"url_path,omitempty"`
			Body    []byte            `json:"body,omitempty"`
			Headers map[string]string `json:"headers,omitempty"`
		} `json:"matching_rules,omitempty"`
	}

	type responseBody struct {
		ID string `json:"id,omitempty"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New()

		i := &requestBody{}

		err := json.NewDecoder(r.Body).Decode(&i)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(r.Body)

		rules, err := dispatcher.NewMatchingRules(i.MatchingRules.URLPath, i.MatchingRules.Headers, i.MatchingRules.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		d, err := dispatcher.NewDispatcher(id, i.Token, i.Url, rules)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		err = database.Save(d)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		err = json.NewEncoder(w).Encode(responseBody{ID: id.String()})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		go d.Listen(server.Subscribe(id, bufferSize))
	})
}

func BuildRemoveSubscriberEndpointHandler(database db.DB, server broadcaster.BroadcastServer[dispatcher.RequestInfo]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := uuid.Parse(r.PathValue("id"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		if !database.Exists(id, r.Header.Get("Token")) {
			http.Error(w, err.Error(), http.StatusNotFound)

			return
		}

		server.Unsubscribe(id)

		err = database.Remove(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}
	})
}
