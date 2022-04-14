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

func (g *GitRepo) Init(url string, username string, token string) error {

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
		return fmt.Errorf("plainclone: git repo not initialized")
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
		return fmt.Errorf("plainclone: %w", err)
	}

	g.repo = r

	return nil
}

func (g *GitRepo) GetBranchNames() ([]string, error) {
	var branchNames []string

	if g.info == nil {
		return nil, fmt.Errorf("getBranchNames: git repo not initialized")
	}
	if g.repo == nil {
		return nil, fmt.Errorf("getBranchNames: git repository not cloned")
	}
	remote, err := g.repo.Remote("origin")
	if err != nil {

		return nil, fmt.Errorf("getBranchNames: %w", err)
	}
	auth := transport.BasicAuth{
		Username: g.info.Username,
		Password: g.info.Token,
	}
	refList, err := remote.List(&gogit.ListOptions{
		Auth: &auth,
	})
	if err != nil {
		return nil, fmt.Errorf("getBranchNames: %w", err)
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
		return fmt.Errorf("fetch: git repo not initialized")
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
		return fmt.Errorf("fetch: %w", err)
	}

	return nil
}

//Checkout checks out cloned repository
func (g *GitRepo) Checkout(branchName string) error {
	var err error

	if g.worktree == nil {
		if g.worktree, err = g.repo.Worktree(); err != nil {
			return fmt.Errorf("checkout: %w", err)
		}
	}
	err = g.worktree.Checkout(&gogit.CheckoutOptions{
		Branch: plumbing.ReferenceName("refs/heads/" + branchName),
		Force:  true,
	})
	if err != nil {
		return fmt.Errorf("checkout: %w", err)
	}

	return nil
}

func (g *GitRepo) LsRemote() error {
	var remotes []*plumbing.Reference
	var err error

	if g.info == nil {
		return fmt.Errorf("lsremote: git repo not initialized")
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
		return fmt.Errorf("lsremote: %w", err)
	}
	g.logger.WithFields(logger.Fields{
		"remotes": remotes,
	}).Info("Remote List")

	return nil
}
