package deployment

import (
	"context"
	"net/url"

	"github.com/danifv27/soup/pkg/k8s"
	"github.com/pkg/errors"
	krest "k8s.io/client-go/rest"
	kcmd "k8s.io/client-go/tools/clientcmd"
)

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

func clusterConfig(path string) (*krest.Config, error) {
	var config *krest.Config
	var err error

	if isValidUrl(path) {
		config, err = kcmd.BuildConfigFromFlags(path, "")
	} else {
		// creates the rest configuration. If neither masterUrl or kubeconfigPath are passed in we fallback to inClusterConfig
		config, err = kcmd.BuildConfigFromFlags("", path)
	}
	if err != nil {
		return nil, err
	}

	return config, nil
}

// Deploy
func Deploy(path string, namespace string, yaml []byte) error {
	config, err := clusterConfig(path)
	if err != nil {
		errors.Wrap(err, "deploy")
		return err
	}
	ctx := context.TODO()
	// for _, manifest := range manifests {
	err = k8s.DoSSA(ctx, config, namespace, yaml)
	if err != nil {
		return errors.Wrap(err, "deploy")
	}
	// }
	return nil
}
