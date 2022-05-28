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
	// GetSoupInfo read and decrypts .soup.yaml file
	GetSoupInfo(root string) SoupInfo
}

type Probe interface {
	//GetLivenessInfo Returns the liveness status
	GetLivenessInfo() (ProbeInfo, error)
	//GetReadinessInfo Returns the readiness status
	GetReadinessInfo() (ProbeInfo, error)
}

type Deploy interface {
	Init(path string, context *string) error
	//Apply  apply yaml resources configuration
	Apply(namespace string, yaml []byte) error
	//Ping Check k8s cluster availability
	Ping() error
}
