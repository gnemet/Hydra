package fetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Fetcher struct {
	client *http.Client
}

func NewFetcher(timeoutSeconds int) *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
	}
}

func (f *Fetcher) Fetch(targetURL string) (string, error) {
	return f.doRequest("GET", targetURL, nil)
}

func (f *Fetcher) FetchPost(targetURL string, data map[string]string) (string, error) {
	return f.doRequest("POST", targetURL, data)
}

func (f *Fetcher) doRequest(method, targetURL string, data map[string]string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), f.client.Timeout)
	defer cancel()

	var bodyReader io.Reader
	if method == "POST" && data != nil {
		form := url.Values{}
		for k, v := range data {
			form.Set(k, v)
		}
		bodyReader = strings.NewReader(form.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, method, targetURL, bodyReader)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	if method == "POST" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}
