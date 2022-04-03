package git

type GitRepo struct {
}

func NewGitRepo() GitRepo {
	return GitRepo{}
}

func (g GitRepo) PlainClone() error {
	return nil
}
