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
func (e *Evaluator) Elevate(htmlContent string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	result := ""
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if e.matches(n) {
				result = e.getText(n)
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

func (e *Evaluator) matches(n *html.Node) bool {
	selector := e.TargetSelector
	if strings.HasPrefix(selector, "#") {
		id := strings.TrimPrefix(selector, "#")
		for _, attr := range n.Attr {
			if attr.Key == "id" && attr.Val == id {
				return true
			}
		}
		return false
	}

	if strings.HasPrefix(selector, ".") {
		class := strings.TrimPrefix(selector, ".")
		for _, attr := range n.Attr {
			if attr.Key == "class" && strings.Contains(attr.Val, class) {
				return true
			}
		}
		return false
	}

	return n.Data == selector
}

func (e *Evaluator) getText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(e.getText(c))
	}
	return sb.String()
}
