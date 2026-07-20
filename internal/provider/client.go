package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const defaultBaseURL = "https://bwwzefaddmyueynayize.supabase.co/functions/v1"

// ErrUnauthorized is returned by ListProjects when the ZiCON API responds
// with 401, which means the project's API key has been revoked or the
// project itself no longer exists.
var ErrUnauthorized = errors.New("zicon: unauthorized")

// Client is a minimal HTTP client for the ZiCON Cloud API.
type Client struct {
	BaseURL     string
	AccessToken string
	HTTPClient  *http.Client
}

// NewClient creates a ZiCON API client authenticated with a Supabase
// session token, used for project creation.
func NewClient(accessToken string) *Client {
	return &Client{
		BaseURL:     defaultBaseURL,
		AccessToken: strings.TrimSpace(accessToken),
		HTTPClient:  http.DefaultClient,
	}
}

// Project represents a ZiCON project as returned by the API. Not every
// endpoint populates every field: create-project includes user_id and
// api_key, while list-projects omits both.
type Project struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Status      string `json:"status"`
	APIKey      string `json:"api_key"`
	CreatedAt   string `json:"created_at"`
}

// CreateProject calls POST create-project using session-token auth.
func (c *Client) CreateProject(ctx context.Context, name, category, description string) (*Project, error) {
	body, err := json.Marshal(map[string]string{
		"name":        name,
		"category":    category,
		"description": description,
	})
	if err != nil {
		return nil, fmt.Errorf("encoding create-project request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/create-project", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("building create-project request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	respBody, err := c.do(req)
	if err != nil {
		return nil, err
	}

	var project Project
	if err := json.Unmarshal(respBody, &project); err != nil {
		return nil, fmt.Errorf("decoding create-project response: %w", err)
	}
	return &project, nil
}

// ListProjects calls GET list-projects using the project's API key. The
// endpoint is scoped to the key's own project and returns it as a bare
// JSON object (not an array). Returns ErrUnauthorized if the key has been
// revoked or the project was deleted.
func (c *Client) ListProjects(ctx context.Context, apiKey string) (*Project, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/list-projects", nil)
	if err != nil {
		return nil, fmt.Errorf("building list-projects request: %w", err)
	}
	req.Header.Set("X-ZiCON-Key", apiKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling list-projects: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading list-projects response: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("list-projects returned %d: %s", resp.StatusCode, string(respBody))
	}

	var project Project
	if err := json.Unmarshal(respBody, &project); err != nil {
		return nil, fmt.Errorf("decoding list-projects response: %w", err)
	}

	return &project, nil
}

// DeleteProject calls DELETE delete-project?id=... using the project's API key.
func (c *Client) DeleteProject(ctx context.Context, apiKey, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.BaseURL+"/delete-project?id="+id, nil)
	if err != nil {
		return fmt.Errorf("building delete-project request: %w", err)
	}
	req.Header.Set("X-ZiCON-Key", apiKey)

	respBody, err := c.do(req)
	if err != nil {
		return err
	}

	var result struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("decoding delete-project response: %w", err)
	}
	if !result.Success {
		return fmt.Errorf("delete-project did not report success: %s", string(respBody))
	}

	return nil
}

// do executes a request and returns the response body, treating any
// non-2xx status code as an error.
func (c *Client) do(req *http.Request) ([]byte, error) {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling %s: %w", req.URL.Path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response from %s: %w", req.URL.Path, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s returned %d: %s", req.URL.Path, resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
