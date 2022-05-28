package k8s // import "github.com/caldito/soup/pkg/k8s"

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	yamlk8s "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

//DoSSA deploy resources using declarative configuration (Server Side Apply).
func DoSSA(ctx context.Context, cfg *rest.Config, namespace string, yamlFile []byte) error {
	var decUnstructured = yamlk8s.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	// 1. Prepare a RESTMapper to find GVR
	// DiscoveryClient queries API server about the resources
	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	// 2. Prepare the dynamic client
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return err
	}

	// 3. Decode YAML manifest into unstructured.Unstructured
	obj := &unstructured.Unstructured{}
	_, gvk, err := decUnstructured.Decode(yamlFile, nil, obj)
	if err != nil {
		return err
	}

	// 4. Find the corresponding GVR (available in *meta.RESTMapping) for GVK
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return err
	}

	// 5. Obtain REST interface for the GVR and set namespace
	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// namespaced resources should specify the namespace
		obj.SetNamespace(namespace)
		dr = dyn.Resource(mapping.Resource).Namespace(obj.GetNamespace())
	} else {
		// for cluster-wide resources
		dr = dyn.Resource(mapping.Resource)
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
