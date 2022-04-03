package soup

// Version is responsible for exposing version operations to the application layer
type Version interface {
	GetVersionInfo() (*VersionInfo, error)
}

type Git interface {
	PlainClone() error
}
