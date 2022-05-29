package queries

import (
	"fmt"

	"github.com/danifv27/soup/internal/domain/soup"
)

type GetLivenessInfoResult struct {
	Result soup.ProbeResultType
	Msg    string
}

func (v GetLivenessInfoResult) String() string {

	return fmt.Sprintf("%d - %s)", v.Result, v.Msg)
}

//GetCragRequestHandler provides an interfaces to handle a GetCragRequest and return a *GetCragResult
type GetLivenessInfoHandler interface {
	Handle() (GetLivenessInfoResult, error)
}

type getLivenessInfoHandler struct {
	probesrepo soup.Probe
}

//NewGetLivenessInfoHandler Handler Constructor
func NewGetLivenessInfoHandler(repo soup.Probe) GetLivenessInfoHandler {

	return getLivenessInfoHandler{probesrepo: repo}
}

//Handle Handlers the GetLivenessInfo query
func (h getLivenessInfoHandler) Handle() (GetLivenessInfoResult, error) {
	var info soup.ProbeInfo
	var err error
	var result GetLivenessInfoResult

	if info, err = h.probesrepo.GetLivenessInfo(); err != nil {
		result = GetLivenessInfoResult{
			Result: info.Result,
			Msg:    err.Error(),
		}
	} else {
		result = GetLivenessInfoResult{
			Result: info.Result,
			Msg:    info.Msg,
		}
	}

	return result, nil
}
