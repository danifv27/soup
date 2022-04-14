package git

import (
	"fmt"
	"strings"

	"github.com/danifv27/soup/internal/application/logger"
	"github.com/danifv27/soup/internal/domain/soup"
	gogit "github.com/go-git/go-git/v5"
	config "github.com/go-git/go-git/v5/config"
	plumbing "github.com/go-git/go-git/v5/plumbing"
	transport "github.com/go-git/go-git/v5/plumbing/transport/http"
	memory "github.com/go-git/go-git/v5/storage/memory"
	"github.com/pkg/errors"
)

type GitRepo struct {
	logger   logger.Logger
	info     *soup.GitInfo
	repo     *gogit.Repository
	worktree *gogit.Worktree
}

func NewGitRepo(l logger.Logger) GitRepo {
	return GitRepo{
		logger: l,
	}
}

func (g *GitRepo) InitRepo(url string, username string, token string) error {

	if g.info == nil {
		g.info = new(soup.GitInfo)
	}
	g.info.Url = url
	if username == "" {
		g.info.Username = "dummy"
	} else {
		g.info.Username = username
	}
	g.info.Token = token

	return nil
}

func (g *GitRepo) PlainClone(location string) error {
	var err error
	var r *gogit.Repository

	if g.info == nil {
		return errors.Wrap(fmt.Errorf("git repo not initialized"), "PlainClone")
	}
	g.logger.WithFields(logger.Fields{
		"location": location,
		"info":     g.info,
	}).Info("Cloning git repository")
	// Authentication
	auth := transport.BasicAuth{
		Username: g.info.Username,
		Password: g.info.Token,
	}
	if r, err = gogit.PlainClone(location, false, &gogit.CloneOptions{
		Auth: &auth,
		URL:  g.info.Url,
	}); err != nil {

		return errors.Wrap(err, "plainclone")
	}

	g.repo = r

	return nil
}

func (g *GitRepo) GetBranchNames() ([]string, error) {
	var branchNames []string

	if g.info == nil {
		return nil, errors.Wrap(fmt.Errorf("git repo not initialized"), "GetBranchNames")
	}
	if g.repo == nil {
		return nil, fmt.Errorf("git repository not cloned")
	}
	remote, err := g.repo.Remote("origin")
	if err != nil {
		return nil, errors.Wrap(err, "remote")
	}
	auth := transport.BasicAuth{
		Username: g.info.Username,
		Password: g.info.Token,
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

func (g *GitRepo) Fetch() error {

	if g.info == nil {
		return errors.Wrap(fmt.Errorf("git repo not initialized"), "Fetch")
	}
	auth := transport.BasicAuth{
		Username: g.info.Username,
		Password: g.info.Token,
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

func (g *GitRepo) LsRemote() error {
	var remotes []*plumbing.Reference
	var err error

	if g.info == nil {
		return errors.Wrap(fmt.Errorf("git repo not initialized"), "LsRemote")
	}
	g.logger.WithFields(logger.Fields{
		"info": g.info,
	}).Debug("ls-remote repository")
	// Authentication
	auth := transport.BasicAuth{
		Username: g.info.Username,
		Password: g.info.Token,
	}
	// Create the remote with repository URL
	rem := gogit.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{g.info.Url},
	})

	if remotes, err = rem.List(&gogit.ListOptions{
		Auth: &auth,
	}); err != nil {
		return errors.Wrap(err, "LsRemote")
	}
	g.logger.WithFields(logger.Fields{
		"remotes": remotes,
	}).Info("Remote List")

	return nil
}
