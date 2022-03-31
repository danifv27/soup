package soup

// VersionRepository is responsible for exposing version operations to the application layer
type VersionRepository interface {
	GetVersionInfo() (*VersionInfo, error)
}
