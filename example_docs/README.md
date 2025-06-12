# EdgeRAG - Offline RAG CLI Tool

EdgeRAG is a command-line tool for building and querying a Retrieval-Augmented Generation (RAG) system that works completely offline. It uses sentence-transformers for embeddings and Ollama for LLM inference.

## Features

- 🔍 **Index documents** from files or directories
- 🧠 **Generate embeddings** using sentence-transformers
- 💾 **Store vectors in memory** for fast retrieval
- 🤖 **Query using natural language** with Ollama LLM
- 🚫 **Completely offline operation**
- 📁 **Multiple file format support** (.txt, .md, .go, .py, .js, etc.)
- ⚡ **Fast similarity search** with cosine similarity
- 🎯 **Customizable chunking** strategies

## Prerequisites

### 1. Go 1.21+
Install Go from [https://golang.org/dl/](https://golang.org/dl/)

### 2. Python 3.8+ with sentence-transformers
```bash
pip install sentence-transformers
```

### 3. Ollama
Install and run Ollama from [https://ollama.ai/](https://ollama.ai/)

```bash
# Install a model (e.g., llama2)
ollama pull llama2

# Start Ollama server (usually runs on localhost:11434)
ollama serve
```

## Installation

1. Clone or download the project
2. Build the CLI tool:

```bash
cd edgerag
go mod tidy
go build -o edgerag .
```

## Quick Start

### 1. Index your documents

```bash
# Index a single file
./edgerag index document.txt

# Index a directory recursively
./edgerag index ./docs --recursive

# Index with custom settings
./edgerag index ./src --recursive --chunk-size 1024 --extensions .go,.py,.js
```

### 2. Query your documents

```bash
# Ask a question
./edgerag query "How do I initialize a Go module?"

# Get more results
./edgerag query "What are the main features?" --top-k 5

# Use different similarity threshold
./edgerag query "Explain the architecture" --threshold 0.6
```

## Usage

### Global Options

```bash
--model string           sentence-transformer model (default "all-MiniLM-L6-v2")
--ollama-model string    Ollama model name (default "llama2")  
--ollama-url string      Ollama server URL (default "http://localhost:11434")
--config string          config file (default is $HOME/.edgerag.yaml)
```

### Index Command

```bash
edgerag index [file/directory] [flags]

Flags:
  -r, --recursive              Recursively index directories
  -e, --extensions strings     File extensions to index (default [.txt,.md,.go,.py,.js])
  -c, --chunk-size int         Maximum chunk size for document splitting (default 512)
  -o, --chunk-overlap int      Overlap between chunks when splitting documents (default 50)
```

### Query Command

```bash
edgerag query [question] [flags]

Flags:
  -k, --top-k int              Number of most relevant chunks to retrieve (default 3)
  -t, --threshold float32      Similarity threshold for retrieval (default 0.7)
  -p, --prompt-template string Custom prompt template for LLM
  -s, --show-sources           Show source documents in the response (default true)
```

## Configuration

Create a config file at `$HOME/.edgerag.yaml`:

```yaml
model: "all-MiniLM-L6-v2"
ollama_model: "llama2"
ollama_url: "http://localhost:11434"
```

## Architecture

EdgeRAG consists of several key components:

1. **Document Processor**: Loads and chunks documents into manageable pieces
2. **Embedding Service**: Generates vector embeddings using Python sentence-transformers
3. **Vector Store**: In-memory storage with cosine similarity search
4. **Ollama Client**: Interfaces with Ollama for LLM inference
5. **RAG Pipeline**: Orchestrates retrieval and generation

## Supported File Types

- `.txt` - Plain text files
- `.md` - Markdown files  
- `.go` - Go source code
- `.py` - Python source code
- `.js` - JavaScript source code
- `.ts` - TypeScript source code
- `.java` - Java source code
- `.cpp`, `.c` - C/C++ source code
- `.rs` - Rust source code
- `.rb` - Ruby source code
- `.php` - PHP source code
- `.sh` - Shell scripts
- And more...

## Examples

### Index a Go project
```bash
./edgerag index ./myproject --recursive --extensions .go,.md
```

### Query with custom prompt
```bash
./edgerag query "Explain the main function" \
  --prompt-template "Based on the code context below, explain the main function:\n\nContext:\n{{.Context}}\n\nQuestion: {{.Question}}\n\nDetailed explanation:"
```

### Use different models
```bash
./edgerag index ./docs --model "all-mpnet-base-v2"
./edgerag query "What is this about?" --ollama-model "codellama"
```

## Performance Tips

1. **Chunk Size**: Smaller chunks (256-512) work better for specific questions, larger chunks (1024+) for broader context
2. **Model Selection**: 
   - `all-MiniLM-L6-v2`: Fast, good for general use
   - `all-mpnet-base-v2`: Better quality, slower
   - `all-distilroberta-v1`: Good balance
3. **Top-K**: Start with 3-5 results, increase if needed
4. **Threshold**: Lower values (0.5-0.7) for broader search, higher (0.8+) for precise matches

## Troubleshooting

### Python/sentence-transformers issues
```bash
# Install/upgrade sentence-transformers
pip install --upgrade sentence-transformers

# Test Python script directly
echo '{"text": "test", "model": "all-MiniLM-L6-v2"}' | python3 scripts/embeddings.py
```

### Ollama connection issues
```bash
# Check if Ollama is running
curl http://localhost:11434/api/tags

# Start Ollama
ollama serve

# List available models
ollama list
```

### Build issues
```bash
# Clean and rebuild
go mod tidy
go clean
go build -o edgerag .
```

## Development

To extend or modify EdgeRAG:

1. Fork the repository
2. Make changes to the relevant packages in `internal/`
3. Test with `go test ./...`
4. Build with `go build`

## License

MIT License - see LICENSE file for details.

## Contributing

Pull requests welcome! Please ensure:
- Code follows Go conventions
- Tests pass
- Documentation is updated
- Examples work as expected