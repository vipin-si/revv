package sandbox

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCreateTarArchive(t *testing.T) {
	tempDir := t.TempDir()

	// Create some files
	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(tempDir, "subdir", "file2.txt")
	gitFile := filepath.Join(tempDir, ".git", "config")
	agentFile := filepath.Join(tempDir, ".agents", "plan.md")
	nodeFile := filepath.Join(tempDir, "node_modules", "package.json")

	if err := os.MkdirAll(filepath.Dir(file2), 0755); err != nil {
		t.Fatalf("failed to mkdir: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(gitFile), 0755); err != nil {
		t.Fatalf("failed to mkdir: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(agentFile), 0755); err != nil {
		t.Fatalf("failed to mkdir: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(nodeFile), 0755); err != nil {
		t.Fatalf("failed to mkdir: %v", err)
	}

	if err := os.WriteFile(file1, []byte("content1"), 0644); err != nil {
		t.Fatalf("failed to write file1: %v", err)
	}
	if err := os.WriteFile(file2, []byte("content2"), 0644); err != nil {
		t.Fatalf("failed to write file2: %v", err)
	}
	if err := os.WriteFile(gitFile, []byte("git config content"), 0644); err != nil {
		t.Fatalf("failed to write gitFile: %v", err)
	}
	if err := os.WriteFile(agentFile, []byte("agent plan"), 0644); err != nil {
		t.Fatalf("failed to write agentFile: %v", err)
	}
	if err := os.WriteFile(nodeFile, []byte("node json"), 0644); err != nil {
		t.Fatalf("failed to write nodeFile: %v", err)
	}

	// Create tar archive
	reader, err := createTarArchive(tempDir)
	if err != nil {
		t.Fatalf("createTarArchive failed: %v", err)
	}
	defer reader.Close()

	// Parse tar archive
	tr := tar.NewReader(reader)
	foundFiles := make(map[string]string)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("failed to read next tar entry: %v", err)
		}

		if header.Typeflag == tar.TypeReg {
			content, err := io.ReadAll(tr)
			if err != nil {
				t.Fatalf("failed to read file content from tar: %v", err)
			}
			foundFiles[header.Name] = string(content)
		}
	}

	// Verify files found
	if foundFiles["file1.txt"] != "content1" {
		t.Errorf("expected file1.txt content 'content1', got %q", foundFiles["file1.txt"])
	}
	if foundFiles["subdir/file2.txt"] != "content2" {
		t.Errorf("expected subdir/file2.txt content 'content2', got %q", foundFiles["subdir/file2.txt"])
	}

	// Verify excluded directories
	if _, exists := foundFiles[".git/config"]; exists {
		t.Errorf(".git/config should be excluded from tar archive")
	}
	if _, exists := foundFiles[".agents/plan.md"]; exists {
		t.Errorf(".agents/plan.md should be excluded from tar archive")
	}
	if _, exists := foundFiles["node_modules/package.json"]; exists {
		t.Errorf("node_modules/package.json should be excluded from tar archive")
	}
}

func TestDetectInstaller(t *testing.T) {
	installer, err := detectInstaller()
	if err != nil {
		// If on unsupported OS, this is fine
		if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
			return
		}
		t.Fatalf("detectInstaller failed on supported OS: %v", err)
	}

	if installer == nil {
		t.Fatalf("expected installer, got nil")
	}

	desc := installer.Description()
	if len(desc) == 0 {
		t.Errorf("expected non-empty description")
	}

	manual := installer.ManualInstructions()
	if manual == "" {
		t.Errorf("expected non-empty manual instructions")
	}

	start := installer.StartCommand()
	if start == "" {
		t.Errorf("expected non-empty start command")
	}
}

func TestEnsureDockerHost(t *testing.T) {
	// Test idempotency and behaviour
	oldHost := os.Getenv("DOCKER_HOST")
	defer func() {
		if oldHost == "" {
			os.Unsetenv("DOCKER_HOST")
		} else {
			os.Setenv("DOCKER_HOST", oldHost)
		}
	}()

	// If DOCKER_HOST is already set, it should return immediately
	os.Setenv("DOCKER_HOST", "unix:///mock.sock")
	ensureDockerHost()
	if os.Getenv("DOCKER_HOST") != "unix:///mock.sock" {
		t.Errorf("ensureDockerHost should not override existing DOCKER_HOST")
	}
}
