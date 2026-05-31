package importkit

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Format — тип источника данных.
type Format string

const (
	FormatCSV  Format = "csv"
	FormatXLSX Format = "xlsx"
	FormatXML  Format = "xml"
	FormatYML  Format = "yml"  // Яндекс.Маркет
	FormatOneC Format = "onec" // 1С CommerceML
)

// FieldType — целевой тип значения после coerce.
type FieldType string

const (
	TypeString FieldType = "string"
	TypeInt    FieldType = "int"
	TypeFloat  FieldType = "float"
	TypeBool   FieldType = "bool"
	TypeDate   FieldType = "date"
)

// Config — корневая декларативная конфигурация импорта.
type Config struct {
	Format      Format         `yaml:"format" json:"format"`
	Source      SourceOptions  `yaml:"source" json:"source"`
	Mappings    []FieldMapping `yaml:"mappings" json:"mappings"`
	StrictMode  bool           `yaml:"strict" json:"strict"`
	SkipOnError bool           `yaml:"skip_on_error" json:"skip_on_error"`
}

// SourceOptions — общие опции источника. Поля интерпретируются конкретным Source.
type SourceOptions struct {
	Delimiter   string            `yaml:"delimiter" json:"delimiter"`   // CSV
	HasHeader   bool              `yaml:"has_header" json:"has_header"` // CSV/XLSX
	Sheet       string            `yaml:"sheet" json:"sheet"`           // XLSX
	HeaderRow   int               `yaml:"header_row" json:"header_row"` // XLSX (1-based)
	DataStart   int               `yaml:"data_start" json:"data_start"` // XLSX
	RootElement string            `yaml:"root" json:"root"`             // XML
	ItemElement string            `yaml:"item" json:"item"`             // XML
	Encoding    string            `yaml:"encoding" json:"encoding"`     // windows-1251 для 1С
	Extra       map[string]string `yaml:"extra" json:"extra"`           // расширения
}

// FieldMapping — описание одного поля.
type FieldMapping struct {
	Source    string    `yaml:"source" json:"source"` // имя колонки / XPath / атрибут
	Target    string    `yaml:"target" json:"target"` // имя ключа в Record
	Type      FieldType `yaml:"type" json:"type"`
	Required  bool      `yaml:"required" json:"required"`
	Default   any       `yaml:"default" json:"default"`
	Transform []string  `yaml:"transform" json:"transform"` // ["trim", "lower"]
	Validate  []string  `yaml:"validate" json:"validate"`   // ["min=0", "regex=^\\d+$"]
	Format    string    `yaml:"format" json:"format"`       // напр. "02.01.2006" для дат
}

// Validate проверяет согласованность конфига.
func (c *Config) Validate() error {
	if c.Format == "" {
		return fmt.Errorf("%w: format is empty", ErrConfigInvalid)
	}
	if len(c.Mappings) == 0 {
		return fmt.Errorf("%w: mappings is empty", ErrConfigInvalid)
	}
	seen := map[string]struct{}{}
	for i, m := range c.Mappings {
		if m.Source == "" || m.Target == "" {
			return fmt.Errorf("%w: mapping[%d] source/target required", ErrConfigInvalid, i)
		}
		if _, dup := seen[m.Target]; dup {
			return fmt.Errorf("%w: duplicated target %q", ErrConfigInvalid, m.Target)
		}
		seen[m.Target] = struct{}{}
		if m.Type == "" {
			c.Mappings[i].Type = TypeString
		}
	}
	return nil
}

// LoadConfig читает YAML или JSON в зависимости от содержимого первого непробельного символа.
func LoadConfig(r io.Reader) (*Config, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var cfg Config
	trimmed := strings.TrimSpace(string(data))
	if strings.HasPrefix(trimmed, "{") {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
	} else {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// LoadConfigFile — удобный shortcut.
func LoadConfigFile(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadConfig(f)
}
