package dispatcher

import (
	"github.com/google/uuid"
	"log"
	"net/http"
	"regexp"
)

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

	s := Dispatcher{
		ID:          ID,
		Token:       Token,
		URL:         URL,
		MatchingURL: MatchingURL,
		rules: matchingRules{
			Path: rg,
		},
		client: &http.Client{},
	}

	return &s, err
}

func (s *Dispatcher) match(req *http.Request) bool {
	return s.rules.Path.MatchString(req.URL.Path)
}

func (s *Dispatcher) dispatch(r *http.Request) error {
	r.Header.Set("Token", s.Token)
	_, err := s.client.Do(r)

	return err
}

func (s *Dispatcher) Send(req *http.Request) {
	if s.match(req) {
		err := s.dispatch(req)
		if err != nil {
			log.Println(err)
		}
	}
}
