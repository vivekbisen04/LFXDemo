package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"strings"
)

type Config struct {
	PRNumber      string
	ChangedFiles  string
	RepoOwner     string
	RepoName      string
	GithubToken   string
	GeminiAPIKey  string
	CoverageThreshold float64
}

func main() {
	config := parseFlags()
	
	if config.ChangedFiles == "" {
		log.Println("No changed files to process")
		return
	}

	ctx := context.Background()
	
	// Initialize services
	coverageAnalyzer := NewCoverageAnalyzer()
	testGenerator := NewTestGenerator(config.GeminiAPIKey)
	prCreator := NewPRCreator(config.GithubToken, config.RepoOwner, config.RepoName)

	// Process each changed file
	changedFiles := strings.Split(config.ChangedFiles, "\n")
	for _, file := range changedFiles {
		file = strings.TrimSpace(file)
		if file == "" {
			continue
		}

		log.Printf("Processing file: %s", file)
		
		// Check if file needs tests
		needsTests, coverage, err := coverageAnalyzer.AnalyzeFile(ctx, file, config.CoverageThreshold)
		if err != nil {
			log.Printf("Error analyzing coverage for %s: %v", file, err)
			prCreator.CommentOnPR(ctx, config.PRNumber, fmt.Sprintf("❌ Failed to analyze coverage for `%s`: %v", file, err))
			continue
		}

		if !needsTests {
			log.Printf("File %s has sufficient coverage (%.2f%%), skipping", file, coverage)
			continue
		}

		log.Printf("File %s needs tests (coverage: %.2f%%)", file, coverage)

		// Extract functions that need testing
		functions, err := coverageAnalyzer.ExtractModifiedFunctions(ctx, file)
		if err != nil {
			log.Printf("Error extracting functions from %s: %v", file, err)
			prCreator.CommentOnPR(ctx, config.PRNumber, fmt.Sprintf("❌ Failed to extract functions from `%s`: %v", file, err))
			continue
		}

		if len(functions) == 0 {
			log.Printf("No functions found in %s that need testing", file)
			continue
		}

		// Generate tests using LLM
		testContent, err := testGenerator.GenerateTests(ctx, file, functions)
		if err != nil {
			log.Printf("Error generating tests for %s: %v", file, err)
			prCreator.CommentOnPR(ctx, config.PRNumber, fmt.Sprintf("❌ Failed to generate tests for `%s`: %v", file, err))
			continue
		}

		// Create PR with generated tests
		testFileName := strings.TrimSuffix(file, ".go") + "_test.go"
		branchName := createBranchName(file)
		
		err = prCreator.CreateTestPR(ctx, file, testFileName, testContent, branchName, coverage)
		if err != nil {
			log.Printf("Error creating PR for %s: %v", file, err)
			prCreator.CommentOnPR(ctx, config.PRNumber, fmt.Sprintf("❌ Failed to create PR for tests of `%s`: %v", file, err))
			continue
		}

		log.Printf("Successfully created test PR for %s", file)
		prCreator.CommentOnPR(ctx, config.PRNumber, fmt.Sprintf("✅ Generated unit tests for `%s` (coverage was %.2f%%). New PR created with branch `%s`", file, coverage, branchName))
	}

	log.Println("Test generation process completed")
}

func parseFlags() *Config {
	config := &Config{}
	
	flag.StringVar(&config.PRNumber, "pr-number", "", "PR number that was merged")
	flag.StringVar(&config.ChangedFiles, "changed-files", "", "Newline-separated list of changed files")
	flag.StringVar(&config.RepoOwner, "repo-owner", "", "Repository owner")
	flag.StringVar(&config.RepoName, "repo-name", "", "Repository name")
	flag.StringVar(&config.GithubToken, "github-token", "", "GitHub token")
	flag.StringVar(&config.GeminiAPIKey, "gemini-api-key", "", "Gemini API key")
	flag.Float64Var(&config.CoverageThreshold, "coverage-threshold", 40.0, "Coverage threshold percentage")
	
	flag.Parse()

	// Validate required flags
	if config.RepoOwner == "" || config.RepoName == "" || config.GithubToken == "" || config.GeminiAPIKey == "" {
		log.Fatal("Missing required flags")
	}

	return config
}

func createBranchName(filePath string) string {
	dir := filepath.Dir(filePath)
	if dir == "." {
		return "auto-tests-root"
	}
	
	// Replace path separators and special characters
	branchName := strings.ReplaceAll(dir, "/", "-")
	branchName = strings.ReplaceAll(branchName, "_", "-")
	branchName = strings.ToLower(branchName)
	
	return "auto-tests-" + branchName
}