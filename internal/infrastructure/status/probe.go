package status

import (
	"github.com/danifv27/soup/internal/deployment"
	"github.com/danifv27/soup/internal/domain/soup"
)

type ProbeRepo struct {
	gitrepo soup.Git
}

//NewProbeRepo Constructor
func NewProbeRepo(git soup.Git) ProbeRepo {

	return ProbeRepo{
		gitrepo: git,
	}
}

//GetLivenessInfo Returns the liveness status
func (m ProbeRepo) GetLivenessInfo() (soup.ProbeInfo, error) {

	i := new(soup.ProbeInfo)

	i.Result = soup.Healthy
	i.Msg = "System Alive"

	return *i, nil
}

//GetReadinessInfo Returns the liveness status
func (m ProbeRepo) GetReadinessInfo() (soup.ProbeInfo, error) {

	i := new(soup.ProbeInfo)

	if err := m.gitrepo.LsRemote(); err != nil {
		i.Result = soup.Unhealthy
		i.Msg = err.Error()
		return *i, err
	}

	if err := deployment.Ping("http://localhost:8001"); err != nil {
		i.Result = soup.Unhealthy
		i.Msg = err.Error()
		return *i, err
	}

	i.Result = soup.Healthy
	i.Msg = "System Ready"

	return *i, nil
}
