package db

import (
	"errors"
	"github.com/google/uuid"
	"webhook-dispatcher/dispatcher"
)

type Dispatcher struct {
	ID          uuid.UUID
	Token       string
	URL         string
	MatchingURL string
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
}

func (d database) FetchAll() ([]*dispatcher.Dispatcher, error) {
	result := make([]*dispatcher.Dispatcher, 0)

	for _, e := range d.entries {
		d, err := dispatcher.NewDispatcher(e.ID, e.Token, e.URL, e.MatchingURL)
		if err != nil {
			return nil, err
		}

		result = append(result, d)
	}

	return result, nil
}

func (d database) Save(sub *dispatcher.Dispatcher) error {
	d.entries[sub.ID] = Dispatcher{
		ID:          sub.ID,
		Token:       sub.Token,
		URL:         sub.URL,
		MatchingURL: sub.MatchingURL,
	}

	return nil
}

func (d database) Fetch(id uuid.UUID, token string) (*dispatcher.Dispatcher, error) {
	e, ok := d.entries[id]
	if !ok {
		return nil, errors.New("not found")
	}

	if e.Token != token {
		return nil, errors.New("invalid token")
	}

	return dispatcher.NewDispatcher(e.ID, e.Token, e.URL, e.MatchingURL)
}

func (d database) Exists(id uuid.UUID, token string) bool {
	e, ok := d.entries[id]
	if !ok {
		return false
	}

	return e.Token != token
}

func (d database) Remove(id uuid.UUID) error {
	delete(d.entries, id)
	return nil
}

func NewDB() (DB, error) {
	return database{entries: make(map[uuid.UUID]Dispatcher)}, nil
}
