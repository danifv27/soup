package git

import (
	"fmt"
	"strings"

	gogit "github.com/go-git/go-git/v5"
)

type GitRepo struct {
	repo *gogit.Repository
}

func NewGitRepo() GitRepo {
	return GitRepo{}
}

func (g *GitRepo) PlainClone(location string, url string) error {
	var err error
	var r *gogit.Repository

	if r, err = gogit.PlainClone(location, false, &gogit.CloneOptions{
		URL: url,
	}); err != nil {
		return err
	}

	g.repo = r

	return nil
}

func (g *GitRepo) GetBranchNames() ([]string, error) {
	var branchNames []string

	if g.repo == nil {
		return nil, fmt.Errorf("git repository not cloned")
	}
	remote, err := g.repo.Remote("origin")
	if err != nil {
		return nil, err
	}
	refList, err := remote.List(&gogit.ListOptions{})
	if err != nil {
		return nil, err
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
