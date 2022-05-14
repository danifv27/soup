package git

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/danifv27/soup/internal/application/audit"
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
	auditer  audit.Auditer
	info     *soup.GitInfo
	repo     *gogit.Repository
	worktree *gogit.Worktree
}

func NewGitRepo(l logger.Logger, audit audit.Auditer) GitRepo {
	return GitRepo{
		logger:  l,
		auditer: audit,
	}
}

func (g *GitRepo) Init(address string, username string, token string) error {
	var u *url.URL
	var err error

	if g.info == nil {
		g.info = new(soup.GitInfo)
	}
	if username == "" {
		g.info.Username = "dummy"
	} else {
		g.info.Username = username
	}
	g.info.Token = token
	if u, err = url.Parse(address); err != nil {
		return err
	}
	//u.User = url.UserPassword(g.info.Username, g.info.Token)
	g.info.Url = u.String()

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
		"url":      g.info.Url,
		"username": g.info.Username,
		"token":    g.info.Token,
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

	event := audit.Event{
		Action:  "PlainClone",
		Actor:   g.info.Username,
		Message: fmt.Sprintf("cloned %s in %s", g.info.Url, location),
	}
	g.auditer.Audit(&event)

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

	event := audit.Event{
		Action:  "GetBranchNames",
		Actor:   g.info.Username,
		Message: fmt.Sprintf("retrieved %d branches", len(branchNames)),
	}
	g.auditer.Audit(&event)

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

	event := audit.Event{
		Action:  "Fetch",
		Actor:   g.info.Username,
		Message: fmt.Sprintf("fetched %s", g.info.Url),
	}
	g.auditer.Audit(&event)
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
	ref := plumbing.ReferenceName(branchName)
	if !(ref.IsBranch() || ref.IsTag()) {
		//if not branch or tag, assume branch
		ref = plumbing.NewBranchReferenceName(branchName)
	}
	err = g.worktree.Checkout(&gogit.CheckoutOptions{
		Branch: ref,
		Force:  true,
	})
	if err != nil {
		return fmt.Errorf("checkout: %w", err)
	}
	event := audit.Event{
		Action:  "Checkout",
		Actor:   g.info.Username,
		Message: fmt.Sprintf("checkout %s", g.info.Url),
	}
	g.auditer.Audit(&event)

	return nil
}

func (g *GitRepo) LsRemote() error {
	var remotes []*plumbing.Reference
	var err error

	if g.info == nil {
		return fmt.Errorf("lsremote: git repo not initialized")
	}
	g.logger.WithFields(logger.Fields{
		"url":      g.info.Url,
		"username": g.info.Username,
		"token":    g.info.Token,
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

	event := audit.Event{
		Action:  "LsRemote",
		Actor:   g.info.Username,
		Message: fmt.Sprintf("listed %d remotes from %s", len(remotes), g.info.Url),
	}
	g.auditer.Audit(&event)

	return nil
}
