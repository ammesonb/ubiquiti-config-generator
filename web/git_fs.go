package web

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// cloneRepo will clone a repository and return the folder it was cloned to
func cloneRepo(url string, branch string) (string, error) {
	dir := os.TempDir()
	repo, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL:      "url",
		Progress: os.Stdout,
	})
	if err != nil {
		defer func() {
			if err = os.Remove(dir); err != nil {
				fmt.Printf("Failed to remove temp dir: %v\n", err)
			}
		}()

		return "", fmt.Errorf("failed to clone repository %s: %v", url, err)
	}

	workTree, err := repo.Worktree()
	if err != nil {
		defer func() {
			if err = os.Remove(dir); err != nil {
				fmt.Printf("Failed to remove temp dir: %v\n", err)
			}
		}()

		return "", fmt.Errorf("failed to get repository %s worktree: %v", url, err)
	}

	if err = workTree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
	}); err != nil {
		defer func() {
			if err = os.Remove(dir); err != nil {
				fmt.Printf("Failed to remove temp dir: %v\n", err)
			}
		}()

		return "", fmt.Errorf("failed to check out branch %s: %v", plumbing.NewBranchReferenceName(branch), err)
	}

	return dir, nil
}
