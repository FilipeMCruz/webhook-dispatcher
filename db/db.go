package db

import (
	"errors"
	"github.com/google/uuid"
	"sync"
	"webhook-dispatcher/dispatcher"
)

type Dispatcher struct {
	ID            uuid.UUID
	Token         string
	URL           string
	MatchingRules MatchingRules
}

type MatchingRules struct {
	Path    string
	Headers map[string]string
	Body    []byte
}

type DB interface {
	Save(sub *dispatcher.Dispatcher) error
	Fetch(id uuid.UUID, token string) (*dispatcher.Dispatcher, error)
	Exists(id uuid.UUID, token string) bool
	Remove(id uuid.UUID) error
	FetchAll() ([]*dispatcher.Dispatcher, error)
}

type database struct {
	entries map[uuid.UUID]Dispatcher
	m       sync.Mutex
}

func (d *database) FetchAll() ([]*dispatcher.Dispatcher, error) {
	d.m.Lock()
	defer d.m.Unlock()
	result := make([]*dispatcher.Dispatcher, 0)

	for _, e := range d.entries {
		r, err := dispatcher.NewMatchingRules(e.MatchingRules.Path, e.MatchingRules.Headers, e.MatchingRules.Body)
		if err != nil {
			return nil, err
		}

		d, err := dispatcher.NewDispatcher(e.ID, e.Token, e.URL, r)
		if err != nil {
			return nil, err
		}

		result = append(result, d)
	}

	return result, nil
}

func (d *database) Save(sub *dispatcher.Dispatcher) error {
	d.m.Lock()
	defer d.m.Unlock()
	d.entries[sub.ID] = Dispatcher{
		ID:    sub.ID,
		Token: sub.Token,
		URL:   sub.URL,
		MatchingRules: MatchingRules{
			Path:    sub.Rules.Path.String(),
			Headers: sub.Rules.Headers,
			Body:    sub.Rules.Body,
		},
	}

	return nil
}

func (d *database) Fetch(id uuid.UUID, token string) (*dispatcher.Dispatcher, error) {
	d.m.Lock()
	defer d.m.Unlock()
	e, ok := d.entries[id]
	if !ok {
		return nil, errors.New("not found")
	}

	if e.Token != token {
		return nil, errors.New("invalid token")
	}

	r, err := dispatcher.NewMatchingRules(e.MatchingRules.Path, e.MatchingRules.Headers, e.MatchingRules.Body)
	if err != nil {
		return nil, err
	}

	return dispatcher.NewDispatcher(e.ID, e.Token, e.URL, r)
}

func (d *database) Exists(id uuid.UUID, token string) bool {
	d.m.Lock()
	defer d.m.Unlock()
	e, ok := d.entries[id]
	if !ok {
		return false
	}

	return e.Token != token
}

func (d *database) Remove(id uuid.UUID) error {
	d.m.Lock()
	defer d.m.Unlock()
	delete(d.entries, id)
	return nil
}

func NewDB() (DB, error) {
	return &database{entries: make(map[uuid.UUID]Dispatcher)}, nil
}
