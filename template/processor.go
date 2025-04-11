package template

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Processor handles template processing for request data
type Processor struct {
	functions template.FuncMap
}

// NewProcessor creates a new template processor
func NewProcessor() *Processor {
	return &Processor{
		functions: template.FuncMap{
			"randomInt": func(min, max int) int {
				return rand.Intn(max-min+1) + min
			},
			"randomUUID": func() string {
				return fmt.Sprintf("%x-%x-%x-%x-%x",
					rand.Uint32(),
					uint16(rand.Uint32()),
					uint16(rand.Uint32()),
					uint16(rand.Uint32()),
					rand.Uint64())
			},
			"timestamp": func() string {
				return time.Now().Format(time.RFC3339)
			},
			"readCSV": func(filename string) string {
				file, err := os.Open(filename)
				if err != nil {
					return ""
				}
				defer file.Close()

				reader := csv.NewReader(file)
				records, err := reader.ReadAll()
				if err != nil || len(records) == 0 {
					return ""
				}

				// Return a random row from the CSV
				return records[rand.Intn(len(records))][0]
			},
			"env": func(key string) string {
				return os.Getenv(key)
			},
		},
	}
}

// ProcessTemplate processes a template string with the given data
func (p *Processor) ProcessTemplate(tmpl string, data interface{}) (string, error) {
	t, err := template.New("request").Funcs(sprig.FuncMap()).Funcs(p.functions).Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("error parsing template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("error executing template: %w", err)
	}

	return buf.String(), nil
}

// ProcessMap processes all string values in a map that contain template expressions
func (p *Processor) ProcessMap(m map[string]string) (map[string]string, error) {
	result := make(map[string]string)
	for k, v := range m {
		processed, err := p.ProcessTemplate(v, nil)
		if err != nil {
			return nil, fmt.Errorf("error processing template for key %s: %w", k, err)
		}
		result[k] = processed
	}
	return result, nil
}
