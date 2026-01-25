package service

import (
	"fmt"
	"regexp"
	"strings"
)

// Renderer defines the contract for template parsing
type Renderer interface {
	Render(template string, data map[string]interface{}) string
}

type templateEngine struct {
	pattern *regexp.Regexp
}

func NewTemplateEngine() Renderer {
	return &templateEngine{
		// Regex to find placeholders like {{ firstName }} or {{amount}}
		pattern: regexp.MustCompile(`\{\{\s*(\w+)\s*\}\}`),
	}
}

// Render replaces placeholders in the template string with values from the data map
func (e *templateEngine) Render(template string, data map[string]interface{}) string {
	return e.pattern.ReplaceAllStringFunc(template, func(match string) string {
		// Extract key: "{{ name }}" -> "name"
		key := strings.Trim(match, "{} ")
		if val, ok := data[key]; ok {
			return fmt.Sprintf("%v", val)
		}
		// If key is missing, return the placeholder for debugging purposes
		return match
	})
}
