package vectorstore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// PersistentStore implements a persistent vector store that saves to disk
type PersistentStore struct {
	*MemoryStore
	dataDir string
}

// NewPersistentStore creates a new persistent vector store
func NewPersistentStore(dataDir string) (*PersistentStore, error) {
	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	store := &PersistentStore{
		MemoryStore: NewMemoryStore(),
		dataDir:     dataDir,
	}

	// Load existing vectors from disk
	if err := store.loadFromDisk(); err != nil {
		return nil, fmt.Errorf("failed to load vectors from disk: %w", err)
	}

	return store, nil
}

// Add stores a vector and persists it to disk
func (p *PersistentStore) Add(id string, embedding []float32, content string, metadata map[string]interface{}) error {
	// Add to memory first
	if err := p.MemoryStore.Add(id, embedding, content, metadata); err != nil {
		return err
	}

	// Persist to disk
	return p.saveToDisk(id)
}

// Delete removes a vector from memory and disk
func (p *PersistentStore) Delete(id string) error {
	// Remove from memory
	if err := p.MemoryStore.Delete(id); err != nil {
		return err
	}

	// Remove from disk
	filePath := filepath.Join(p.dataDir, id+".json")
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete vector file: %w", err)
	}

	return nil
}

// Clear removes all vectors from memory and disk
func (p *PersistentStore) Clear() {
	p.MemoryStore.Clear()

	// Remove all files from disk
	files, err := filepath.Glob(filepath.Join(p.dataDir, "*.json"))
	if err == nil {
		for _, file := range files {
			os.Remove(file)
		}
	}
}

// saveToDisk saves a single vector to disk
func (p *PersistentStore) saveToDisk(id string) error {
	p.mutex.RLock()
	vector, exists := p.vectors[id]
	p.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("vector with ID %s not found", id)
	}

	filePath := filepath.Join(p.dataDir, id+".json")
	data, err := json.MarshalIndent(vector, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal vector: %w", err)
	}

	return os.WriteFile(filePath, data, 0644)
}

// loadFromDisk loads all vectors from disk into memory
func (p *PersistentStore) loadFromDisk() error {
	files, err := filepath.Glob(filepath.Join(p.dataDir, "*.json"))
	if err != nil {
		return fmt.Errorf("failed to glob vector files: %w", err)
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue // Skip corrupted files
		}

		var vector Vector
		if err := json.Unmarshal(data, &vector); err != nil {
			continue // Skip corrupted files
		}

		p.mutex.Lock()
		p.vectors[vector.ID] = &vector
		p.mutex.Unlock()
	}

	return nil
}

// GetDataDir returns the data directory path
func (p *PersistentStore) GetDataDir() string {
	return p.dataDir
} 