package kiwi_sdk

import (
	"fmt"
	"regexp"
	"strings"
)

type FilterParams map[string]any

type Filter struct {
	Params  FilterParams
	Content string
}

// NewFilter creates a new filter.
func NewFilter(content string, params FilterParams) *Filter {
	return &Filter{
		Content: content,
		Params:  params,
	}
}

// Build builds the filter.
// example:
// content: "id = {:id}"
// params: map[string]any{"id": "123"}
// result: "id = '123'"
func (f *Filter) Build() string {
	result := f.Content
	for key, value := range f.Params {
		placeholderRegex := regexp.MustCompile(`\{:` + key + `\}`)

		var replacement string
		switch v := value.(type) {
		case string:
			replacement = fmt.Sprintf("'%s'", v)
		default:
			replacement = fmt.Sprintf("%v", v)
		}

		result = placeholderRegex.ReplaceAllString(result, replacement)
	}

	result = strings.ReplaceAll(result, " ", "")
	result = "(" + result + ")"

	return result
}
