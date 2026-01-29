package collection

import (
	"time"
)

type Collection struct {
	Name        string                 `json:"name" yaml:"name"`
	BaseURL     string                 `json:"baseUrl" yaml:"baseUrl"`
	Environment map[string]Environment `json:"environment" yaml:"environment"`
	Requests    []Request              `json:"requests" yaml:"requests"`
	PreRequest  string                 `json:"preRequest,omitempty" yaml:"preRequest,omitempty"`
	Tests       []string               `json:"tests,omitempty" yaml:"tests,omitempty"`
}

type Environment struct {
	BaseURL   string            `json:"baseUrl" yaml:"baseUrl"`
	Variables map[string]string `json:"variables,omitempty" yaml:"variables,omitempty"`
}

type Request struct {
	Name       string            `json:"name" yaml:"name"`
	Method     string            `json:"method" yaml:"method"`
	URL        string            `json:"url" yaml:"url"`
	Headers    map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	QueryParams map[string]string `json:"queryParams,omitempty" yaml:"queryParams,omitempty"`
	Body       interface{}       `json:"body,omitempty" yaml:"body,omitempty"`
	Auth       *Auth             `json:"auth,omitempty" yaml:"auth,omitempty"`
	PreRequest string            `json:"preRequest,omitempty" yaml:"preRequest,omitempty"`
	Tests      []string          `json:"tests,omitempty" yaml:"tests,omitempty"`
	Assertions []string          `json:"assertions,omitempty" yaml:"assertions,omitempty"`
}

type Auth struct {
	Type   string            `json:"type" yaml:"type"`
	Config map[string]string `json:"config" yaml:"config"`
}

type Response struct {
	StatusCode int
	Status     string
	Headers    map[string][]string
	Body       []byte
	Time       time.Duration
	Size       int64
}

type ExecutionResult struct {
	Request   Request
	Response  Response
	Error     error
	StartTime time.Time
	EndTime   time.Time
	Passed    bool
	Failures  []string
}
