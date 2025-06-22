package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-github/v56/github"
	"golang.org/x/oauth2"
)

// MockClient mocks the github.Client for testing purposes.
type MockClient struct {
	MockGetRef          func(ctx context.Context, owner, repo, ref string) (*github.Reference, *github.Response, error)
	MockDeleteRef       func(ctx context.Context, owner, repo, ref string) (*github.Response, error)
	MockCreateRef       func(ctx context.Context, owner, repo *github.Reference) (*github.Reference, *github.Response, error)
	MockGetContents     func(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentGetOptions) (*github.RepositoryContent, *github.Response, error)
	MockCreateFile      func(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentFileOptions) (*github.RepositoryContent, *github.Response, error)
	MockUpdateFile      func(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentFileOptions) (*github.RepositoryContent, *github.Response, error)
	MockPullRequestsCreate func(ctx context.Context, owner, repo string, pr *github.NewPullRequest) (*github.PullRequest, *github.Response, error)
	MockIssuesCreateComment func(ctx context.Context, owner, repo string, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error)

}

// NewMockClient creates a new MockClient.
func NewMockClient() *MockClient {
	return &MockClient{}
}

func (m *MockClient) GitGetRef(ctx context.Context, owner, repo, ref string) (*github.Reference, *github.Response, error) {
	if m.MockGetRef != nil {
		return m.MockGetRef(ctx, owner, repo, ref)
	}
	return nil, nil, nil
}

func (m *MockClient) GitDeleteRef(ctx context.Context, owner, repo, ref string) (*github.Response, error) {
	if m.MockDeleteRef != nil {
		return m.MockDeleteRef(ctx, owner, repo, ref)
	}
	return nil, nil
}

func (m *MockClient) GitCreateRef(ctx context.Context, owner, repo string, ref *github.Reference) (*github.Reference, *github.Response, error) {
	if m.MockCreateRef != nil {
		return m.MockCreateRef(ctx, owner, repo, ref)
	}
	return nil, nil, nil
}

func (m *MockClient) RepositoriesGetContents(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentGetOptions) (*github.RepositoryContent, *github.Response, error) {
	if m.MockGetContents != nil {
		return m.MockGetContents(ctx, owner, repo, path, opts)
	}
	return nil, nil, nil
}

func (m *MockClient) RepositoriesCreateFile(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentFileOptions) (*github.RepositoryContent, *github.Response, error) {
	if m.MockCreateFile != nil {
		return m.MockCreateFile(ctx, owner, repo, path, opts)
	}
	return nil, nil, nil
}

func (m *MockClient) RepositoriesUpdateFile(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentFileOptions) (*github.RepositoryContent, *github.Response, error) {
	if m.MockUpdateFile != nil {
		return m.MockUpdateFile(ctx, owner, repo, path, opts)
	}
	return nil, nil, nil
}

func (m *MockClient) PullRequestsCreate(ctx context.Context, owner, repo string, pr *github.NewPullRequest) (*github.PullRequest, *github.Response, error) {
	if m.MockPullRequestsCreate != nil {
		return m.MockPullRequestsCreate(ctx, owner, repo, pr)
	}
	return nil, nil, nil
}

func (m *MockClient) IssuesCreateComment(ctx context.Context, owner, repo string, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error) {
	if m.MockIssuesCreateComment != nil {
		return m.MockIssuesCreateComment(ctx, owner, repo, number, comment)
	}
	return nil, nil, nil
}


// TestNewPRCreator tests the NewPRCreator function.
func TestNewPRCreator(t *testing.T) {
	token := "test-token"
	repoOwner := "test-owner"
	repoName := "test-repo"

	pc := NewPRCreator(token, repoOwner, repoName)

	if pc == nil {
		t.Error("NewPRCreator returned nil")
	}

	if pc.client == nil {
		t.Error("NewPRCreator returned a PRCreator with a nil client")
	}

	if pc.repoOwner != repoOwner {
		t.Errorf("NewPRCreator returned a PRCreator with incorrect repoOwner: got %s, want %s", pc.repoOwner, repoOwner)
	}

	if pc.repoName != repoName {
		t.Errorf("NewPRCreator returned a PRCreator with incorrect repoName: got %s, want %s", pc.repoName, repoName)
	}
}


