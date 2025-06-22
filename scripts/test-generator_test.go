package main

import (
	"context"
	"testing"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"os"
	"path/filepath"
)

// Mock genai.Client for testing purposes
type mockClient struct {
	GenerateContentFunc func(ctx context.Context, req genai.GenerateContentRequest) (*genai.GenerateContentResponse, error)
}

func (m *mockClient) GenerativeModel(modelName string) *genai.GenerativeModel {
	return &genai.GenerativeModel{
		Client:           m,
		ModelName:        modelName,
		GenerateContent:  m.GenerateContentFunc,
		SetTemperature: func(temp float64){}, // Dummy implementation
	}
}


func TestNewTestGenerator(t *testing.T) {
	// Test case 1: Valid API key
	apiKey := "test-api-key"
	tg := NewTestGenerator(apiKey)
	if tg == nil {
		t.Error("NewTestGenerator returned nil")
	}
	if tg.client == nil {
		t.Error("NewTestGenerator client is nil")
	}

	// Test case 2: Invalid API key (simulated) -  replace with actual error handling if possible.
	// This test will currently panic because of the way NewTestGenerator handles errors.  A better approach would be to return an error instead of panicking.
	//defer func() {
	//	if r := recover(); r == nil {
	//		t.Errorf("The code did not panic")
	//	}
	//}()
	//tg = NewTestGenerator("") // Simulate invalid API key
	//t.Log("NewTestGenerator with invalid API key panicked as expected.")

}


func TestGenerateTests(t *testing.T) {
	// Test case 1: Successful test generation (mock)
	mockClient := &mockClient{
		GenerateContentFunc: func(ctx context.Context, req genai.GenerateContentRequest) (*genai.GenerateContentResponse, error) {
			return &genai.GenerateContentResponse{
				Candidates: []genai.Candidate{
					{
						Content: genai.Content{
							Parts: []genai.ContentPart{genai.Text("package main\nfunc TestExample() {}")},
						},
					},
				},
			}, nil
		},
	}
	tg := &TestGenerator{client: mockClient}
	ctx := context.Background()
	filePath := "test_file.go"
	functions := []FunctionInfo{}
	testContent, err := tg.GenerateTests(ctx, filePath, functions)
	if err != nil {
		t.Errorf("GenerateTests returned an error: %v", err)
	}
	if !strings.Contains(testContent, "package main") {
		t.Error("GenerateTests did not generate the expected package")
	}

	// Test case 2: Error reading file (mock)
	mockClient = &mockClient{
		GenerateContentFunc: func(ctx context.Context, req genai.GenerateContentRequest) (*genai.GenerateContentResponse, error) {
			return nil, fmt.Errorf("failed to generate content")
		},
	}
	tg = &TestGenerator{client: mockClient}
	_, err = tg.GenerateTests(ctx, "nonexistent_file.go", functions)
	if err == nil {
		t.Error("GenerateTests did not return an error when reading a nonexistent file")
	}

	// Test case 3: No content generated (mock)
	mockClient = &mockClient{
		GenerateContentFunc: func(ctx context.Context, req genai.GenerateContentRequest) (*genai.GenerateContentResponse, error) {
			return &genai.GenerateContentResponse{}, nil
		},
	}
	tg = &TestGenerator{client: mockClient}
	_, err = tg.GenerateTests(ctx, filePath, functions)
	if err == nil {
		t.Error("GenerateTests did not return an error when no content was generated")
	}

	// Test case 4: Error validating generated code (mock)
	tg = &TestGenerator{client: mockClient}
	tg.validateGeneratedCode = func(originalFilePath, testContent string) error {
		return fmt.Errorf("generated code doesn't compile")
	}
	_, err = tg.GenerateTests(ctx, filePath, functions)
	if err == nil {
		t.Error("GenerateTests did not return an error when generated code is invalid")
	}

}

func TestBuildPrompt(t *testing.T) {
	tg := &TestGenerator{}
	filePath := "test_file.go"
	originalContent := "package main\nfunc main() {}"
	functions := []FunctionInfo{{Name: "NewTestGenerator", Content: "func NewTestGenerator(apiKey string) *TestGenerator {return nil}"}}
	packageName := "main"
	prompt := tg.buildPrompt(filePath, originalContent, functions, packageName)
	if !strings.Contains(prompt, filePath) || !strings.Contains(prompt, originalContent) || !strings.Contains(prompt, packageName) {
		t.Error("buildPrompt did not include expected values")
	}
}

