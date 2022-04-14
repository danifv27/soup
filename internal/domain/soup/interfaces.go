package soup

// Version is responsible for exposing version operations to the application layer
type Version interface {
	GetVersionInfo() (*VersionInfo, error)
}

type Git interface {
	Init(url string, username string, token string) error
	PlainClone(location string) error
	GetBranchNames() ([]string, error)
	Fetch() error
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
