package soup

// Version is responsible for exposing version operations to the application layer
type Version interface {
	GetVersionInfo() (*VersionInfo, error)
}

type Git interface {
	InitRepo(url string, username string, token string) error
	PlainClone(location string, url string, username string, token string) error
	GetBranchNames(username string, token string) ([]string, error)
	Fetch(username string, token string) error
	Checkout(branchName string) error
	LsRemote() error
}

type Config interface {
	GetSoupInfo(root string) SoupInfo
}

type Probe interface {
	GetLivenessInfo() (ProbeInfo, error)
	GetReadinessInfo() (ProbeInfo, error)
}
