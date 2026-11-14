// Package piston provides an HTTP client for the Piston code execution API (v2).
package piston

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Request is the JSON body sent to the Piston execute endpoint.
type Request struct {
	Language           string `json:"language"`
	Version            string `json:"version"`
	Files              []File `json:"files"`
	Stdin              string `json:"stdin"`
	RunTimeoutMs       *int   `json:"run_timeout,omitempty"`
	RunMemoryLimit     *int   `json:"run_memory_limit,omitempty"`
	CompileMemoryLimit *int   `json:"compile_memory_limit,omitempty"`
}

// File represents a single source file in a Piston request.
type File struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// Response is the top-level JSON response from the Piston execute endpoint.
type Response struct {
	Language string `json:"language"`
	Version  string `json:"version"`
	Run      *Stage `json:"run"`
	Compile  *Stage `json:"compile"`
}

// Stage represents a single stage (compile or run) in a Piston response.
type Stage struct {
	Stdout   string  `json:"stdout"`
	Stderr   string  `json:"stderr"`
	Code     int     `json:"code"`
	Signal   *string `json:"signal"`
	Output   string  `json:"output"`
	CPUTime  float64 `json:"cpu_time"`
	WallTime float64 `json:"wall_time"`
}

// Client is an HTTP client for the Piston code execution API.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new Piston API client with the given base URL.
func NewClient(baseURL string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: baseURL,
	}
}

// Execute sends a code execution request to the Piston API.
// If runTimeoutMs > 0, the run_timeout field is set in the request.
func (c *Client) Execute(ctx context.Context, language, version, fileName, code, stdin string, runTimeoutMs, memoryLimitMB int) (*Response, error) {
	req := Request{
		Language: language,
		Version:  version,
		Files: []File{
			{Name: fileName, Content: code},
		},
		Stdin: stdin,
	}
	if runTimeoutMs > 0 {
		req.RunTimeoutMs = &runTimeoutMs
	}

	if memoryLimitMB > 0 {
		limit := memoryLimitMB * 1024 * 1024
		req.RunMemoryLimit = &limit
		req.CompileMemoryLimit = &limit
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("piston: failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v2/execute", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("piston: failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("piston: request failed: %w", err)
	}
	defer resp.Body.Close()

	const maxResponseBytes = 4 * 1024 * 1024
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("piston: failed to read response: %w", err)
	}
	if len(respBody) > maxResponseBytes {
		return nil, fmt.Errorf("piston: response exceeds size limit")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("piston: unexpected status %d", resp.StatusCode)
	}

	var result Response
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("piston: failed to decode response: %w", err)
	}
	type stageStatus struct {
		Code *int `json:"code"`
	}
	var stages struct{ Run, Compile *stageStatus }
	if err = json.Unmarshal(respBody, &stages); err != nil {
		return nil, fmt.Errorf("piston: invalid execution response")
	}
	if (stages.Run != nil && stages.Run.Code == nil) || (stages.Compile != nil && stages.Compile.Code == nil) || (result.Run == nil && (result.Compile == nil || result.Compile.Code == 0)) {
		return nil, fmt.Errorf("piston: missing execution result")
	}
	if result.Compile != nil && result.Compile.Code != 0 && result.Compile.Stderr == "" {
		result.Compile.Stderr = "Compilation error"
	}

	return &result, nil
}
