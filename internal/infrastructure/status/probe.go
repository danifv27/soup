package status

import (
	"github.com/danifv27/soup/internal/domain/soup"
)

type ProbeRepo struct {
}

//NewProbeRepo Constructor
func NewProbeRepo() ProbeRepo {

	return ProbeRepo{}
}

//GetLivenessInfo Returns the liveness status
func (m ProbeRepo) GetLivenessInfo() (soup.ProbeInfo, error) {

	i := new(soup.ProbeInfo)

	// Liven
	i.Result = soup.Healthy
	i.Msg = "System UP"

	return *i, nil
}
