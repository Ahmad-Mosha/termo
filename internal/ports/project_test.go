package ports

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestProjectForRegularCheckout(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".git", "HEAD"), "ref: refs/heads/main\n")
	sub := filepath.Join(root, "src", "pkg")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	got := ProjectFor(sub)
	if got.Dir != root || got.Branch != "main" {
		t.Errorf("got %+v, want dir %q branch main", got, root)
	}
}

func TestProjectForWorktree(t *testing.T) {
	root := t.TempDir()
	realGitdir := filepath.Join(root, "main-checkout", ".git", "worktrees", "feature")
	writeFile(t, filepath.Join(realGitdir, "HEAD"), "ref: refs/heads/feature/cart\n")
	worktree := filepath.Join(root, "feature-checkout")
	writeFile(t, filepath.Join(worktree, ".git"), "gitdir: "+realGitdir+"\n")

	got := ProjectFor(worktree)
	if got.Dir != worktree || got.Branch != "feature/cart" {
		t.Errorf("got %+v, want dir %q branch feature/cart", got, worktree)
	}
}

func TestProjectForDetachedHead(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".git", "HEAD"), "4b825dc642cb6eb9a060e54bf8d69288fbee4904\n")

	got := ProjectFor(root)
	if got.Branch != "4b825dc" {
		t.Errorf("Branch = %q, want the short commit 4b825dc", got.Branch)
	}
}

func TestProjectForNoGit(t *testing.T) {
	dir := t.TempDir()
	got := ProjectFor(dir)
	if got.Dir != dir || got.Branch != "" {
		t.Errorf("got %+v, want dir %q with no branch", got, dir)
	}
}
