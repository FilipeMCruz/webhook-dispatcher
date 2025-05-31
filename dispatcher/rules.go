package dispatcher

import (
	"bytes"
	"regexp"
	"strings"
)

type MatchingRules struct {
	Path    *regexp.Regexp
	Headers map[string]string
	Body    []byte
}

func NewMatchingRules(path string, headers map[string]string, body []byte) (*MatchingRules, error) {
	rg, err := regexp.Compile(path)
	if err != nil {
		return nil, err
	}

	return &MatchingRules{
		Path:    rg,
		Headers: headers,
		Body:    body,
	}, nil
}

func (r *MatchingRules) match(req RequestInfo) bool {
	_, after, _ := strings.Cut(req.URL, "/events/")
	if !r.Path.MatchString(after) {
		return false
	}

	for k, v := range r.Headers {
		if req.Header.Get(k) != v {
			return false
		}
	}

	return bytes.Contains(req.Body, r.Body)
}
