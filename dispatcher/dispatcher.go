package dispatcher

import (
	"bytes"
	"github.com/google/uuid"
	"log"
	"net/http"
	"net/url"
	"time"
)

type RequestInfo struct {
	Method string
	URL    string
	Body   []byte
	Header http.Header
}

type Dispatcher struct {
	ID     uuid.UUID
	Token  string
	URL    string
	Rules  *MatchingRules
	client *http.Client
}

func NewDispatcher(id uuid.UUID, token string, downstreamURL string, matchingRules *MatchingRules) (*Dispatcher, error) {

	_, err := url.Parse(downstreamURL)
	if err != nil {
		return nil, err
	}

	s := Dispatcher{
		ID:    id,
		Token: token,
		URL:   downstreamURL,
		Rules: matchingRules,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	return &s, nil
}

func (s *Dispatcher) Listen(ch <-chan RequestInfo) {
	for req := range ch {
		if s.Rules.match(req) {
			err := s.dispatch(req)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func (s *Dispatcher) dispatch(req RequestInfo) error {
	log.Printf("sending request to %s", s.URL)

	r, err := http.NewRequest(req.Method, s.URL, bytes.NewReader(req.Body))
	if err != nil {
		return err
	}
	for k, v := range req.Header {
		r.Header[k] = v
	}
	r.Header.Set("Token", s.Token)

	_, err = s.client.Do(r)

	return err
}
