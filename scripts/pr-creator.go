package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/go-github/v56/github"
	"golang.org/x/oauth2"
)

type PRCreator struct {
	client    *github.Client
	repoOwner string
	repoName  string
}

func NewPRCreator(token, repoOwner, repoName string) *PRCreator {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &PRCreator{
		client:    client,
		repoOwner: repoOwner,
		repoName:  repoName,
	}
}

func (pc *PRCreator) CreateTestPR(ctx context.Context, originalFile, testFileName, testContent, branchName string, coverage float64) error {
	// Get the main branch ref
	mainRef, _, err := pc.client.Git.GetRef(ctx, pc.repoOwner, pc.repoName, "refs/heads/main")
	if err != nil {
		return fmt.Errorf("failed to get main branch ref: %v", err)
	}

	// Check if branch already exists and delete it if it does
	_, resp, err := pc.client.Git.GetRef(ctx, pc.repoOwner, pc.repoName, "refs/heads/"+branchName)
	if err == nil {
		// Branch exists, delete it
		_, err = pc.client.Git.DeleteRef(ctx, pc.repoOwner, pc.repoName, "refs/heads/"+branchName)
		if err != nil {
			return fmt.Errorf("failed to delete existing branch: %v", err)
		}
	} else if resp.StatusCode != 404 {
		// Some other error occurred
		return fmt.Errorf("failed to check branch existence: %v", err)
	}

	// Create new branch
	newRef := &github.Reference{
		Ref: github.String("refs/heads/" + branchName),
		Object: &github.GitObject{
			SHA: mainRef.Object.SHA,
		},
	}

	_, _, err = pc.client.Git.CreateRef(ctx, pc.repoOwner, pc.repoName, newRef)
	if err != nil {
		return fmt.Errorf("failed to create branch: %v", err)
	}

	// Create or update the test file
	err = pc.createOrUpdateFile(ctx, testFileName, testContent, branchName)
	if err != nil {
		return fmt.Errorf("failed to create test file: %v", err)
	}

	// Create pull request
	title := fmt.Sprintf("🧪 Auto-generated tests for %s", originalFile)
	body := pc.buildPRDescription(originalFile, coverage)
	
	pr := &github.NewPullRequest{
		Title: github.String(title),
		Head:  github.String(branchName),
		Base:  github.String("main"),
		Body:  github.String(body),
	}

	_, _, err = pc.client.PullRequests.Create(ctx, pc.repoOwner, pc.repoName, pr)
	if err != nil {
		return fmt.Errorf("failed to create pull request: %v", err)
	}

	return nil
}

func (pc *PRCreator) createOrUpdateFile(ctx context.Context, filePath, content, branchName string) error {
	// Check if file already exists
	existingFile, _, resp, err := pc.client.Repositories.GetContents(ctx, pc.repoOwner, pc.repoName, filePath, &github.RepositoryContentGetOptions{
		Ref: branchName,
	})

	// Encode content
	encodedContent := base64.StdEncoding.EncodeToString([]byte(content))

	if err != nil && resp.StatusCode == 404 {
		// File doesn't exist, create it
		fileOptions := &github.RepositoryContentFileOptions{
			Message: github.String(fmt.Sprintf("Add auto-generated tests for %s", filePath)),
			Content: []byte(encodedContent),
			Branch:  github.String(branchName),
		}

		_, _, err = pc.client.Repositories.CreateFile(ctx, pc.repoOwner, pc.repoName, filePath, fileOptions)
		if err != nil {
			return fmt.Errorf("failed to create file: %v", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check file existence: %v", err)
	} else {
		// File exists, update it
		fileOptions := &github.RepositoryContentFileOptions{
			Message: github.String(fmt.Sprintf("Update auto-generated tests for %s", filePath)),
			Content: []byte(encodedContent),
			Branch:  github.String(branchName),
			SHA:     existingFile.SHA,
		}

		_, _, err = pc.client.Repositories.UpdateFile(ctx, pc.repoOwner, pc.repoName, filePath, fileOptions)
		if err != nil {
			return fmt.Errorf("failed to update file: %v", err)
		}
	}

	return nil
}

func (pc *PRCreator) buildPRDescription(originalFile string, coverage float64) string {
	var body strings.Builder
	
	body.WriteString("## 🤖 Auto-Generated Unit Tests\n\n")
	body.WriteString(fmt.Sprintf("This PR contains automatically generated unit tests for `%s`.\n\n", originalFile))
	body.WriteString("### 📊 Coverage Information\n")
	body.WriteString(fmt.Sprintf("- **Original Coverage**: %.2f%%\n", coverage))
	body.WriteString("- **Coverage Threshold**: 40.00%\n")
	body.WriteString("- **Status**: ⚠️ Below threshold, tests generated\n\n")
	
	body.WriteString("### 🧪 Generated Tests Include\n")
	body.WriteString("- Basic functionality tests\n")
	body.WriteString("- Edge case handling\n")
	body.WriteString("- Error condition testing\n")
	body.WriteString("- Input validation tests\n\n")
	
	body.WriteString("### ✅ Review Checklist\n")
	body.WriteString("- [ ] Tests cover the main functionality\n")
	body.WriteString("- [ ] Tests include proper error handling\n")
	body.WriteString("- [ ] Test names are descriptive\n")
	body.WriteString("- [ ] Tests are independent and repeatable\n")
	body.WriteString("- [ ] No hardcoded values in tests\n\n")
	
	body.WriteString("### 🔧 Next Steps\n")
	body.WriteString("1. Review the generated tests\n")
	body.WriteString("2. Run `go test` to ensure all tests pass\n")
	body.WriteString("3. Modify or add additional tests if needed\n")
	body.WriteString("4. Merge when tests are satisfactory\n\n")
	
	body.WriteString("---\n")
	body.WriteString("*This PR was automatically created by the Auto Test Generator workflow.*")
	
	return body.String()
}

func (pc *PRCreator) CommentOnPR(ctx context.Context, prNumber, message string) error {
	if prNumber == "" {
		// No PR to comment on, skip
		return nil
	}

	prNum, err := strconv.Atoi(prNumber)
	if err != nil {
		return fmt.Errorf("invalid PR number: %v", err)
	}

	comment := &github.IssueComment{
		Body: github.String(message),
	}

	_, _, err = pc.client.Issues.CreateComment(ctx, pc.repoOwner, pc.repoName, prNum, comment)
	if err != nil {
		return fmt.Errorf("failed to create comment: %v", err)
	}

	return nil
}