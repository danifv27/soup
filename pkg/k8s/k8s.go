package k8s // import "github.com/caldito/soup/pkg/k8s"

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	yamlk8s "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

//NewClusterConfig creates the rest configuration. If neither masterUrl or kubeconfigPath are passed in path argument we fallback to inClusterConfig
func NewClusterConfig(path string, context *string) (*krest.Config, error) {
	var config *krest.Config
	var err error

	if isValidUrl(path) {
		// masterUrl detectected
		config, err = kcmd.BuildConfigFromFlags(path, "")
	} else {
		if context != nil {
			// kubeconfig plus context
			config, err = kcmd.BuildConfigFromFlags("", path)
		} else {
			// inClusterConfig
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

//DoSSA deploy resources using declarative configuration (Server Side Apply).
func DoSSA(ctx context.Context, cfg *rest.Config, namespace string, yamlFile []byte) error {
	var decUnstructured = yamlk8s.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	var client *Client
	var err error

	// 1. Prepare a RESTMapper to find GVR
	// 2. Prepare the dynamic client
	if client, err = NewClient(cfg); err != nil {
		return err
	}

	// 3. Decode YAML manifest into unstructured.Unstructured
	obj := &unstructured.Unstructured{}
	_, gvk, err := decUnstructured.Decode(yamlFile, nil, obj)
	if err != nil {
		return err
	}

	// 4. Find the corresponding GVR (available in *meta.RESTMapping) for GVK
	mapping, err := client.DiscoveryMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return err
	}

	// 5. Obtain REST interface for the GVR and set namespace
	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// namespaced resources should specify the namespace
		obj.SetNamespace(namespace)
		dr = client.DynamicClient.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	} else {
		// for cluster-wide resources
		dr = client.DynamicClient.Resource(mapping.Resource)
	}

	// 6. Marshal object into JSON
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	// 7. Create or Update the object with SSA
	//     types.ApplyPatchType indicates SSA.
	//     FieldManager specifies the field owner ID.
	_, err = dr.Patch(ctx, obj.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
		FieldManager: "soup-controller",
	})

	return err
}

func DoPing(ctx context.Context, cfg *rest.Config) error {

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}
	path := "/healthz"
	content, err := client.Discovery().RESTClient().Get().AbsPath(path).DoRaw(ctx)
	if err != nil {
		return err
	}
	contentStr := string(content)
	if contentStr != "ok" {
		return fmt.Errorf("unreachable k8s cluster")
	}

	return nil
}
