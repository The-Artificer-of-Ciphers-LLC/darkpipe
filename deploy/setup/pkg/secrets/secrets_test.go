package secrets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGeneratePassword_Length(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"short", 8},
		{"default", DefaultPasswordLength},
		{"long", 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pw := GeneratePassword(tt.length)
			if len(pw) != tt.length {
				t.Errorf("GeneratePassword(%d) returned length %d, want %d", tt.length, len(pw), tt.length)
			}
		})
	}
}

func TestGeneratePassword_Unique(t *testing.T) {
	// Two generated passwords should not be equal (probabilistically)
	a := GeneratePassword(DefaultPasswordLength)
	b := GeneratePassword(DefaultPasswordLength)
	if a == b {
		t.Errorf("GeneratePassword produced identical passwords: %q", a)
	}
}

func TestGenerateSecrets_CreatesFiles(t *testing.T) {
	dir := t.TempDir()
	secretsDir := filepath.Join(dir, "secrets")

	if err := GenerateSecrets(secretsDir); err != nil {
		t.Fatalf("GenerateSecrets failed: %v", err)
	}

	expectedFiles := []string{"admin_password.txt", "dkim_private_key.pem"}
	for _, name := range expectedFiles {
		path := filepath.Join(secretsDir, name)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("expected file %s not found: %v", name, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("file %s is empty", name)
		}
	}
}

func TestGenerateSecrets_AdminPasswordContent(t *testing.T) {
	dir := t.TempDir()
	secretsDir := filepath.Join(dir, "secrets")

	if err := GenerateSecrets(secretsDir); err != nil {
		t.Fatalf("GenerateSecrets failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(secretsDir, "admin_password.txt"))
	if err != nil {
		t.Fatalf("failed to read admin_password.txt: %v", err)
	}

	if len(content) != DefaultPasswordLength {
		t.Errorf("admin password length = %d, want %d", len(content), DefaultPasswordLength)
	}
}

func TestReadSecret_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	want := "test-secret-value-123"

	if err := writeSecret(dir, "test.txt", want); err != nil {
		t.Fatalf("writeSecret failed: %v", err)
	}

	got, err := ReadSecret(dir, "test.txt")
	if err != nil {
		t.Fatalf("ReadSecret failed: %v", err)
	}

	if got != want {
		t.Errorf("ReadSecret = %q, want %q", got, want)
	}
}

func TestReadSecret_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := ReadSecret(dir, "nonexistent.txt")
	if err == nil {
		t.Error("ReadSecret on nonexistent file should return error")
	}
}

func TestListSecrets(t *testing.T) {
	dir := t.TempDir()

	// Write some secrets
	files := []string{"alpha.txt", "beta.txt", "gamma.txt"}
	for _, name := range files {
		if err := writeSecret(dir, name, "value"); err != nil {
			t.Fatalf("writeSecret(%s) failed: %v", name, err)
		}
	}

	got, err := ListSecrets(dir)
	if err != nil {
		t.Fatalf("ListSecrets failed: %v", err)
	}

	if len(got) != len(files) {
		t.Errorf("ListSecrets returned %d files, want %d", len(got), len(files))
	}

	// Check all expected files are present
	gotMap := make(map[string]bool)
	for _, name := range got {
		gotMap[name] = true
	}
	for _, name := range files {
		if !gotMap[name] {
			t.Errorf("ListSecrets missing file %q", name)
		}
	}
}

func TestListSecrets_ExcludesDirectories(t *testing.T) {
	dir := t.TempDir()

	// Create a file and a subdirectory
	if err := writeSecret(dir, "secret.txt", "value"); err != nil {
		t.Fatalf("writeSecret failed: %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0700); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}

	got, err := ListSecrets(dir)
	if err != nil {
		t.Fatalf("ListSecrets failed: %v", err)
	}

	if len(got) != 1 {
		t.Errorf("ListSecrets returned %d entries (want 1, should exclude dirs): %v", len(got), got)
	}
}

func TestListSecrets_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	got, err := ListSecrets(dir)
	if err != nil {
		t.Fatalf("ListSecrets failed: %v", err)
	}

	if len(got) != 0 {
		t.Errorf("ListSecrets on empty dir returned %d entries, want 0", len(got))
	}
}
