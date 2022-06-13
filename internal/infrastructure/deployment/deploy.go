package deployment

import (
	"context"
	"fmt"

	"github.com/danifv27/soup/pkg/k8s"

	"github.com/danifv27/soup/internal/application/logger"
	"github.com/danifv27/soup/internal/domain/soup"
)

type DeployHandler struct {
	logger logger.Logger
	info   *soup.K8sInfo
}

func NewDeployHandler(l logger.Logger) DeployHandler {

	return DeployHandler{
		logger: l,
	}
}

func (d *DeployHandler) Init(path string, c *string) error {

	if d.info == nil {
		d.info = new(soup.K8sInfo)
	}
	d.info.Path = path
	d.info.Context = c

	return nil
}

// Ping Check k8s cluster availability
func (d *DeployHandler) Ping() error {

	if d.info == nil {
		return fmt.Errorf("ping: deploy repo not initialized")
	}
	config, err := k8s.NewClusterConfig(d.info.Path, d.info.Context)
	if err != nil {
		return fmt.Errorf("ping: %w", err)
	}
	ctx := context.TODO()
	err = k8s.DoPing(ctx, config)
	if err != nil {
		return fmt.Errorf("ping: %w", err)
	}

	return nil
}

// Apply apply yaml resources configuration
func (d *DeployHandler) Apply(namespace string, yaml []byte) error {

	config, err := k8s.NewClusterConfig(d.info.Path, d.info.Context)
	if err != nil {
		return fmt.Errorf("deploy: %w", err)
	}
	ctx := context.TODO()
	if err = k8s.DoSSA(ctx, config, namespace, yaml); err != nil {
		return fmt.Errorf("deploy: %w", err)
	}

	return nil
}
