package status

import (
	"github.com/danifv27/soup/internal/domain/soup"
)

type ProbeRepo struct {
	gitrepo    soup.Git
	deployrepo soup.Deploy
}

//NewProbeRepo Constructor
func NewProbeRepo(git soup.Git, deploy soup.Deploy) ProbeRepo {

	return ProbeRepo{
		gitrepo:    git,
		deployrepo: deploy,
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

	if err := m.deployrepo.Ping(); err != nil {
		i.Result = soup.Unhealthy
		i.Msg = err.Error()
		return *i, err
	}

	i.Result = soup.Healthy
	i.Msg = "System Ready"

	return *i, nil
}