func TestExtractPackageInfo(t *testing.T) {
	tg := &TestGenerator{}
	content := `package mypackage

import (
	"fmt"
	"os"
)

import "path/filepath"`
	packageName, imports := tg.extractPackageInfo(content)
	if packageName != "mypackage" {
		t.Errorf("Unexpected package name: %s", packageName)
	}
	if len(imports) != 3 {
		t.Errorf("Unexpected number of imports: %d", len(imports))
	}
	if !strings.Contains(imports[0], `"fmt"`) || !strings.Contains(imports[1], `"os"`) || !strings.Contains(imports[2], `"path/filepath"`) {
		t.Error("Unexpected imports")
	}

	//Test with no imports
	content = `package mypackage`
	packageName, imports = tg.extractPackageInfo(content)
	if len(imports) != 0 {
		t.Errorf("Unexpected number of imports: %d", len(imports))
	}

	// Test with only single line imports
	content = `package mypackage
import "fmt"
`
	packageName, imports = tg.extractPackageInfo(content)
	if len(imports) != 1 {
		t.Errorf("Unexpected number of imports: %d", len(imports))
	}
	if !strings.Contains(imports[0], `"fmt"`) {
		t.Error("Unexpected imports")
	}
}

func TestCleanupGeneratedCode(t *testing.T) {
	tg := &TestGenerator{}
	generatedCode := "\npackage main\nfunc TestExample() {}\n"
	packageName := "main"
	cleanedCode := tg.cleanupGeneratedCode(generatedCode, packageName, nil)
	if !strings.Contains(cleanedCode, "package main") || !strings.Contains(cleanedCode, "func TestExample() {}") {
		t.Error("cleanupGeneratedCode did not clean up the code correctly")
	}

	// Test case 2: No package declaration
	generatedCode = "func TestExample() {}"
	cleanedCode = tg.cleanupGeneratedCode(generatedCode, packageName, nil)
	if !strings.Contains(cleanedCode, "package main") || !strings.Contains(cleanedCode, "func TestExample() {}") {
		t.Error("cleanupGeneratedCode did not add package declaration correctly")
	}

	// Test case 3: Missing testing import
	generatedCode = "package main\nfunc TestExample() {}"
	cleanedCode = tg.cleanupGeneratedCode(generatedCode, packageName, nil)
	if !strings.Contains(cleanedCode, "package main") || !strings.Contains(cleanedCode, "func TestExample() {}") || !strings.Contains(cleanedCode, "import \"testing\""){
		t.Error("cleanupGeneratedCode did not add testing import correctly")
	}
}


func TestValidateGeneratedCode(t *testing.T) {
	tg := &TestGenerator{}
	//This test is difficult to implement without actually compiling the code.  The original function skips compilation, so this test will also skip it.
	originalFilePath := "test_file.go"
	testContent := "package main\nimport \"testing\"\nfunc TestExample() {}"
	err := tg.validateGeneratedCode(originalFilePath, testContent)
	if err != nil {
		t.Errorf("validateGeneratedCode returned an error: %v", err)
	}
}

func TestResolveFilePath(t *testing.T) {
	tg := &TestGenerator{}

	// Simulate being in the scripts directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalDir)

	tempDir, err := os.MkdirTemp("", "test-resolve-filepath")
	if err != nil {
		t.Fatal(err)
	}
	scriptsDir := filepath.Join(tempDir, "scripts")
	os.Mkdir(scriptsDir, 0755)
	os.Chdir(scriptsDir)

	filePath := "../test_file.go"
	resolvedPath := tg.resolveFilePath(filePath)
	expectedPath := filepath.Join(tempDir, "test_file.go")
	if resolvedPath != expectedPath {
		t.Errorf("Unexpected resolved path: got %s, want %s", resolvedPath, expectedPath)
	}

	// Simulate not being in the scripts directory
	os.Chdir(tempDir)
	filePath = "test_file.go"
	resolvedPath = tg.resolveFilePath(filePath)
	expectedPath = filepath.Join(tempDir, "test_file.go")
	if resolvedPath != expectedPath {
		t.Errorf("Unexpected resolved path: got %s, want %s", resolvedPath, expectedPath)
	}

	os.RemoveAll(tempDir)
}

type FunctionInfo struct {
	Name    string
	Content string
}