package dispatcher

import (
	"bytes"
	"github.com/google/uuid"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type RequestInfo struct {
	Method string
	URL    string
	Body   []byte
	Header http.Header
}

type Dispatcher struct {
	ID          uuid.UUID
	Token       string
	URL         string
	MatchingURL string
	rules       matchingRules
	client      *http.Client
}

type matchingRules struct {
	Path *regexp.Regexp
}

func NewDispatcher(ID uuid.UUID, Token string, URL string, MatchingURL string) (*Dispatcher, error) {
	rg, err := regexp.Compile(MatchingURL)
	if err != nil {
		return nil, err
	}

	_, err = url.Parse(URL)
	if err != nil {
		return nil, err
	}

	s := Dispatcher{
		ID:          ID,
		Token:       Token,
		URL:         URL,
		MatchingURL: MatchingURL,
		rules: matchingRules{
			Path: rg,
		},
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	return &s, nil
}

func (s *Dispatcher) Listen(ch <-chan RequestInfo) {
	for req := range ch {
		if s.match(req) {
			err := s.dispatch(req)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func (s *Dispatcher) match(req RequestInfo) bool {
	_, after, _ := strings.Cut(req.URL, "/events/")
	return s.rules.Path.MatchString(after)
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
