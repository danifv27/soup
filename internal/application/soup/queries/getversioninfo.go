package queries

import (
	"fmt"

	"github.com/danifv27/soup/internal/domain/soup"
)

type GetVersionInfoResult struct {
	Version     string
	Revision    string
	CommitID    string
	BuildAt     string
	BuildBy     string
	BuildBranch string
}

func (v GetVersionInfoResult) String() string {

	return fmt.Sprintf("Version:\t%s\nGit commit:\t%s\nBuilt:\t\t%s (from %s by %s)",
		v.Version, v.CommitID, v.BuildAt, v.BuildBranch, v.BuildBy)
}

//GetCragRequestHandler provides an interfaces to handle a GetCragRequest and return a *GetCragResult
type GetVersionInfoHandler interface {
	Handle() (*GetVersionInfoResult, error)
}

type getVersionInfoHandler struct {
	versionrepo soup.VersionRepository
}

//NewGetCragRequestHandler Handler Constructor
func NewGetVersionInfoHandler(versionrepo soup.VersionRepository) GetVersionInfoHandler {

	return getVersionInfoHandler{versionrepo: versionrepo}
}

//Handle Handlers the GetVersionInfo query
func (h getVersionInfoHandler) Handle() (*GetVersionInfoResult, error) {
	var info *soup.VersionInfo
	var err error

	if info, err = h.versionrepo.GetVersionInfo(); err != nil {
		return nil, err
	}
	result := &GetVersionInfoResult{
		Version:     fmt.Sprint(info.Version),
		Revision:    fmt.Sprint(info.Revision),
		CommitID:    fmt.Sprint(info.GitCommit),
		BuildAt:     fmt.Sprint(info.BuildDate),
		BuildBy:     fmt.Sprint(info.BuildUser),
		BuildBranch: fmt.Sprint(info.Branch),
	}

	return result, nil
}
