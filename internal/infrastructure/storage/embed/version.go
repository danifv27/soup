package embed

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/danifv27/soup/internal/domain/soup"
)

var (
	//embed doesn't allow cross package boundaries, so version.json should be in this folder
	//go:embed version.json
	version string
	//Application Name
	Name string
)

type VersionRepo struct {
	versioninfo soup.VersionInfo
}

//NewVersionRepo Constructor
func NewVersionRepo() VersionRepo {
	var j map[string]interface{}

	if err := json.Unmarshal([]byte(version), &j); err != nil {
		return VersionRepo{}
	}

	v := new(soup.VersionInfo)
	// The value of your map associated with key "git" is of type map[string]interface{}.
	// And we want to access the element of that map associated with the key "commit".
	// .(string) type assertion to convert interface{} to string
	v.GitCommit = j["git"].(map[string]interface{})["commit"].(string)
	v.Branch = j["git"].(map[string]interface{})["branch"].(string)
	v.Version = j["version"].(string)
	v.Revision = j["revision"].(string)
	v.BuildDate = j["build"].(map[string]interface{})["date"].(string)
	v.BuildUser = j["build"].(map[string]interface{})["user"].(string)
	v.GoVersion = runtime.Version()
	v.OsArch = fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)
	// v.Name = Name

	return VersionRepo{
		versioninfo: *v,
	}
}

//GetVersionInfo Returns the version information
func (m VersionRepo) GetVersionInfo() (*soup.VersionInfo, error) {

	return &m.versioninfo, nil
}
