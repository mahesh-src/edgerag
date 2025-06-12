package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OllamaClient represents a client for the Ollama API
type OllamaClient struct {
	baseURL string
	model   string
	client  *http.Client
}

// OllamaRequest represents a request to the Ollama API
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// OllamaResponse represents a response from the Ollama API
type OllamaResponse struct {
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	Response  string    `json:"response"`
	Done      bool      `json:"done"`
	Context   []int     `json:"context,omitempty"`
}

// NewOllamaClient creates a new Ollama client
func NewOllamaClient(baseURL, model string) (*OllamaClient, error) {
	client := &OllamaClient{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		model:   model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}

	// Test the connection
	if err := client.ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to Ollama: %w", err)
	}

	return client, nil
}

// Generate generates text using the Ollama model
func (c *OllamaClient) Generate(prompt string) (string, error) {
	request := OllamaRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.client.Post(c.baseURL+"/api/generate", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Response, nil
}

// GenerateStream generates text using the Ollama model with streaming
func (c *OllamaClient) GenerateStream(prompt string, callback func(string)) error {
	request := OllamaRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: true,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.client.Post(c.baseURL+"/api/generate", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	decoder := json.NewDecoder(resp.Body)
	for {
		var response OllamaResponse
		if err := decoder.Decode(&response); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode streaming response: %w", err)
		}

		if response.Response != "" {
			callback(response.Response)
		}

		if response.Done {
			break
		}
	}

	return nil
}

// ping tests the connection to the Ollama server
func (c *OllamaClient) ping() error {
	resp, err := c.client.Get(c.baseURL + "/api/tags")
	if err != nil {
		return fmt.Errorf("failed to connect to Ollama server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama server returned status %d", resp.StatusCode)
	}

	return nil
}

// ListModels returns a list of available models
func (c *OllamaClient) ListModels() ([]string, error) {
	resp, err := c.client.Get(c.baseURL + "/api/tags")
	if err != nil {
		return nil, fmt.Errorf("failed to get models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get models, status: %d", resp.StatusCode)
	}

	var response struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	models := make([]string, len(response.Models))
	for i, model := range response.Models {
		models[i] = model.Name
	}

	return models, nil
}

// SetModel changes the model used for generation
func (c *OllamaClient) SetModel(model string) {
	c.model = model
}

// GetModel returns the current model
func (c *OllamaClient) GetModel() string {
	return c.model
} 