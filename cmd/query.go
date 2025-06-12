package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"edgerag/internal/embedding"
	"edgerag/internal/llm"
	"edgerag/internal/rag"
	"edgerag/internal/vectorstore"
)

var queryCmd = &cobra.Command{
	Use:   "query [question]",
	Short: "Query the indexed documents using RAG",
	Long: `Query the indexed documents using Retrieval-Augmented Generation (RAG).
This command will:
1. Generate an embedding for your question
2. Find the most relevant document chunks
3. Use Ollama LLM to generate an answer based on the retrieved context

Examples:
  edgerag query "How do I initialize a Go module?"
  edgerag query "What are the main features of this project?" --top-k 5`,
	Args: cobra.ExactArgs(1),
	RunE: runQuery,
}

func init() {
	rootCmd.AddCommand(queryCmd)
	
	queryCmd.Flags().IntP("top-k", "k", 3, "Number of most relevant chunks to retrieve")
	queryCmd.Flags().Float32P("threshold", "t", 0.3, "Similarity threshold for retrieval")
	queryCmd.Flags().StringP("prompt-template", "p", "", "Custom prompt template for LLM")
	queryCmd.Flags().BoolP("show-sources", "s", true, "Show source documents in the response")
}

func runQuery(cmd *cobra.Command, args []string) error {
	question := args[0]
	topK, _ := cmd.Flags().GetInt("top-k")
	threshold, _ := cmd.Flags().GetFloat32("threshold")
	promptTemplate, _ := cmd.Flags().GetString("prompt-template")
	showSources, _ := cmd.Flags().GetBool("show-sources")

	// Initialize services
	model := viper.GetString("model")
	embeddingService, err := embedding.NewService(model)
	if err != nil {
		return fmt.Errorf("failed to initialize embedding service: %w", err)
	}
	defer embeddingService.Close()

	// Initialize persistent vector store
	dataDir := filepath.Join(os.Getenv("HOME"), ".edgerag", "vectors")
	vectorStore, err := vectorstore.NewPersistentStore(dataDir)
	if err != nil {
		return fmt.Errorf("failed to initialize vector store: %w", err)
	}
	
	if vectorStore.Count() == 0 {
		return fmt.Errorf("no documents indexed. Please run 'edgerag index' first")
	}

	ollamaURL := viper.GetString("ollama_url")
	ollamaModel := viper.GetString("ollama_model")
	llmClient, err := llm.NewOllamaClient(ollamaURL, ollamaModel)
	if err != nil {
		return fmt.Errorf("failed to initialize Ollama client: %w", err)
	}

	// Initialize RAG pipeline
	ragPipeline := rag.NewPipeline(embeddingService, vectorStore, llmClient)

	// Set custom prompt template if provided
	if promptTemplate != "" {
		ragPipeline.SetPromptTemplate(promptTemplate)
	}

	fmt.Printf("🔍 Searching for relevant information...\n")

	// Perform RAG query
	response, sources, err := ragPipeline.Query(question, topK, threshold)
	if err != nil {
		return fmt.Errorf("failed to process query: %w", err)
	}

	// Display results
	fmt.Printf("\n📖 Answer:\n")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println(response)
	fmt.Println(strings.Repeat("-", 80))

	if showSources && len(sources) > 0 {
		fmt.Printf("\n📚 Sources (%d found):\n", len(sources))
		for i, source := range sources {
			fmt.Printf("\n[%d] Similarity: %.3f\n", i+1, source.Score)
			if source.Metadata["file"] != nil {
				fmt.Printf("File: %s\n", source.Metadata["file"])
			}
			if source.Metadata["chunk_id"] != nil {
				fmt.Printf("Chunk: %s\n", source.Metadata["chunk_id"])
			}
			fmt.Printf("Content: %s...\n", truncateString(source.Content, 200))
		}
	}

	return nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
} 