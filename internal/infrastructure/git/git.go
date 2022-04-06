package git

import (
	"fmt"
	"strings"

	"github.com/danifv27/soup/internal/application/logger"
	gogit "github.com/go-git/go-git/v5"
	config "github.com/go-git/go-git/v5/config"
	plumbing "github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/pkg/errors"
)

type GitRepo struct {
	logger   logger.Logger
	repo     *gogit.Repository
	worktree *gogit.Worktree
}

func NewGitRepo(l logger.Logger) GitRepo {
	return GitRepo{
		logger: l,
	}
}

func (g *GitRepo) PlainClone(location string, url string, username string, token string) error {
	var err error
	var r *gogit.Repository

	if username == "" {
		username = "dummy"
	}
	g.logger.WithFields(logger.Fields{
		"url":      url,
		"location": location,
		"token":    token,
		"username": username,
	}).Info("Cloning git repository")
	// Authentication
	auth := http.BasicAuth{
		Username: username,
		Password: token,
	}
	if r, err = gogit.PlainClone(location, false, &gogit.CloneOptions{
		Auth: &auth,
		URL:  url,
	}); err != nil {

		return errors.Wrap(err, "plainclone")
	}

	g.repo = r

	return nil
}

func (g *GitRepo) GetBranchNames(username string, token string) ([]string, error) {
	var branchNames []string

	if g.repo == nil {
		return nil, fmt.Errorf("git repository not cloned")
	}
	remote, err := g.repo.Remote("origin")
	if err != nil {
		return nil, errors.Wrap(err, "remote")
	}
	if username == "" {
		username = "dummy"
	}
	auth := http.BasicAuth{
		Username: username,
		Password: token,
	}
	refList, err := remote.List(&gogit.ListOptions{
		Auth: &auth,
	})
	if err != nil {
		return nil, errors.Wrap(err, "remote listing")
	}
	refPrefix := "refs/heads/"
	for _, ref := range refList {
		refName := ref.Name().String()
		if !strings.HasPrefix(refName, refPrefix) {
			continue
		}
		branchName := refName[len(refPrefix):]
		branchNames = append(branchNames, branchName)
	}

	return branchNames, nil
}

func (g *GitRepo) Fetch(username string, token string) error {

	if username == "" {
		username = "dummy"
	}
	auth := http.BasicAuth{
		Username: username,
		Password: token,
	}

	err := g.repo.Fetch(&gogit.FetchOptions{
		Auth:     &auth,
		RefSpecs: []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"},
	})
	if err != nil {
		return errors.Wrap(err, "fetch")
	}

	return nil
}

//Checkout checks out cloned repository
func (g *GitRepo) Checkout(branchName string) error {
	var err error

	if g.worktree == nil {
		if g.worktree, err = g.repo.Worktree(); err != nil {
			return errors.Wrap(err, "can not create workspace")
		}
	}
	err = g.worktree.Checkout(&gogit.CheckoutOptions{
		Branch: plumbing.ReferenceName("refs/heads/" + branchName),
		Force:  true,
	})
	if err != nil {
		return errors.Wrap(err, "checkout")
	}

	return nil
}
