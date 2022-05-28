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

type DeployHandler struct {
	logger logger.Logger
	info   *soup.DeployInfo
}

func NewDeployHandler(l logger.Logger) DeployHandler {

	return DeployHandler{
		logger: l,
	}
}

func (d *DeployHandler) Init(path string, c *string) error {

	if d.info == nil {
		d.info = new(soup.DeployInfo)
	}
	d.info.Path = path
	d.info.Context = c

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

// Ping Check k8s cluster availability
func (d *DeployHandler) Ping() error {

	if d.info == nil {
		return fmt.Errorf("ping: deploy repo not initialized")
	}
	config, err := clusterConfig(d.info.Path, d.info.Context)
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

	config, err := clusterConfig(d.info.Path, d.info.Context)
	if err != nil {
		return fmt.Errorf("deploy: %w", err)
	}
	ctx := context.TODO()
	if err = k8s.DoSSA(ctx, config, namespace, yaml); err != nil {
		return fmt.Errorf("deploy: %w", err)
	}

	return nil
}
