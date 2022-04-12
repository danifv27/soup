package status

import (
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

	err := m.gitrepo.LsRemote("https://github.com/danifv27/helloDeploy.git", "danifv27", "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")
	if err != nil {
		i.Result = soup.Unhealthy
		i.Msg = err.Error()
	} else {
		i.Result = soup.Healthy
		i.Msg = "System Ready"
	}

	return *i, nil
}
