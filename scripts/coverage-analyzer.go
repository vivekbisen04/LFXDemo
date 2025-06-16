package main

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type CoverageAnalyzer struct {
	fileSet *token.FileSet
}

type FunctionInfo struct {
	Name     string
	Content  string
	StartLine int
	EndLine   int
}

func NewCoverageAnalyzer() *CoverageAnalyzer {
	return &CoverageAnalyzer{
		fileSet: token.NewFileSet(),
	}
}

// resolveFilePath converts relative paths to absolute paths from repo root
func (ca *CoverageAnalyzer) resolveFilePath(filePath string) string {
	// Check if we're running from scripts directory
	if wd, err := os.Getwd(); err == nil {
		if strings.HasSuffix(wd, "/scripts") || strings.HasSuffix(wd, "\\scripts") {
			// We're in scripts directory, so prepend ../ to access repo root files
			return filepath.Join("..", filePath)
		}
	}
	// Otherwise, use the path as-is
	return filePath
}

func (ca *CoverageAnalyzer) AnalyzeFile(ctx context.Context, filePath string, threshold float64) (needsTests bool, coverage float64, err error) {
	// Resolve the file path
	resolvedPath := ca.resolveFilePath(filePath)
	
	// Check if test file already exists
	testFile := strings.TrimSuffix(resolvedPath, ".go") + "_test.go"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		// No test file exists, definitely needs tests
		return true, 0.0, nil
	}

	// Run coverage analysis for the specific package
	packageDir := filepath.Dir(resolvedPath)
	if packageDir == "" {
		packageDir = "."
	}

	// Change to the package directory for running tests
	originalDir, err := os.Getwd()
	if err != nil {
		return false, 0.0, fmt.Errorf("failed to get current directory: %v", err)
	}

	// Convert package directory to absolute path
	absPackageDir, err := filepath.Abs(packageDir)
	if err != nil {
		return false, 0.0, fmt.Errorf("failed to get absolute path: %v", err)
	}

	// Change to package directory
	if err := os.Chdir(absPackageDir); err != nil {
		return false, 0.0, fmt.Errorf("failed to change to package directory: %v", err)
	}

	// Ensure we change back to original directory
	defer func() {
		os.Chdir(originalDir)
		// Clean up coverage file
		os.Remove("coverage.out")
	}()

	cmd := exec.CommandContext(ctx, "go", "test", "-cover", "-coverprofile=coverage.out", ".")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If tests fail to run, we might still want to generate tests
		// Check if it's because of compilation errors vs no tests
		if strings.Contains(string(output), "no test files") {
			return true, 0.0, nil
		}
		return false, 0.0, fmt.Errorf("failed to run coverage: %v, output: %s", err, string(output))
	}

	// Parse coverage output
	coverage, err = ca.parseCoverageOutput(string(output))
	if err != nil {
		return false, 0.0, fmt.Errorf("failed to parse coverage output: %v", err)
	}

	needsTests = coverage < threshold
	return needsTests, coverage, nil
}

func (ca *CoverageAnalyzer) parseCoverageOutput(output string) (float64, error) {
	// Look for coverage percentage in output
	// Format: "coverage: XX.X% of statements"
	re := regexp.MustCompile(`coverage: (\d+\.?\d*)% of statements`)
	matches := re.FindStringSubmatch(output)
	
	if len(matches) < 2 {
		// No coverage found, assume 0%
		return 0.0, nil
	}

	coverage, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0.0, fmt.Errorf("failed to parse coverage percentage: %v", err)
	}

	return coverage, nil
}

func (ca *CoverageAnalyzer) ExtractModifiedFunctions(ctx context.Context, filePath string) ([]FunctionInfo, error) {
	// Resolve the file path
	resolvedPath := ca.resolveFilePath(filePath)
	
	// Read the file
	content, err := os.ReadFile(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	// Parse the Go file
	node, err := parser.ParseFile(ca.fileSet, resolvedPath, content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %v", err)
	}

	var functions []FunctionInfo

	// Extract all functions
	ast.Inspect(node, func(n ast.Node) bool {
		switch fn := n.(type) {
		case *ast.FuncDecl:
			if fn.Name.IsExported() || ca.shouldIncludeFunction(fn) {
				startPos := ca.fileSet.Position(fn.Pos())
				endPos := ca.fileSet.Position(fn.End())
				
				// Extract function content
				lines := strings.Split(string(content), "\n")
				var funcContent strings.Builder
				
				for i := startPos.Line - 1; i < endPos.Line && i < len(lines); i++ {
					funcContent.WriteString(lines[i])
					funcContent.WriteString("\n")
				}

				functions = append(functions, FunctionInfo{
					Name:      fn.Name.Name,
					Content:   funcContent.String(),
					StartLine: startPos.Line,
					EndLine:   endPos.Line,
				})
			}
		}
		return true
	})

	return functions, nil
}

func (ca *CoverageAnalyzer) shouldIncludeFunction(fn *ast.FuncDecl) bool {
	// Include functions that are:
	// 1. Exported (public)
	// 2. Have significant logic (more than just getters/setters)
	// 3. Are not test functions
	
	if fn.Name == nil {
		return false
	}

	name := fn.Name.Name
	
	// Skip test functions
	if strings.HasPrefix(name, "Test") || strings.HasPrefix(name, "Benchmark") {
		return false
	}

	// Skip simple getters/setters (heuristic: less than 3 statements)
	if fn.Body != nil && len(fn.Body.List) < 3 {
		return false
	}

	return true
}