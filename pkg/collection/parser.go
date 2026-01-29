package collection

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) ParseFile(path string) (*Collection, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		return p.ParseYAML(data)
	case ".json":
		return p.ParseJSON(data)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}
}

func (p *Parser) ParseYAML(data []byte) (*Collection, error) {
	var coll Collection
	if err := yaml.Unmarshal(data, &coll); err != nil {
		return nil, fmt.Errorf("unmarshal yaml: %w", err)
	}
	return &coll, nil
}

func (p *Parser) ParseJSON(data []byte) (*Collection, error) {
	var coll Collection
	if err := json.Unmarshal(data, &coll); err != nil {
		return nil, fmt.Errorf("unmarshal json: %w", err)
	}
	return &coll, nil
}

func (p *Parser) SaveFile(coll *Collection, path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	var data []byte
	var err error

	switch ext {
	case ".yaml", ".yml":
		data, err = yaml.Marshal(coll)
	case ".json":
		data, err = json.MarshalIndent(coll, "", "  ")
	default:
		return fmt.Errorf("unsupported file format: %s", ext)
	}

	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}