// TestCreateTestPR tests the CreateTestPR function.
func TestCreateTestPR(t *testing.T) {
	ctx := context.Background()
	mockClient := NewMockClient()
	pc := &PRCreator{client: mockClient, repoOwner: "test-owner", repoName: "test-repo"}

	// Test successful PR creation
	mockClient.MockGetRef = func(ctx context.Context, owner, repo, ref string) (*github.Reference, *github.Response, error) {
		return &github.Reference{Object: &github.GitObject{SHA: github.String("test-sha")}}, &github.Response{StatusCode: http.StatusOK}, nil
	}
	mockClient.MockDeleteRef = func(ctx context.Context, owner, repo, ref string) (*github.Response, error) {
		return &github.Response{StatusCode: http.StatusOK}, nil
	}
	mockClient.MockCreateRef = func(ctx context.Context, owner, repo string, ref *github.Reference) (*github.Reference, *github.Response, error) {
		return ref, &github.Response{StatusCode: http.StatusOK}, nil
	}
	mockClient.MockCreateFile = func(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentFileOptions) (*github.RepositoryContent, *github.Response, error) {
		return &github.RepositoryContent{}, &github.Response{StatusCode: http.StatusOK}, nil
	}
	mockClient.MockPullRequestsCreate = func(ctx context.Context, owner, repo string, pr *github.NewPullRequest) (*github.PullRequest, *github.Response, error) {
		return &github.PullRequest{}, &github.Response{StatusCode: http.StatusOK}, nil
	}

	err := pc.CreateTestPR(ctx, "originalFile.go", "testFile.go", "testContent", "testBranch", 80.0)
	if err != nil {
		t.Errorf("CreateTestPR returned an error: %v", err)
	}


	// Test error cases -  Simulate various errors from Github API calls.  More comprehensive error testing would require mocking more specific error responses.
	testCases := []struct {
		name           string
		mockGetRefErr  error
		mockDeleteRefErr error
		mockCreateRefErr error
		mockCreateFileErr error
		mockCreatePRErr error
		expectedErrContains string
	}{
		{"GetRefError", fmt.Errorf("get ref failed"), nil, nil, nil, nil, "failed to get main branch ref"},
		{"DeleteRefError", nil, fmt.Errorf("delete ref failed"), nil, nil, nil, "failed to delete existing branch"},
		{"CreateRefError", nil, nil, fmt.Errorf("create ref failed"), nil, nil, "failed to create branch"},
		{"CreateFileError", nil, nil, nil, fmt.Errorf("create file failed"), nil, "failed to create test file"},
		{"CreatePRError", nil, nil, nil, nil, fmt.Errorf("create pr failed"), "failed to create pull request"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient.MockGetRef = func(ctx context.Context, owner, repo, ref string) (*github.Reference, *github.Response, error) {
				return nil, nil, tc.mockGetRefErr
			}
			mockClient.MockDeleteRef = func(ctx context.Context, owner, repo, ref string) (*github.Response, error) {
				return nil, tc.mockDeleteRefErr
			}
			mockClient.MockCreateRef = func(ctx context.Context, owner, repo string, ref *github.Reference) (*github.Reference, *github.Response, error) {
				return nil, nil, tc.mockCreateRefErr
			}
			mockClient.MockCreateFile = func(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentFileOptions) (*github.RepositoryContent, *github.Response, error) {
				return nil, nil, tc.mockCreateFileErr
			}
			mockClient.MockPullRequestsCreate = func(ctx context.Context, owner, repo string, pr *github.NewPullRequest) (*github.PullRequest, *github.Response, error) {
				return nil, nil, tc.mockCreatePRErr
			}

			err := pc.CreateTestPR(ctx, "originalFile.go", "testFile.go", "testContent", "testBranch", 80.0)
			if err == nil {
				t.Errorf("Expected error, but got nil")
			}
			if !strings.Contains(err.Error(), tc.expectedErrContains) {
				t.Errorf("Error message does not contain expected substring: got %s, want %s", err.Error(), tc.expectedErrContains)
			}
		})
	}
}

