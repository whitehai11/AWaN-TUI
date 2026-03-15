package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultBaseURL = "http://localhost:7452"

// Client talks to the AWaN Core local API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// AgentRunRequest is sent to /agent/run.
type AgentRunRequest struct {
	Agent  string `json:"agent"`
	Model  string `json:"model,omitempty"`
	Prompt string `json:"prompt"`
}

// AgentRunResponse is returned from /agent/run.
type AgentRunResponse struct {
	Agent     string `json:"agent"`
	Model     string `json:"model"`
	Output    string `json:"output"`
	StartedAt string `json:"startedAt"`
	EndedAt   string `json:"endedAt"`
}

// MemoryRecord represents a memory entry returned from /memory.
type MemoryRecord struct {
	ID        string `json:"id"`
	Agent     string `json:"agent"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
}

// MemorySnapshot is returned from /memory.
type MemorySnapshot struct {
	Agent      string         `json:"agent"`
	ShortTerm  []MemoryRecord `json:"shortTerm"`
	LongTerm   []MemoryRecord `json:"longTerm"`
	StoredAt   string         `json:"storedAt"`
	Vectorized bool           `json:"vectorized"`
}

// NewClient creates a runtime API client.
func NewClient(baseURL string) *Client {
	base := strings.TrimSpace(baseURL)
	if base == "" {
		base = defaultBaseURL
	}

	return &Client{
		baseURL: strings.TrimRight(base, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// BaseURL returns the configured runtime endpoint.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// RunAgent executes a prompt through the selected agent.
func (c *Client) RunAgent(request AgentRunRequest) (*AgentRunResponse, error) {
	var response AgentRunResponse
	if err := c.doJSON(http.MethodPost, "/agent/run", request, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// GetMemory fetches the current memory snapshot for an agent.
func (c *Client) GetMemory(agent string) (*MemorySnapshot, error) {
	path := "/memory"
	if strings.TrimSpace(agent) != "" {
		path += "?agent=" + url.QueryEscape(agent)
	}

	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("memory request failed with status %s", resp.Status)
	}

	var snapshot MemorySnapshot
	if err := json.NewDecoder(resp.Body).Decode(&snapshot); err != nil {
		return nil, err
	}

	return &snapshot, nil
}

func (c *Client) doJSON(method, path string, body any, target any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("%s request failed with status %s", path, resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}
