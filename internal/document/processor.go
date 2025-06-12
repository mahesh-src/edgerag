package document

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// Document represents a loaded document
type Document struct {
	ID       string                 `json:"id"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
}

// Chunk represents a chunk of a document
type Chunk struct {
	ID       string                 `json:"id"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
}

// LoadFromFile loads a document from a file
func LoadFromFile(filePath string) (*Document, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Validate UTF-8 encoding
	if !utf8.Valid(content) {
		return nil, fmt.Errorf("file %s contains invalid UTF-8", filePath)
	}

	// Generate document ID based on file path and content
	hasher := md5.New()
	hasher.Write([]byte(filePath))
	hasher.Write(content)
	docID := fmt.Sprintf("%x", hasher.Sum(nil))

	doc := &Document{
		ID:      docID,
		Content: string(content),
		Metadata: map[string]interface{}{
			"file":      filePath,
			"filename":  filepath.Base(filePath),
			"extension": filepath.Ext(filePath),
			"size":      len(content),
		},
	}

	return doc, nil
}

// LoadFromString creates a document from a string
func LoadFromString(content string, metadata map[string]interface{}) *Document {
	hasher := md5.New()
	hasher.Write([]byte(content))
	docID := fmt.Sprintf("%x", hasher.Sum(nil))

	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["size"] = len(content)

	return &Document{
		ID:       docID,
		Content:  content,
		Metadata: metadata,
	}
}

// ChunkDocument splits a document into smaller chunks
func ChunkDocument(doc *Document, chunkSize int, overlap int) []*Chunk {
	content := doc.Content
	if len(content) == 0 {
		return []*Chunk{}
	}

	var chunks []*Chunk
	start := 0
	chunkIndex := 0

	for start < len(content) {
		end := start + chunkSize
		if end > len(content) {
			end = len(content)
		}

		// Try to break at word boundaries
		if end < len(content) && content[end] != ' ' && content[end] != '\n' {
			// Look for the nearest space or newline before the end
			for i := end - 1; i > start && i > end-50; i-- {
				if content[i] == ' ' || content[i] == '\n' {
					end = i
					break
				}
			}
		}

		chunkContent := strings.TrimSpace(content[start:end])
		if len(chunkContent) == 0 {
			start = end + 1
			continue
		}

		// Create chunk metadata
		chunkMetadata := make(map[string]interface{})
		for k, v := range doc.Metadata {
			chunkMetadata[k] = v
		}
		chunkMetadata["chunk_index"] = chunkIndex
		chunkMetadata["chunk_start"] = start
		chunkMetadata["chunk_end"] = end
		chunkMetadata["parent_id"] = doc.ID

		chunk := &Chunk{
			ID:       fmt.Sprintf("%s_chunk_%d", doc.ID, chunkIndex),
			Content:  chunkContent,
			Metadata: chunkMetadata,
		}

		chunks = append(chunks, chunk)
		chunkIndex++

		// Move start position, accounting for overlap
		nextStart := end - overlap
		if nextStart <= start {
			// Prevent infinite loop - if we can't advance, we're done
			break
		}
		start = nextStart
	}

	return chunks
}

// ChunkByLines splits text into chunks by lines (useful for code files)
func ChunkByLines(doc *Document, maxLines int, overlap int) []*Chunk {
	lines := strings.Split(doc.Content, "\n")
	if len(lines) == 0 {
		return []*Chunk{}
	}

	var chunks []*Chunk
	start := 0
	chunkIndex := 0

	for start < len(lines) {
		end := start + maxLines
		if end > len(lines) {
			end = len(lines)
		}

		chunkLines := lines[start:end]
		chunkContent := strings.Join(chunkLines, "\n")
		chunkContent = strings.TrimSpace(chunkContent)

		if len(chunkContent) == 0 {
			start = end
			continue
		}

		// Create chunk metadata
		chunkMetadata := make(map[string]interface{})
		for k, v := range doc.Metadata {
			chunkMetadata[k] = v
		}
		chunkMetadata["chunk_index"] = chunkIndex
		chunkMetadata["line_start"] = start + 1 // 1-based line numbers
		chunkMetadata["line_end"] = end
		chunkMetadata["parent_id"] = doc.ID

		chunk := &Chunk{
			ID:       fmt.Sprintf("%s_lines_%d-%d", doc.ID, start+1, end),
			Content:  chunkContent,
			Metadata: chunkMetadata,
		}

		chunks = append(chunks, chunk)
		chunkIndex++

		// Move start position, accounting for overlap
		start = end - overlap
		if start < 0 {
			start = 0
		}
		if start >= end {
			break
		}
	}

	return chunks
}

// GetFileType determines the type of file based on extension
func GetFileType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".md", ".markdown":
		return "markdown"
	case ".txt":
		return "text"
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".js", ".jsx":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".java":
		return "java"
	case ".cpp", ".cc", ".cxx", ".c++":
		return "cpp"
	case ".c":
		return "c"
	case ".h", ".hpp":
		return "header"
	case ".rs":
		return "rust"
	case ".rb":
		return "ruby"
	case ".php":
		return "php"
	case ".sh", ".bash":
		return "shell"
	case ".sql":
		return "sql"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".xml":
		return "xml"
	case ".html", ".htm":
		return "html"
	case ".css":
		return "css"
	default:
		return "text"
	}
} 