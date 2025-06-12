package vectorstore

import (
	"fmt"
	"math"
	"sort"
	"sync"
)

// Vector represents a stored vector with metadata
type Vector struct {
	ID        string                 `json:"id"`
	Embedding []float32              `json:"embedding"`
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// SearchResult represents a search result with similarity score
type SearchResult struct {
	Vector
	Score float32 `json:"score"`
}

// MemoryStore implements an in-memory vector store
type MemoryStore struct {
	vectors map[string]*Vector
	mutex   sync.RWMutex
}

// NewMemoryStore creates a new in-memory vector store
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		vectors: make(map[string]*Vector),
	}
}

// Add stores a vector in the memory store
func (m *MemoryStore) Add(id string, embedding []float32, content string, metadata map[string]interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	m.vectors[id] = &Vector{
		ID:        id,
		Embedding: embedding,
		Content:   content,
		Metadata:  metadata,
	}

	return nil
}

// Get retrieves a vector by ID
func (m *MemoryStore) Get(id string) (*Vector, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	vector, exists := m.vectors[id]
	if !exists {
		return nil, fmt.Errorf("vector with ID %s not found", id)
	}

	return vector, nil
}

// Delete removes a vector by ID
func (m *MemoryStore) Delete(id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.vectors[id]; !exists {
		return fmt.Errorf("vector with ID %s not found", id)
	}

	delete(m.vectors, id)
	return nil
}

// Search finds the most similar vectors to the query embedding
func (m *MemoryStore) Search(queryEmbedding []float32, topK int, threshold float32) ([]*SearchResult, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if len(m.vectors) == 0 {
		return []*SearchResult{}, nil
	}

	results := make([]*SearchResult, 0)

	// Calculate similarity scores for all vectors
	for _, vector := range m.vectors {
		similarity := cosineSimilarity(queryEmbedding, vector.Embedding)
		
		if similarity >= threshold {
			results = append(results, &SearchResult{
				Vector: *vector,
				Score:  similarity,
			})
		}
	}

	// Sort by similarity score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Limit to topK results
	if len(results) > topK {
		results = results[:topK]
	}

	return results, nil
}

// Count returns the number of vectors in the store
func (m *MemoryStore) Count() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.vectors)
}

// List returns all vector IDs
func (m *MemoryStore) List() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	ids := make([]string, 0, len(m.vectors))
	for id := range m.vectors {
		ids = append(ids, id)
	}
	return ids
}

// Clear removes all vectors from the store
func (m *MemoryStore) Clear() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.vectors = make(map[string]*Vector)
}

// cosineSimilarity calculates the cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProduct, normA, normB float32

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0.0 || normB == 0.0 {
		return 0.0
	}

	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

// GetStats returns statistics about the vector store
func (m *MemoryStore) GetStats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_vectors": len(m.vectors),
	}

	if len(m.vectors) > 0 {
		// Get dimension from first vector
		for _, vector := range m.vectors {
			stats["dimension"] = len(vector.Embedding)
			break
		}
	}

	return stats
} 