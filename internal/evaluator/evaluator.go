package evaluator

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

type Evaluator struct {
	TargetSelector string
}

func NewEvaluator(selector string) *Evaluator {
	return &Evaluator{
		TargetSelector: selector,
	}
}

// Elevate extracts data from the HTML based on the target selector.
// In this context, "elevation" refers to promoting raw HTML to structured data.
func (e *Evaluator) Elevate(htmlContent string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	result := ""
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == e.TargetSelector {
			if n.FirstChild != nil {
				result = n.FirstChild.Data
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
			if result != "" {
				return
			}
		}
	}
	f(doc)

	if result == "" {
		return "", fmt.Errorf("target element %s not found", e.TargetSelector)
	}

	return result, nil
}
