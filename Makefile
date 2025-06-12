.PHONY: build clean install test setup-python run-example install-global

# Build the CLI binary
build:
	go mod tidy
	go build -o edgerag .

# Clean build artifacts
clean:
	go clean
	rm -f edgerag

# Install dependencies
install:
	go mod tidy
	pip3 install -r requirements.txt

# Setup Python environment
setup-python:
	pip3 install -r requirements.txt

# Run tests
test:
	go test ./...

# Make Python script executable
setup-scripts:
	chmod +x scripts/embeddings.py

# Build and setup everything
setup: install setup-scripts build

# Run example (make sure Ollama is running first)
run-example:
	@echo "=== EdgeRAG Example ==="
	@echo "1. Creating sample documents..."
	@mkdir -p example_docs
	@echo "# Go Programming\n\nGo is a programming language developed by Google. It's known for its simplicity and performance.\n\n## Key Features\n- Fast compilation\n- Garbage collection\n- Strong typing\n- Concurrency support" > example_docs/go_intro.md
	@echo "# Python Programming\n\nPython is a high-level programming language known for its readability and versatility.\n\n## Key Features\n- Easy to learn\n- Large ecosystem\n- Dynamic typing\n- Great for data science" > example_docs/python_intro.md
	@echo "Sample documents created in example_docs/"
	@echo ""
	@echo "2. Indexing documents..."
	./edgerag index example_docs --recursive
	@echo ""
	@echo "3. Querying documents..."
	./edgerag query "What are the key features of Go?"
	@echo ""
	@echo "Example complete! Try your own queries:"
	@echo "./edgerag query \"Tell me about programming languages\""

# Install globally (requires sudo)
install-global: build
	@echo "Installing edgerag to /usr/local/bin (requires sudo)..."
	sudo cp edgerag /usr/local/bin/
	@echo "EdgeRAG installed globally! You can now use 'edgerag' from anywhere."

# Help
help:
	@echo "EdgeRAG Makefile Commands:"
	@echo "  build        - Build the CLI binary"
	@echo "  clean        - Clean build artifacts"
	@echo "  install      - Install all dependencies"
	@echo "  setup-python - Install Python dependencies"
	@echo "  test         - Run tests"
	@echo "  setup        - Full setup (install + build)"
	@echo "  run-example  - Run complete example"
	@echo "  install-global - Install edgerag system-wide"
	@echo "  help         - Show this help" 