package embedding

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// Service handles text embeddings using sentence-transformers via Python
type Service struct {
	model      string
	scriptPath string
	cmd        *exec.Cmd
	stdin      io.WriteCloser
	stdout     *bufio.Scanner
}

// EmbeddingRequest represents the request structure for the Python script
type EmbeddingRequest struct {
	Text  string `json:"text"`
	Model string `json:"model"`
}

// EmbeddingResponse represents the response structure from the Python script
type EmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
	Error     string    `json:"error,omitempty"`
	Status    string    `json:"status,omitempty"`
}

// NewService creates a new embedding service
func NewService(model string) (*Service, error) {
	// Get the path to the Python script
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("failed to get current file path")
	}
	
	scriptPath := filepath.Join(filepath.Dir(currentFile), "..", "..", "scripts", "embeddings.py")
	
	service := &Service{
		model:      model,
		scriptPath: scriptPath,
	}

	// Start the persistent Python process
	if err := service.start(); err != nil {
		return nil, fmt.Errorf("failed to start embedding service: %w", err)
	}

	// Test the service
	if err := service.test(); err != nil {
		service.Close()
		return nil, fmt.Errorf("embedding service test failed: %w", err)
	}

	return service, nil
}

// start initializes the persistent Python process
func (s *Service) start() error {
	s.cmd = exec.Command("python3", s.scriptPath)
	
	// Set up pipes
	stdin, err := s.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	s.stdin = stdin
	
	stdout, err := s.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	s.stdout = bufio.NewScanner(stdout)
	
	// Start the process
	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start Python process: %w", err)
	}
	
	// Wait for ready signal
	if !s.stdout.Scan() {
		return fmt.Errorf("failed to read ready signal")
	}
	
	var readyResponse EmbeddingResponse
	if err := json.Unmarshal(s.stdout.Bytes(), &readyResponse); err != nil {
		return fmt.Errorf("failed to parse ready signal: %w", err)
	}
	
	if readyResponse.Status != "ready" {
		return fmt.Errorf("unexpected ready signal: %s", readyResponse.Status)
	}
	
	return nil
}

// GetEmbedding generates an embedding for the given text
func (s *Service) GetEmbedding(text string) ([]float32, error) {
	request := EmbeddingRequest{
		Text:  text,
		Model: s.model,
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request to Python process
	if _, err := fmt.Fprintf(s.stdin, "%s\n", requestJSON); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Read response with timeout
	done := make(chan bool)
	var response EmbeddingResponse
	var scanErr error

	go func() {
		if s.stdout.Scan() {
			scanErr = json.Unmarshal(s.stdout.Bytes(), &response)
		} else {
			scanErr = fmt.Errorf("failed to read response")
		}
		done <- true
	}()

	select {
	case <-done:
		if scanErr != nil {
			return nil, fmt.Errorf("failed to parse response: %w", scanErr)
		}
	case <-time.After(2 * time.Minute):
		return nil, fmt.Errorf("embedding generation timed out after 2 minutes")
	}

	if response.Error != "" {
		return nil, fmt.Errorf("embedding error: %s", response.Error)
	}

	return response.Embedding, nil
}

// GetEmbeddings generates embeddings for multiple texts
func (s *Service) GetEmbeddings(texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	
	for i, text := range texts {
		embedding, err := s.GetEmbedding(text)
		if err != nil {
			return nil, fmt.Errorf("failed to get embedding for text %d: %w", i, err)
		}
		embeddings[i] = embedding
	}
	
	return embeddings, nil
}

// test verifies that the embedding service is working
func (s *Service) test() error {
	testText := "This is a test sentence for embedding generation."
	_, err := s.GetEmbedding(testText)
	return err
}

// GetDimension returns the dimension of embeddings for the current model
func (s *Service) GetDimension() (int, error) {
	// Get a test embedding to determine dimension
	testEmbedding, err := s.GetEmbedding("test")
	if err != nil {
		return 0, err
	}
	return len(testEmbedding), nil
}

// Close shuts down the persistent Python process
func (s *Service) Close() error {
	if s.stdin != nil {
		// Send quit signal
		fmt.Fprintf(s.stdin, "QUIT\n")
		s.stdin.Close()
	}
	
	if s.cmd != nil && s.cmd.Process != nil {
		// Wait for process to exit gracefully
		done := make(chan error)
		go func() {
			done <- s.cmd.Wait()
		}()
		
		select {
		case <-done:
			// Process exited gracefully
		case <-time.After(5 * time.Second):
			// Force kill if it doesn't exit
			s.cmd.Process.Kill()
		}
	}
	
	return nil
} 