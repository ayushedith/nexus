package mock

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Server struct {
	endpoints map[string]*Endpoint
	mu        sync.RWMutex
}

type Endpoint struct {
	Path     string
	Method   string
	Response Response
	Matcher  *Matcher
	Delay    time.Duration
}

type Response struct {
	StatusCode int
	Headers    map[string]string
	Body       interface{}
}

type Matcher struct {
	HeaderMatchers map[string]*regexp.Regexp
	BodyMatcher    *regexp.Regexp
}

func NewServer() *Server {
	return &Server{
		endpoints: make(map[string]*Endpoint),
	}
}

func (s *Server) AddEndpoint(e *Endpoint) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s:%s", e.Method, e.Path)
	s.endpoints[key] = e
}

func (s *Server) RemoveEndpoint(method, path string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s:%s", method, path)
	delete(s.endpoints, key)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	endpoint := s.findEndpoint(r)
	if endpoint == nil {
		http.NotFound(w, r)
		return
	}

	if endpoint.Delay > 0 {
		time.Sleep(endpoint.Delay)
	}

	for k, v := range endpoint.Response.Headers {
		w.Header().Set(k, v)
	}

	w.WriteHeader(endpoint.Response.StatusCode)

	switch body := endpoint.Response.Body.(type) {
	case string:
		w.Write([]byte(body))
	case []byte:
		w.Write(body)
	default:
		data, err := json.Marshal(body)
		if err != nil {
			slog.Error("marshal response", "error", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}

func (s *Server) findEndpoint(r *http.Request) *Endpoint {
	key := fmt.Sprintf("%s:%s", r.Method, r.URL.Path)
	if endpoint, ok := s.endpoints[key]; ok {
		if endpoint.Matcher == nil || s.matchesRequest(endpoint.Matcher, r) {
			return endpoint
		}
	}

	for _, endpoint := range s.endpoints {
		if endpoint.Method != r.Method {
			continue
		}

		if s.pathMatches(endpoint.Path, r.URL.Path) {
			if endpoint.Matcher == nil || s.matchesRequest(endpoint.Matcher, r) {
				return endpoint
			}
		}
	}

	return nil
}

func (s *Server) pathMatches(pattern, path string) bool {
	if strings.Contains(pattern, "*") {
		re := regexp.MustCompile("^" + strings.ReplaceAll(pattern, "*", ".*") + "$")
		return re.MatchString(path)
	}
	return pattern == path
}

func (s *Server) matchesRequest(matcher *Matcher, r *http.Request) bool {
	for header, re := range matcher.HeaderMatchers {
		value := r.Header.Get(header)
		if !re.MatchString(value) {
			return false
		}
	}

	return true
}

func (s *Server) Start(addr string) error {
	slog.Info("mock server starting", "addr", addr)
	return http.ListenAndServe(addr, s)
}
