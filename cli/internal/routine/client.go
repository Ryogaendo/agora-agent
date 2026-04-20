package routine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	baseURL    = "https://api.anthropic.com/v1/claude_code/routines"
	betaHeader = "experimental-cc-routine-2026-04-01"
	apiVersion = "2023-06-01"
)

// FireRequest is the body sent to the /fire endpoint.
type FireRequest struct {
	Text string `json:"text,omitempty"`
}

// FireResponse is the response from a successful /fire call.
type FireResponse struct {
	Type             string `json:"type"`
	SessionID        string `json:"claude_code_session_id"`
	SessionURL       string `json:"claude_code_session_url"`
}

// Fire triggers a routine via its trigger ID and bearer token.
// An optional text payload provides run-specific context.
func Fire(triggerID, token, text string) (*FireResponse, error) {
	url := fmt.Sprintf("%s/%s/fire", baseURL, triggerID)

	var body io.Reader
	if text != "" {
		payload, err := json.Marshal(FireRequest{Text: text})
		if err != nil {
			return nil, fmt.Errorf("marshaling request: %w", err)
		}
		body = bytes.NewReader(payload)
	} else {
		body = bytes.NewReader([]byte("{}"))
	}

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("anthropic-beta", betaHeader)
	req.Header.Set("anthropic-version", apiVersion)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var result FireResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &result, nil
}