// TestCreateOrUpdateFile tests the createOrUpdateFile function.
func TestCreateOrUpdateFile(t *testing.T) {
	ctx := context.Background()
	mockClient := NewMockClient()
	pc := &PRCreator{client: mockClient, repoOwner: "test-owner", repoName: "test-repo"}

	// Test file creation
	mockClient.MockGetContents = func(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentGetOptions) (*github.RepositoryContent, *github.Response, error) {
		return nil, &github.Response{StatusCode: http.StatusNotFound}, fmt.Errorf("not found")
	}
	mockClient.MockCreateFile = func(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentFileOptions) (*github.RepositoryContent, *github.Response, error) {
		return &github.RepositoryContent{}, &github.Response{StatusCode: http.StatusOK}, nil
	}
	err := pc.createOrUpdateFile(ctx, "testFile.go", "testContent", "testBranch")
	if err != nil {
		t.Errorf("createOrUpdateFile (create) returned an error: %v", err)
	}

	// Test file update
	mockClient.MockGetContents = func(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentGetOptions) (*github.RepositoryContent, *github.Response, error) {
		return &github.RepositoryContent{SHA: github.String("test-sha")}, &github.Response{StatusCode: http.StatusOK}, nil
	}
	mockClient.MockUpdateFile = func(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentFileOptions) (*github.RepositoryContent, *github.Response, error) {
		return &github.RepositoryContent{}, &github.Response{StatusCode: http.StatusOK}, nil
	}
	err = pc.createOrUpdateFile(ctx, "testFile.go", "updatedContent", "testBranch")
	if err != nil {
		t.Errorf("createOrUpdateFile (update) returned an error: %v", err)
	}

	// Test error cases
	testCases := []struct {
		name              string
		mockGetContentsErr error
		mockCreateFileErr  error
		mockUpdateFileErr  error
		expectedErrContains string
	}{
		{"GetContentsError", fmt.Errorf("get contents failed"), nil, nil, "failed to check file existence"},
		{"CreateFileError", nil, fmt.Errorf("create file failed"), nil, "failed to create file"},
		{"UpdateFileError", nil, nil, fmt.Errorf("update file failed"), "failed to update file"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient.MockGetContents = func(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentGetOptions) (*github.RepositoryContent, *github.Response, error) {
				return nil, nil, tc.mockGetContentsErr
			}
			mockClient.MockCreateFile = func(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentFileOptions) (*github.RepositoryContent, *github.Response, error) {
				return nil, nil, tc.mockCreateFileErr
			}
			mockClient.MockUpdateFile = func(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentFileOptions) (*github.RepositoryContent, *github.Response, error) {
				return nil, nil, tc.mockUpdateFileErr
			}

			err := pc.createOrUpdateFile(ctx, "testFile.go", "testContent", "testBranch")
			if err == nil {
				t.Errorf("Expected error, but got nil")
			}
			if !strings.Contains(err.Error(), tc.expectedErrContains) {
				t.Errorf("Error message does not contain expected substring: got %s, want %s", err.Error(), tc.expectedErrContains)
			}
		})
	}
}

// TestBuildPRDescription tests the buildPRDescription function.
func TestBuildPRDescription(t *testing.T) {
	pc := &PRCreator{}
	description := pc.buildPRDescription("originalFile.go", 90.5)
	if !strings.Contains(description, "originalFile.go") {
		t.Errorf("buildPRDescription did not include original file name")
	}
	if !strings.Contains(description, "90.50%") {
		t.Errorf("buildPRDescription did not include coverage percentage")
	}

}

// TestCommentOnPR tests the CommentOnPR function.
func TestCommentOnPR(t *testing.T) {
	ctx := context.Background()
	mockClient := NewMockClient()
	pc := &PRCreator{client: mockClient, repoOwner: "test-owner", repoName: "test-repo"}

	// Test successful comment creation
	mockClient.MockIssuesCreateComment = func(ctx context.Context, owner, repo string, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error) {
		return &github.IssueComment{}, &github.Response{StatusCode: http.StatusOK}, nil
	}
	err := pc.CommentOnPR(ctx, "123", "test comment")
	if err != nil {
		t.Errorf("CommentOnPR returned an error: %v", err)
	}

	// Test empty PR number
	err = pc.CommentOnPR(ctx, "", "test comment")
	if err != nil {
		t.Errorf("CommentOnPR returned an error for empty PR number: %v", err)
	}

	// Test invalid PR number
	err = pc.CommentOnPR(ctx, "abc", "test comment")
	if err == nil {
		t.Error("CommentOnPR did not return an error for invalid PR number")
	}

	// Test error from Github API
	mockClient.MockIssuesCreateComment = func(ctx context.Context, owner, repo string, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error) {
		return nil, nil, fmt.Errorf("comment creation failed")
	}
	err = pc.CommentOnPR(ctx, "123", "test comment")
	if err == nil {
		t.Error("CommentOnPR did not return an error for Github API error")
	}
}