#!/usr/bin/env python3
"""
Embedding service using sentence-transformers.
This script runs as a persistent server, reading JSON requests from stdin
and outputting embeddings as JSON responses.
"""

import json
import sys
import numpy as np
from sentence_transformers import SentenceTransformer
import warnings
import gc
import torch

# Suppress warnings for cleaner output
warnings.filterwarnings("ignore")

# Global model cache to avoid reloading
_model_cache = {}

def load_model(model_name):
    """Load a sentence transformer model with caching."""
    try:
        if model_name in _model_cache:
            return _model_cache[model_name]
        
        # Load model with CPU-only to save memory
        model = SentenceTransformer(model_name, device='cpu')
        
        # Enable CPU optimizations if available
        if hasattr(torch, 'set_num_threads'):
            torch.set_num_threads(1)  # Use single thread to reduce memory
        
        _model_cache[model_name] = model
        return model
    except Exception as e:
        return None, str(e)

def generate_embedding(model, text):
    """Generate embedding for the given text with memory optimization."""
    try:
        # Generate embedding with reduced precision
        embedding = model.encode(
            text, 
            convert_to_tensor=False,
            show_progress_bar=False,
            batch_size=1  # Process one at a time to save memory
        )
        
        # Convert to float32 for consistency and memory efficiency
        if isinstance(embedding, np.ndarray):
            embedding = embedding.astype(np.float32)
        
        # Force garbage collection
        gc.collect()
        
        return embedding.tolist()
    except Exception as e:
        return None, str(e)

def handle_request(request_data):
    """Handle a single embedding request."""
    try:
        # Parse JSON input
        try:
            request = json.loads(request_data)
        except json.JSONDecodeError as e:
            return {"error": f"Invalid JSON input: {str(e)}"}
        
        # Validate input
        if "text" not in request:
            return {"error": "Missing 'text' field in request"}
        
        if "model" not in request:
            return {"error": "Missing 'model' field in request"}
        
        text = request["text"]
        model_name = request["model"]
        
        # Load model
        model = load_model(model_name)
        if model is None:
            return {"error": f"Failed to load model: {model_name}"}
        
        # Generate embedding
        embedding = generate_embedding(model, text)
        if embedding is None:
            return {"error": "Failed to generate embedding"}
        
        # Return result
        return {"embedding": embedding}
        
    except Exception as e:
        return {"error": f"Unexpected error: {str(e)}"}

def main():
    """Main server loop - process requests continuously."""
    try:
        # Signal that we're ready
        print(json.dumps({"status": "ready"}), flush=True)
        
        # Process requests line by line
        for line in sys.stdin:
            line = line.strip()
            if not line:
                continue
                
            if line == "QUIT":
                break
                
            # Handle the request
            response = handle_request(line)
            print(json.dumps(response), flush=True)
            
    except KeyboardInterrupt:
        pass
    except Exception as e:
        print(json.dumps({"error": f"Server error: {str(e)}"}), flush=True)
    finally:
        # Clean up memory
        gc.collect()

if __name__ == "__main__":
    main() 