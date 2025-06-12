package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "edgerag",
	Short: "EdgeRAG - Offline RAG system with local embeddings and LLM",
	Long: `EdgeRAG is a command-line tool for building and querying a Retrieval-Augmented Generation (RAG) system
that works completely offline. It uses sentence-transformers for embeddings and Ollama for LLM inference.

Features:
- Index documents from files or directories
- Generate embeddings using sentence-transformers
- Store vectors in memory for fast retrieval
- Query using natural language with Ollama LLM
- Completely offline operation`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.edgerag.yaml)")
	rootCmd.PersistentFlags().String("model", "paraphrase-MiniLM-L3-v2", "sentence-transformer model to use for embeddings")
	rootCmd.PersistentFlags().String("ollama-model", "llama3.2", "Ollama model to use for LLM inference")
	rootCmd.PersistentFlags().String("ollama-url", "http://localhost:11434", "Ollama server URL")

	viper.BindPFlag("model", rootCmd.PersistentFlags().Lookup("model"))
	viper.BindPFlag("ollama_model", rootCmd.PersistentFlags().Lookup("ollama-model"))
	viper.BindPFlag("ollama_url", rootCmd.PersistentFlags().Lookup("ollama-url"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".edgerag")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
} 