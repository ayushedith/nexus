package collection

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	nexushttp "github.com/nexusapi/nexus/pkg/http"
)

type Runner struct {
	client   *nexushttp.Client
	Resolver *VariableResolver
	env      string
}

func NewRunner(env string) *Runner {
	return &Runner{
		client:   nexushttp.NewClient(nil),
		Resolver: NewVariableResolver(env),
		env:      env,
	}
}

func (r *Runner) Run(coll *Collection) ([]ExecutionResult, error) {
	r.Resolver.LoadEnvironment(coll, r.env)

	results := make([]ExecutionResult, 0, len(coll.Requests))

	for _, req := range coll.Requests {
		result := r.ExecuteRequest(req)
		results = append(results, result)

		if result.Error == nil && result.Response.StatusCode < 400 {
			r.extractVariables(result.Response)
		}
	}

	return results, nil
}

func (r *Runner) ExecuteRequest(req Request) ExecutionResult {
	startTime := time.Now()

	url := r.Resolver.Resolve(req.URL)
	headers := make(map[string]string)
	for k, v := range req.Headers {
		headers[k] = r.Resolver.Resolve(v)
	}

	queryParams := make(map[string]string)
	for k, v := range req.QueryParams {
		queryParams[k] = r.Resolver.Resolve(v)
	}

	resolvedBody := r.Resolver.ResolveBody(req.Body)
	bodyBytes, err := BodyToBytes(resolvedBody)
	if err != nil {
		return ExecutionResult{
			Request:   req,
			Error:     fmt.Errorf("prepare body: %w", err),
			StartTime: startTime,
			EndTime:   time.Now(),
		}
	}

	if len(bodyBytes) > 0 && headers["Content-Type"] == "" {
		headers["Content-Type"] = "application/json"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := r.client.Do(ctx, &nexushttp.RequestOptions{
		Method:      req.Method,
		URL:         url,
		Headers:     headers,
		QueryParams: queryParams,
		Body:        bodyBytes,
	})

	endTime := time.Now()

	if err != nil {
		return ExecutionResult{
			Request:   req,
			Error:     err,
			StartTime: startTime,
			EndTime:   endTime,
		}
	}

	response := Response{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Headers:    resp.Headers,
		Body:       resp.Body,
		Time:       resp.Time,
		Size:       resp.Size,
	}

	passed, failures := r.runAssertions(req, response)

	return ExecutionResult{
		Request:   req,
		Response:  response,
		StartTime: startTime,
		EndTime:   endTime,
		Passed:    passed,
		Failures:  failures,
	}
}

func (r *Runner) runAssertions(req Request, resp Response) (bool, []string) {
	assertions := append(req.Tests, req.Assertions...)
	if len(assertions) == 0 {
		return true, nil
	}

	failures := []string{}

	for _, assertion := range assertions {
		if !r.evaluateAssertion(assertion, resp) {
			failures = append(failures, assertion)
		}
	}

	return len(failures) == 0, failures
}

func (r *Runner) evaluateAssertion(assertion string, resp Response) bool {
	assertion = strings.TrimSpace(assertion)

	if strings.Contains(assertion, "status") {
		return r.evalStatusAssertion(assertion, resp.StatusCode)
	}

	if strings.Contains(assertion, "body") {
		return r.evalBodyAssertion(assertion, resp.Body)
	}

	if strings.Contains(assertion, "time") || strings.Contains(assertion, "response.time") {
		return r.evalTimeAssertion(assertion, resp.Time)
	}

	return true
}

func (r *Runner) evalStatusAssertion(assertion string, status int) bool {
	re := regexp.MustCompile(`status\s*(==|!=|>|<|>=|<=)\s*(\d+)`)
	matches := re.FindStringSubmatch(assertion)
	if len(matches) < 3 {
		return true
	}

	op := matches[1]
	expected, _ := strconv.Atoi(matches[2])

	switch op {
	case "==":
		return status == expected
	case "!=":
		return status != expected
	case ">":
		return status > expected
	case "<":
		return status < expected
	case ">=":
		return status >= expected
	case "<=":
		return status <= expected
	}

	return true
}

func (r *Runner) evalBodyAssertion(assertion string, body []byte) bool {
	bodyStr := string(body)

	if strings.Contains(assertion, "contains") {
		re := regexp.MustCompile(`body\.contains\("([^"]+)"\)`)
		matches := re.FindStringSubmatch(assertion)
		if len(matches) > 1 {
			return strings.Contains(bodyStr, matches[1])
		}
	}

	if strings.Contains(assertion, "length") {
		re := regexp.MustCompile(`body\.length\s*(>|<|>=|<=|==)\s*(\d+)`)
		matches := re.FindStringSubmatch(assertion)
		if len(matches) > 2 {
			op := matches[1]
			expected, _ := strconv.Atoi(matches[2])
			length := len(body)

			switch op {
			case ">":
				return length > expected
			case "<":
				return length < expected
			case ">=":
				return length >= expected
			case "<=":
				return length <= expected
			case "==":
				return length == expected
			}
		}
	}

	return true
}

func (r *Runner) evalTimeAssertion(assertion string, duration time.Duration) bool {
	re := regexp.MustCompile(`(?:response\.)?time\s*<\s*(\d+)`)
	matches := re.FindStringSubmatch(assertion)
	if len(matches) > 1 {
		maxMs, _ := strconv.Atoi(matches[1])
		return duration < time.Duration(maxMs)*time.Millisecond
	}
	return true
}

func (r *Runner) extractVariables(resp Response) {
}
