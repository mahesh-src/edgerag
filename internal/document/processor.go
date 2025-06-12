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

// ChunkSemanticDocument splits a document using semantic boundaries
func ChunkSemanticDocument(doc *Document, maxChunkSize int, overlap int) []*Chunk {
	content := doc.Content
	if len(content) == 0 {
		return []*Chunk{}
	}

	var chunks []*Chunk
	chunkIndex := 0

	// Split by double newlines (paragraphs) first
	paragraphs := strings.Split(content, "\n\n")
	
	var currentChunk strings.Builder
	var chunkStart int = 0
	var globalOffset int = 0

	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if len(paragraph) == 0 {
			globalOffset += 2 // account for \n\n
			continue
		}

		// If adding this paragraph would exceed max size, finalize current chunk
		if currentChunk.Len() > 0 && currentChunk.Len()+len(paragraph)+2 > maxChunkSize {
			// Create chunk from current content
			chunkContent := strings.TrimSpace(currentChunk.String())
			if len(chunkContent) > 0 {
				chunk := createChunk(doc, chunkIndex, chunkContent, chunkStart, globalOffset-1)
				chunks = append(chunks, chunk)
				chunkIndex++
			}

			// Start new chunk with overlap
			currentChunk.Reset()
			if overlap > 0 && len(chunkContent) > overlap {
				overlapText := chunkContent[len(chunkContent)-overlap:]
				currentChunk.WriteString(overlapText)
				chunkStart = globalOffset - overlap
			} else {
				chunkStart = globalOffset
			}
		}

		// Add paragraph to current chunk
		if currentChunk.Len() > 0 {
			currentChunk.WriteString("\n\n")
		}
		currentChunk.WriteString(paragraph)
		
		globalOffset += len(paragraph) + 2 // account for \n\n after each paragraph
	}

	// Add final chunk if there's content
	if currentChunk.Len() > 0 {
		chunkContent := strings.TrimSpace(currentChunk.String())
		if len(chunkContent) > 0 {
			chunk := createChunk(doc, chunkIndex, chunkContent, chunkStart, globalOffset)
			chunks = append(chunks, chunk)
		}
	}

	return chunks
}

// ChunkSmartDocument uses intelligent splitting based on content structure
func ChunkSmartDocument(doc *Document, maxChunkSize int, overlap int) []*Chunk {
	content := doc.Content
	if len(content) == 0 {
		return []*Chunk{}
	}

	// For markdown files, split on headers first
	if isMarkdown(doc) {
		return chunkMarkdownSections(doc, maxChunkSize, overlap)
	}

	// For other files, use semantic paragraph splitting
	return ChunkSemanticDocument(doc, maxChunkSize, overlap)
}

// chunkMarkdownSections splits markdown content on header boundaries
func chunkMarkdownSections(doc *Document, maxChunkSize int, overlap int) []*Chunk {
	content := doc.Content
	lines := strings.Split(content, "\n")
	
	var chunks []*Chunk
	var currentSection strings.Builder
	var sectionStart int = 0
	chunkIndex := 0
	globalOffset := 0

	for _, line := range lines {
		lineLen := len(line) + 1 // +1 for newline
		
		// Check if this is a header (starts with #)
		isHeader := strings.HasPrefix(strings.TrimSpace(line), "#")
		
		// If we hit a header and have content, finalize current section
		if isHeader && currentSection.Len() > 0 {
			sectionContent := strings.TrimSpace(currentSection.String())
			if len(sectionContent) > 0 {
				// If section is too large, split it further
				if len(sectionContent) > maxChunkSize {
					subChunks := splitLargeSection(doc, sectionContent, maxChunkSize, overlap, sectionStart, &chunkIndex)
					chunks = append(chunks, subChunks...)
				} else {
					chunk := createChunk(doc, chunkIndex, sectionContent, sectionStart, globalOffset)
					chunks = append(chunks, chunk)
					chunkIndex++
				}
			}
			
			// Start new section
			currentSection.Reset()
			sectionStart = globalOffset
		}
		
		// Add line to current section
		if currentSection.Len() > 0 {
			currentSection.WriteString("\n")
		}
		currentSection.WriteString(line)
		globalOffset += lineLen
	}

	// Add final section
	if currentSection.Len() > 0 {
		sectionContent := strings.TrimSpace(currentSection.String())
		if len(sectionContent) > 0 {
			if len(sectionContent) > maxChunkSize {
				subChunks := splitLargeSection(doc, sectionContent, maxChunkSize, overlap, sectionStart, &chunkIndex)
				chunks = append(chunks, subChunks...)
			} else {
				chunk := createChunk(doc, chunkIndex, sectionContent, sectionStart, globalOffset)
				chunks = append(chunks, chunk)
			}
		}
	}

	return chunks
}

// splitLargeSection splits a large section into smaller semantic chunks
func splitLargeSection(doc *Document, content string, maxChunkSize int, overlap int, baseOffset int, chunkIndex *int) []*Chunk {
	// First try splitting by double newlines (paragraphs)
	paragraphs := strings.Split(content, "\n\n")
	var chunks []*Chunk
	var currentChunk strings.Builder
	localOffset := 0

	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if len(paragraph) == 0 {
			localOffset += 2
			continue
		}

		// If adding this paragraph exceeds size, finalize current chunk
		if currentChunk.Len() > 0 && currentChunk.Len()+len(paragraph)+2 > maxChunkSize {
			chunkContent := strings.TrimSpace(currentChunk.String())
			if len(chunkContent) > 0 {
				chunk := createChunk(doc, *chunkIndex, chunkContent, baseOffset+localOffset-len(chunkContent), baseOffset+localOffset)
				chunks = append(chunks, chunk)
				(*chunkIndex)++
			}
			currentChunk.Reset()
		}

		// Add paragraph
		if currentChunk.Len() > 0 {
			currentChunk.WriteString("\n\n")
		}
		currentChunk.WriteString(paragraph)
		localOffset += len(paragraph) + 2
	}

	// Add final chunk
	if currentChunk.Len() > 0 {
		chunkContent := strings.TrimSpace(currentChunk.String())
		if len(chunkContent) > 0 {
			chunk := createChunk(doc, *chunkIndex, chunkContent, baseOffset+localOffset-len(chunkContent), baseOffset+localOffset)
			chunks = append(chunks, chunk)
			(*chunkIndex)++
		}
	}

	return chunks
}

// createChunk helper function to create a chunk with metadata
func createChunk(doc *Document, chunkIndex int, content string, start, end int) *Chunk {
	chunkMetadata := make(map[string]interface{})
	for k, v := range doc.Metadata {
		chunkMetadata[k] = v
	}
	chunkMetadata["chunk_index"] = chunkIndex
	chunkMetadata["chunk_start"] = start
	chunkMetadata["chunk_end"] = end
	chunkMetadata["parent_id"] = doc.ID
	chunkMetadata["chunk_type"] = "semantic"

	return &Chunk{
		ID:       fmt.Sprintf("%s_semantic_%d", doc.ID, chunkIndex),
		Content:  content,
		Metadata: chunkMetadata,
	}
}

// isMarkdown checks if the document is a markdown file
func isMarkdown(doc *Document) bool {
	if ext, ok := doc.Metadata["extension"].(string); ok {
		return ext == ".md" || ext == ".markdown"
	}
	return false
} 