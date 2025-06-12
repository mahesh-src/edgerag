package vectorstore

// VectorStore defines the interface for vector storage and retrieval
type VectorStore interface {
	// Add stores a vector in the store
	Add(id string, embedding []float32, content string, metadata map[string]interface{}) error
	
	// Get retrieves a vector by ID
	Get(id string) (*Vector, error)
	
	// Delete removes a vector by ID
	Delete(id string) error
	
	// Search finds the most similar vectors to the query embedding
	Search(queryEmbedding []float32, topK int, threshold float32) ([]*SearchResult, error)
	
	// Count returns the number of vectors in the store
	Count() int
	
	// List returns all vector IDs
	List() []string
	
	// Clear removes all vectors from the store
	Clear()
	
	// GetStats returns statistics about the vector store
	GetStats() map[string]interface{}
} 