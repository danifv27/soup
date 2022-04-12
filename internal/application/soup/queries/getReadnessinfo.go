package queries

import (
	"fmt"

	"github.com/danifv27/soup/internal/domain/soup"
)

type GetReadinessInfoResult struct {
	Result soup.ProbeResultType
	Msg    string
}

func (v GetReadinessInfoResult) String() string {

	return fmt.Sprintf("%d - %s)", v.Result, v.Msg)
}

//GetCragRequestHandler provides an interfaces to handle a GetCragRequest and return a *GetCragResult
type GetReadinessInfoHandler interface {
	Handle() (GetReadinessInfoResult, error)
}

type getReadnessInfoHandler struct {
	probesrepo soup.Probe
	gitrepo    soup.Git
}

//NewGetCragRequestHandler Handler Constructor
func NewGetReadinessInfoHandler(repo soup.Probe, git soup.Git) GetReadinessInfoHandler {

	return getReadnessInfoHandler{
		probesrepo: repo,
		gitrepo:    git,
	}
}

//Handle Handlers the GetReadinessInfo query
func (h getReadnessInfoHandler) Handle() (GetReadinessInfoResult, error) {
	var info soup.ProbeInfo
	var err error

	if info, err = h.probesrepo.GetReadinessInfo(); err != nil {
		return GetReadinessInfoResult{}, err
	}
	result := GetReadinessInfoResult{
		Result: info.Result,
		Msg:    info.Msg,
	}

	return result, nil
}
