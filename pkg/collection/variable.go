package collection

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type VariableResolver struct {
	env       string
	variables map[string]string
	globals   map[string]string
}

func NewVariableResolver(env string) *VariableResolver {
	return &VariableResolver{
		env:       env,
		variables: make(map[string]string),
		globals:   make(map[string]string),
	}
}

func (vr *VariableResolver) LoadEnvironment(coll *Collection, envName string) {
	if env, ok := coll.Environment[envName]; ok {
		vr.variables["baseUrl"] = env.BaseURL
		for k, v := range env.Variables {
			vr.variables[k] = v
		}
	}
	if coll.BaseURL != "" && vr.variables["baseUrl"] == "" {
		vr.variables["baseUrl"] = coll.BaseURL
	}
}

func (vr *VariableResolver) SetGlobal(key, value string) {
	vr.globals[key] = value
}

func (vr *VariableResolver) SetVariable(key, value string) {
	vr.variables[key] = value
}

func (vr *VariableResolver) Resolve(s string) string {
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		key := strings.TrimSpace(match[2 : len(match)-2])

		if strings.HasPrefix(key, "$") {
			return vr.resolveFunction(key)
		}

		if val, ok := vr.variables[key]; ok {
			return val
		}

		if val, ok := vr.globals[key]; ok {
			return val
		}

		if val := os.Getenv(key); val != "" {
			return val
		}

		return match
	})
}

func (vr *VariableResolver) ResolveBody(body interface{}) interface{} {
	switch v := body.(type) {
	case string:
		return vr.Resolve(v)
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, val := range v {
			result[k] = vr.ResolveBody(val)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = vr.ResolveBody(val)
		}
		return result
	default:
		return v
	}
}

func (vr *VariableResolver) resolveFunction(fn string) string {
	switch fn {
	case "$randomInt":
		return fmt.Sprintf("%d", randomInt(0, 1000000))
	case "$randomUUID":
		return randomUUID()
	case "$randomEmail":
		return fmt.Sprintf("user%d@example.com", randomInt(1000, 9999))
	case "$randomName":
		names := []string{"Alice", "Bob", "Charlie", "Diana", "Eve", "Frank"}
		return names[randomInt(0, len(names))]
	case "$timestamp":
		return fmt.Sprintf("%d", currentTimestamp())
	default:
		return fn
	}
}

func BodyToBytes(body interface{}) ([]byte, error) {
	if body == nil {
		return nil, nil
	}

	switch v := body.(type) {
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	case map[string]interface{}, []interface{}:
		return json.Marshal(v)
	default:
		return json.Marshal(v)
	}
}

func FormatJSON(data []byte) (string, error) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, data, "", "  "); err != nil {
		return string(data), err
	}
	return buf.String(), nil
}
