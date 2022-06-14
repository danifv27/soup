package noop

type GitRepo struct{}

func NewGit() GitRepo {

	return GitRepo{}
}

// func (g GitRepo) Init(address string, username string, token string) error {

// 	return nil
// }

func (g GitRepo) PlainClone(location string) error {

	return nil
}

func (g GitRepo) GetBranchNames() ([]string, error) {

	return []string{}, nil
}

func (g GitRepo) Fetch() error {

	return nil
}

func (g GitRepo) Checkout(branchName string) error {

	return nil
}

func (g GitRepo) LsRemote() error {

	return nil
}
