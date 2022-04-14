package deployment

import (
	"context"
	"fmt"
	"net/url"

	"github.com/danifv27/soup/pkg/k8s"
	krest "k8s.io/client-go/rest"
	kcmd "k8s.io/client-go/tools/clientcmd"

	"github.com/danifv27/soup/internal/application/logger"
	"github.com/danifv27/soup/internal/domain/soup"
)

type DeployRepo struct {
	logger logger.Logger
	info   *soup.DeployInfo
	config *krest.Config
}

func NewDeployRepo(l logger.Logger) DeployRepo {

	return DeployRepo{
		logger: l,
	}
}

func (d *DeployRepo) Init(path string, c *string) error {
	var err error

	if d.info == nil {
		d.info = new(soup.DeployInfo)
	}
	d.info.Path = path

	d.config, err = clusterConfig(d.info.Path, c)
	if err != nil {
		return fmt.Errorf("init: %w", err)
	}

	return nil
}

// isValidUrl tests a string to determine if it is a well-structured url or not.
func isValidUrl(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}

	u, err := url.Parse(toTest)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

func clusterConfig(path string, context *string) (*krest.Config, error) {
	var config *krest.Config
	var err error

	if isValidUrl(path) {
		config, err = kcmd.BuildConfigFromFlags(path, "")
	} else {
		// creates the rest configuration. If neither masterUrl or kubeconfigPath are passed in we fallback to inClusterConfig
		if context == nil {
			config, err = kcmd.BuildConfigFromFlags("", path)
		} else {
			rules := kcmd.NewDefaultClientConfigLoadingRules()
			rules.DefaultClientConfig = &kcmd.DefaultClientConfig
			overrides := &kcmd.ConfigOverrides{ClusterDefaults: kcmd.ClusterDefaults}
			overrides.CurrentContext = *context
			config, err = kcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides).ClientConfig()
		}
	}
	if err != nil {
		return nil, err
	}

	return config, nil
}

// Ping
func (d *DeployRepo) Ping() error {

	if d.info == nil {
		return fmt.Errorf("ping: deploy repo not initialized")
	}

	ctx := context.TODO()
	err := k8s.DoPing(ctx, d.config)
	if err != nil {
		return fmt.Errorf("ping: %w", err)
	}
	return nil
}

// Deploy
func (d *DeployRepo) Deploy(namespace string, yaml []byte) error {

	ctx := context.TODO()
	if err := k8s.DoSSA(ctx, d.config, namespace, yaml); err != nil {
		return fmt.Errorf("deploy: %w", err)
	}

	return nil
}
