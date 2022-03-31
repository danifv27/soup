package soup

type VersionInfo struct {
	//GitCommit The git commit that was compiled. This will be filled in by the compiler.
	GitCommit string
	//Version The main version number that is being run at the moment.
	Version  string
	Revision string
	Branch   string
	//BuildDate This will be filled in by the makefile
	BuildDate string
	BuildUser string
	//GoVersion The runtime version
	GoVersion string
	//OsArch The OS architecture
	OsArch string
	//Application Name
	Name string
}
