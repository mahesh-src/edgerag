package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"edgerag/internal/document"
	"edgerag/internal/embedding"
	"edgerag/internal/vectorstore"
)

var indexCmd = &cobra.Command{
	Use:   "index [file/directory]",
	Short: "Index documents for retrieval",
	Long: `Index documents by processing them, generating embeddings, and storing them in the vector database.
	
Supported file formats:
- .txt (plain text)
- .md (markdown)
- .go (Go source code)
- .py (Python source code)
- .js (JavaScript source code)

Examples:
  edgerag index ./docs
  edgerag index file.txt
  edgerag index . --recursive`,
	Args: cobra.ExactArgs(1),
	RunE: runIndex,
}

func init() {
	rootCmd.AddCommand(indexCmd)
	
	indexCmd.Flags().BoolP("recursive", "r", false, "Recursively index directories")
	indexCmd.Flags().StringSliceP("extensions", "e", []string{".txt", ".md", ".go", ".py", ".js"}, "File extensions to index")
	indexCmd.Flags().IntP("chunk-size", "c", 200, "Maximum chunk size for document splitting")
	indexCmd.Flags().IntP("chunk-overlap", "o", 50, "Overlap between chunks when splitting documents")
}

func runIndex(cmd *cobra.Command, args []string) error {
	path := args[0]
	recursive, _ := cmd.Flags().GetBool("recursive")
	extensions, _ := cmd.Flags().GetStringSlice("extensions")
	chunkSize, _ := cmd.Flags().GetInt("chunk-size")
	chunkOverlap, _ := cmd.Flags().GetInt("chunk-overlap")

	// Initialize embedding service
	model := viper.GetString("model")
	fmt.Printf("🧠 Initializing embedding service (model: %s)...\n", model)
	fmt.Printf("   Note: First run may take longer as the model downloads\n")
	embeddingService, err := embedding.NewService(model)
	if err != nil {
		return fmt.Errorf("failed to initialize embedding service: %w", err)
	}
	defer embeddingService.Close()
	fmt.Printf("✅ Embedding service ready\n")

	// Initialize vector store
	fmt.Printf("💾 Initializing vector store...\n")
	dataDir := filepath.Join(os.Getenv("HOME"), ".edgerag", "vectors")
	vectorStore, err := vectorstore.NewPersistentStore(dataDir)
	if err != nil {
		return fmt.Errorf("failed to initialize vector store: %w", err)
	}
	fmt.Printf("✅ Vector store ready (data dir: %s)\n", dataDir)
	fmt.Println()

	// Get files to process
	files, err := getFilesToProcess(path, recursive, extensions)
	if err != nil {
		return fmt.Errorf("failed to get files: %w", err)
	}

	fmt.Printf("Found %d files to index\n", len(files))

	// Process each file
	for i, file := range files {
		fmt.Printf("Processing [%d/%d] %s\n", i+1, len(files), file)

		// Load and chunk document
		fmt.Printf("  ⏳ Loading document...")
		doc, err := document.LoadFromFile(file)
		if err != nil {
			fmt.Printf(" ❌ Failed to load %s: %v\n", file, err)
			continue
		}
		fmt.Printf(" ✅ Loaded (%d bytes)\n", len(doc.Content))

		fmt.Printf("  ⏳ Chunking document...")
		chunks := document.ChunkDocument(doc, chunkSize, chunkOverlap)
		fmt.Printf(" ✅ Created %d chunks\n", len(chunks))

		// Generate embeddings for each chunk
		fmt.Printf("  ⏳ Generating embeddings...\n")
		for j, chunk := range chunks {
			fmt.Printf("    [%d/%d] Embedding chunk %d (%.1f%%)...", 
				j+1, len(chunks), j+1, float64(j+1)/float64(len(chunks))*100)
			
			embedding, err := embeddingService.GetEmbedding(chunk.Content)
			if err != nil {
				fmt.Printf(" ❌ Failed: %v\n", err)
				continue
			}
			fmt.Printf(" ✅ Done (%d dims)\n", len(embedding))

			// Store in vector database
			err = vectorStore.Add(chunk.ID, embedding, chunk.Content, chunk.Metadata)
			if err != nil {
				fmt.Printf("    ❌ Failed to store chunk %d: %v\n", j, err)
				continue
			}
		}
		fmt.Printf("  ✅ Completed file %s (%d vectors stored)\n", file, len(chunks))
		fmt.Println()
	}

	fmt.Printf("\nIndexing complete! Indexed %d documents with %d total vectors\n", 
		len(files), vectorStore.Count())

	return nil
}

func getFilesToProcess(path string, recursive bool, extensions []string) ([]string, error) {
	var files []string
	
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		// Single file
		if hasValidExtension(path, extensions) {
			files = append(files, path)
		}
		return files, nil
	}

	// Directory
	if recursive {
		err = filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && hasValidExtension(filePath, extensions) {
				files = append(files, filePath)
			}
			return nil
		})
	} else {
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				filePath := filepath.Join(path, entry.Name())
				if hasValidExtension(filePath, extensions) {
					files = append(files, filePath)
				}
			}
		}
	}

	return files, err
}

func hasValidExtension(filename string, extensions []string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	for _, validExt := range extensions {
		if ext == validExt {
			return true
		}
	}
	return false
} 