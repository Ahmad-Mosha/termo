package ports

import (
	"os"
	"path/filepath"
	"strings"
)

// Project is the git checkout a directory belongs to.
type Project struct {
	Dir    string
	Branch string // empty if dir isn't a git checkout, or the branch can't be read
}

// shortSHA is how many characters of a detached HEAD's commit to show.
const shortSHA = 7

// ProjectFor walks up from dir looking for a git checkout and reports its
// root and current branch. If dir isn't inside one, it's returned as-is
// with no branch.
func ProjectFor(dir string) Project {
	for d := dir; d != ""; {
		if branch, ok := branchAt(filepath.Join(d, ".git")); ok {
			return Project{Dir: d, Branch: branch}
		}
		parent := filepath.Dir(d)
		if parent == d { // reached the filesystem root
			break
		}
		d = parent
	}
	return Project{Dir: dir}
}

// branchAt reads the branch checked out at a .git path, which may be a
// directory (a normal checkout) or a file pointing elsewhere (a worktree
// or submodule). found reports whether a git checkout was there at all,
// regardless of whether its branch could be read.
func branchAt(gitPath string) (branch string, found bool) {
	info, err := os.Stat(gitPath)
	if err != nil {
		return "", false
	}
	dir := gitPath
	if !info.IsDir() {
		data, err := os.ReadFile(gitPath)
		if err != nil {
			return "", true
		}
		gitdir, ok := strings.CutPrefix(strings.TrimSpace(string(data)), "gitdir: ")
		if !ok {
			return "", true
		}
		dir = gitdir
	}
	head, err := os.ReadFile(filepath.Join(dir, "HEAD"))
	if err != nil {
		return "", true
	}
	return parseHead(string(head)), true
}

// parseHead reads a HEAD file: a branch ref, or a commit for a detached
// HEAD.
func parseHead(s string) string {
	s = strings.TrimSpace(s)
	if ref, ok := strings.CutPrefix(s, "ref: refs/heads/"); ok {
		return ref
	}
	if len(s) > shortSHA {
		return s[:shortSHA]
	}
	return ""
}
