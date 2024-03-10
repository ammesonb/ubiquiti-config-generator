package web

import (
	"errors"
	"fmt"
	"github.com/ammesonb/ubiquiti-config-generator/utils"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

var (
	errCreateTempDir      = "failed to create temporary directory"
	errCloneRepo          = "failed to clone repository at %s"
	errOpenRepo           = "failed to open repository at %s"
	errGetWorktree        = "failed to get repository %s worktree"
	errGetHead            = "failed to get repository %s head"
	errGetRemotes         = "failed to get repository remotes for 'origin'"
	errFetchRemotes       = "failed to fetch repository origin remotes"
	errHeadCommit         = "failed to get head commit for path and hash: %+v"
	errTreeHead           = "failed to get current tree head for path and hash: %+v"
	errCheckoutAfterFetch = "failed to checkout branch %s after fetch"
	errDiffTrees          = "failed to diff trees at %s"
)

// cloneRepo will clone a repository and return the folder it was cloned to plus the main branch's ref
func cloneRepo(url string, branch string) (string, *plumbing.Hash, error) {
	dir, err := os.MkdirTemp("", "ubiquiti-config-")
	if err != nil {
		return "", nil, utils.ErrWithParent(errCreateTempDir, err)
	}
	repo, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})

	cleanup := true
	defer func() {
		if !cleanup {
			return
		}
		if err = os.RemoveAll(dir); err != nil {
			fmt.Printf("Failed to remove temp dir: %v\n", err)
		}
	}()

	if err != nil {
		return "", nil, utils.ErrWithCtxParent(errCloneRepo, url, err)
	}

	workTree, err := repo.Worktree()
	if err != nil {
		return "", nil, utils.ErrWithCtxParent(errGetWorktree, url, err)
	}

	head, err := repo.Head()
	if err != nil {
		return "", nil, utils.ErrWithCtxParent(errGetHead, url, err)
	}

	opts := &git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
	}
	// branch may need to be fetched prior to checkout, so don't immediately fail
	if err = workTree.Checkout(opts); err != nil {
		// If remote also fails, return an error, otherwise the checkout resolved as expected
		if err = checkoutRemoteBranch(repo, workTree, branch, opts); err != nil {
			return "", nil, err
		}
	}

	cleanup = false
	h := head.Hash()
	return dir, &h, nil
}

func checkoutRemoteBranch(repo *git.Repository, worktree *git.Worktree, branch string, checkoutOpts *git.CheckoutOptions) error {
	// First, need to fetch origin so get the remote reference
	remote, err := repo.Remote("origin")
	if err != nil {
		return utils.ErrWithParent(errGetRemotes, err)
	}

	// Make a branch reference so we know what we are looking for
	branchRef := plumbing.NewBranchReferenceName(branch).String()
	// Try fetching origin using the remote ref spec of branch:branch
	if err = remote.Fetch(&git.FetchOptions{
		RefSpecs: []config.RefSpec{config.RefSpec(branchRef + ":" + branchRef)},
	}); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		// If fetching remote fails, and it is not because the local instance is already updated
		return utils.ErrWithParent(errFetchRemotes, err)
	} else if err = worktree.Checkout(checkoutOpts); err != nil {
		// Try checking out the branch again
		return utils.ErrWithCtxParent(errCheckoutAfterFetch, branch, err)
	}

	return nil
}

func getChangedFiles(path string, mainHash *plumbing.Hash) ([]string, error) {
	files := make([]string, 0)
	repo, err := git.PlainOpen(path)
	if err != nil {
		return files, utils.ErrWithCtxParent(errOpenRepo, path, err)
	}

	head, err := repo.Head()
	if err != nil {
		return files, utils.ErrWithCtxParent(errGetHead, path, err)
	}

	currentTree, err := getTreeForHash(repo, path, head.Hash())
	if err != nil {
		return files, err
	}
	previousTree, err := getTreeForHash(repo, path, *mainHash)
	if err != nil {
		return files, err
	}

	changes, err := object.DiffTree(previousTree, currentTree)
	if err != nil {
		return files, utils.ErrWithCtxParent(errDiffTrees, path, err)
	}

	for _, change := range changes {
		fileName := change.To.Name
		if fileName == "" {
			fileName = change.From.Name
		}
		files = append(files, fileName)
	}

	return files, nil
}

func getTreeForHash(repo *git.Repository, path string, hash plumbing.Hash) (*object.Tree, error) {
	current, err := repo.CommitObject(hash)
	if err != nil {
		return nil, utils.ErrWithCtxParent(errHeadCommit, struct {
			Path string
			Hash string
		}{Path: path, Hash: hash.String()}, err)
	}
	currentTree, err := current.Tree()
	if err != nil {
		return nil, utils.ErrWithCtxParent(errTreeHead, struct {
			Path string
			Hash string
		}{Path: path, Hash: hash.String()}, err)
	}

	return currentTree, nil
}
