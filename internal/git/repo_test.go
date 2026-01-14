package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestPathExists(t *testing.T) {
	// Create temp dir
	tmpDir, err := os.MkdirTemp("", "git-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Should exist
	if !PathExists(tmpDir) {
		t.Error("PathExists should return true for existing directory")
	}

	// Should not exist
	if PathExists(filepath.Join(tmpDir, "nonexistent")) {
		t.Error("PathExists should return false for nonexistent path")
	}
}

func TestIsGitRepo(t *testing.T) {
	// Create temp dir
	tmpDir, err := os.MkdirTemp("", "git-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Should not be a git repo initially
	if IsGitRepo(tmpDir) {
		t.Error("IsGitRepo should return false for non-git directory")
	}

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Skip("git not available")
	}

	// Should be a git repo now
	if !IsGitRepo(tmpDir) {
		t.Error("IsGitRepo should return true for git repository")
	}

	// Nonexistent path
	if IsGitRepo(filepath.Join(tmpDir, "nonexistent")) {
		t.Error("IsGitRepo should return false for nonexistent path")
	}
}

func TestGetRepoRoot(t *testing.T) {
	// Create temp dir and init git
	tmpDir, err := os.MkdirTemp("", "git-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Resolve symlinks (macOS /var -> /private/var)
	tmpDir, err = filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Skip("git not available")
	}

	// Create subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Get root from subdirectory
	root, err := GetRepoRoot(subDir)
	if err != nil {
		t.Fatalf("GetRepoRoot error: %v", err)
	}

	// Should return the parent (root) directory
	if root != tmpDir {
		t.Errorf("GetRepoRoot = %q, want %q", root, tmpDir)
	}

	// Non-git directory should error
	nonGitDir, _ := os.MkdirTemp("", "non-git")
	defer os.RemoveAll(nonGitDir)

	_, err = GetRepoRoot(nonGitDir)
	if err == nil {
		t.Error("GetRepoRoot should error for non-git directory")
	}
}

func TestGetCurrentBranch(t *testing.T) {
	// Create temp dir and init git
	tmpDir, err := os.MkdirTemp("", "git-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Init git
	cmd := exec.Command("git", "init", "-b", "main")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Skip("git not available or doesn't support -b flag")
	}

	// Configure git
	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = tmpDir
	cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Test")
	cmd.Dir = tmpDir
	cmd.Run()

	// Create initial commit
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	cmd.Run()
	cmd = exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = tmpDir
	cmd.Run()

	// Get current branch
	branch, err := GetCurrentBranch(tmpDir)
	if err != nil {
		t.Fatalf("GetCurrentBranch error: %v", err)
	}

	if branch != "main" {
		t.Errorf("GetCurrentBranch = %q, want %q", branch, "main")
	}
}

func TestBranchExists(t *testing.T) {
	// Create temp dir and init git
	tmpDir, err := os.MkdirTemp("", "git-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Init and create commit
	cmd := exec.Command("git", "init", "-b", "main")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Skip("git not available")
	}

	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = tmpDir
	cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Test")
	cmd.Dir = tmpDir
	cmd.Run()

	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	cmd.Run()
	cmd = exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = tmpDir
	cmd.Run()

	// Main should exist
	if !BranchExists(tmpDir, "main") {
		t.Error("BranchExists should return true for main branch")
	}

	// Nonexistent branch
	if BranchExists(tmpDir, "nonexistent-branch") {
		t.Error("BranchExists should return false for nonexistent branch")
	}
}

func TestCreateBranch(t *testing.T) {
	// Create temp dir and init git
	tmpDir, err := os.MkdirTemp("", "git-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Init and create commit
	cmd := exec.Command("git", "init", "-b", "main")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Skip("git not available")
	}

	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = tmpDir
	cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Test")
	cmd.Dir = tmpDir
	cmd.Run()

	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	cmd.Run()
	cmd = exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = tmpDir
	cmd.Run()

	// Create new branch
	err = CreateBranch(tmpDir, "feature-branch", "main")
	if err != nil {
		t.Fatalf("CreateBranch error: %v", err)
	}

	// Should be on new branch
	branch, _ := GetCurrentBranch(tmpDir)
	if branch != "feature-branch" {
		t.Errorf("Current branch = %q, want %q", branch, "feature-branch")
	}
}

func TestCheckout(t *testing.T) {
	// Create temp dir and init git
	tmpDir, err := os.MkdirTemp("", "git-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Init and create commit
	cmd := exec.Command("git", "init", "-b", "main")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Skip("git not available")
	}

	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = tmpDir
	cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Test")
	cmd.Dir = tmpDir
	cmd.Run()

	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	cmd.Run()
	cmd = exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = tmpDir
	cmd.Run()

	// Create and checkout a new branch
	cmd = exec.Command("git", "branch", "other-branch")
	cmd.Dir = tmpDir
	cmd.Run()

	err = Checkout(tmpDir, "other-branch")
	if err != nil {
		t.Fatalf("Checkout error: %v", err)
	}

	branch, _ := GetCurrentBranch(tmpDir)
	if branch != "other-branch" {
		t.Errorf("Current branch = %q, want %q", branch, "other-branch")
	}

	// Checkout back to main
	err = Checkout(tmpDir, "main")
	if err != nil {
		t.Fatalf("Checkout back to main error: %v", err)
	}

	branch, _ = GetCurrentBranch(tmpDir)
	if branch != "main" {
		t.Errorf("Current branch = %q, want %q", branch, "main")
	}
}

func TestDeleteBranch(t *testing.T) {
	// Create temp dir and init git
	tmpDir, err := os.MkdirTemp("", "git-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Init and create commit
	cmd := exec.Command("git", "init", "-b", "main")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Skip("git not available")
	}

	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = tmpDir
	cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Test")
	cmd.Dir = tmpDir
	cmd.Run()

	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	cmd.Run()
	cmd = exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = tmpDir
	cmd.Run()

	// Create branch
	cmd = exec.Command("git", "branch", "to-delete")
	cmd.Dir = tmpDir
	cmd.Run()

	// Verify exists
	if !BranchExists(tmpDir, "to-delete") {
		t.Fatal("Branch should exist before delete")
	}

	// Delete
	err = DeleteBranch(tmpDir, "to-delete")
	if err != nil {
		t.Fatalf("DeleteBranch error: %v", err)
	}

	// Verify deleted
	if BranchExists(tmpDir, "to-delete") {
		t.Error("Branch should not exist after delete")
	}
}
