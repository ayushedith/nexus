package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/net/http2"
)

type Client struct {
	client  *http.Client
	timeout time.Duration
}

type Config struct {
	Timeout            time.Duration
	InsecureSkipVerify bool
	MaxIdleConns       int
	MaxConnsPerHost    int
	EnableHTTP2        bool
}

func NewClient(cfg *Config) *Client {
	if cfg == nil {
		cfg = &Config{
			Timeout:         30 * time.Second,
			MaxIdleConns:    100,
			MaxConnsPerHost: 100,
			EnableHTTP2:     true,
		}
	}

	transport := &http.Transport{
		MaxIdleConns:        cfg.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.MaxConnsPerHost,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.InsecureSkipVerify,
			MinVersion:         tls.VersionTLS12,
		},
	}

	if cfg.EnableHTTP2 {
		http2.ConfigureTransport(transport)
	}

	return &Client{
		client: &http.Client{
			Transport: transport,
			Timeout:   cfg.Timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("stopped after 10 redirects")
				}
				return nil
			},
		},
		timeout: cfg.Timeout,
	}
}

type RequestOptions struct {
	Method      string
	URL         string
	Headers     map[string]string
	QueryParams map[string]string
	Body        []byte
}

func (c *Client) Do(ctx context.Context, opts *RequestOptions) (*Response, error) {
	start := time.Now()

	var bodyReader io.Reader
	if len(opts.Body) > 0 {
		bodyReader = bytes.NewReader(opts.Body)
	}

	req, err := http.NewRequestWithContext(ctx, opts.Method, opts.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	if opts.QueryParams != nil {
		q := req.URL.Query()
		for k, v := range opts.QueryParams {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	duration := time.Since(start)

	return &Response{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Headers:    resp.Header,
		Body:       body,
		Time:       duration,
		Size:       int64(len(body)),
		Proto:      resp.Proto,
	}, nil
}

type Response struct {
	StatusCode int
	Status     string
	Headers    map[string][]string
	Body       []byte
	Time       time.Duration
	Size       int64
	Proto      string
}
