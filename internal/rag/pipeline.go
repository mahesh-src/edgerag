package rag

import (
	"fmt"
	"strings"

	"edgerag/internal/embedding"
	"edgerag/internal/llm"
	"edgerag/internal/vectorstore"
)

// Pipeline represents the RAG pipeline
type Pipeline struct {
	embedder     *embedding.Service
	vectorStore  vectorstore.VectorStore
	llm          *llm.OllamaClient
	promptTemplate string
}

// NewPipeline creates a new RAG pipeline
func NewPipeline(embedder *embedding.Service, vectorStore vectorstore.VectorStore, llmClient *llm.OllamaClient) *Pipeline {
	return &Pipeline{
		embedder:    embedder,
		vectorStore: vectorStore,
		llm:         llmClient,
		promptTemplate: `You are a helpful assistant that answers questions based on the provided context. Use only the information given in the context to answer the question. If the context doesn't contain enough information to answer the question, say so.

Context:
{{.Context}}

Question: {{.Question}}

Answer:`,
	}
}

// SetPromptTemplate sets a custom prompt template
func (p *Pipeline) SetPromptTemplate(template string) {
	p.promptTemplate = template
}

// Query performs a RAG query: retrieve relevant documents and generate an answer
func (p *Pipeline) Query(question string, topK int, threshold float32) (string, []*vectorstore.SearchResult, error) {
	// Step 1: Generate embedding for the question
	questionEmbedding, err := p.embedder.GetEmbedding(question)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate question embedding: %w", err)
	}

	// Step 2: Retrieve relevant documents
	results, err := p.vectorStore.Search(questionEmbedding, topK, threshold)
	if err != nil {
		return "", nil, fmt.Errorf("failed to search vector store: %w", err)
	}

	if len(results) == 0 {
		return "I couldn't find any relevant information in the indexed documents to answer your question.", results, nil
	}

	// Step 3: Prepare context from retrieved documents
	context := p.buildContext(results)

	// Step 4: Generate answer using LLM
	prompt := p.buildPrompt(question, context)
	answer, err := p.llm.Generate(prompt)
	if err != nil {
		return "", results, fmt.Errorf("failed to generate answer: %w", err)
	}

	return strings.TrimSpace(answer), results, nil
}

// QueryStream performs a RAG query with streaming response
func (p *Pipeline) QueryStream(question string, topK int, threshold float32, callback func(string)) ([]*vectorstore.SearchResult, error) {
	// Step 1: Generate embedding for the question
	questionEmbedding, err := p.embedder.GetEmbedding(question)
	if err != nil {
		return nil, fmt.Errorf("failed to generate question embedding: %w", err)
	}

	// Step 2: Retrieve relevant documents
	results, err := p.vectorStore.Search(questionEmbedding, topK, threshold)
	if err != nil {
		return nil, fmt.Errorf("failed to search vector store: %w", err)
	}

	if len(results) == 0 {
		callback("I couldn't find any relevant information in the indexed documents to answer your question.")
		return results, nil
	}

	// Step 3: Prepare context from retrieved documents
	context := p.buildContext(results)

	// Step 4: Generate answer using LLM with streaming
	prompt := p.buildPrompt(question, context)
	err = p.llm.GenerateStream(prompt, callback)
	if err != nil {
		return results, fmt.Errorf("failed to generate streaming answer: %w", err)
	}

	return results, nil
}

// buildContext creates a context string from search results
func (p *Pipeline) buildContext(results []*vectorstore.SearchResult) string {
	var contextParts []string

	for i, result := range results {
		contextPart := fmt.Sprintf("Document %d (Similarity: %.3f):\n%s", 
			i+1, result.Score, result.Content)
		
		// Add file information if available
		if fileName, ok := result.Metadata["file"].(string); ok {
			contextPart = fmt.Sprintf("Document %d from %s (Similarity: %.3f):\n%s",
				i+1, fileName, result.Score, result.Content)
		}
		
		contextParts = append(contextParts, contextPart)
	}

	return strings.Join(contextParts, "\n\n---\n\n")
}

// buildPrompt creates the final prompt for the LLM
func (p *Pipeline) buildPrompt(question, context string) string {
	prompt := p.promptTemplate
	prompt = strings.ReplaceAll(prompt, "{{.Context}}", context)
	prompt = strings.ReplaceAll(prompt, "{{.Question}}", question)
	return prompt
}

// GetStats returns statistics about the RAG pipeline
func (p *Pipeline) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"vector_store_stats": p.vectorStore.GetStats(),
		"llm_model":         p.llm.GetModel(),
	}

	// Try to get embedding dimension
	if dimension, err := p.embedder.GetDimension(); err == nil {
		stats["embedding_dimension"] = dimension
	}

	return stats
} 