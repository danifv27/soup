package kubernetes

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/restmapper"
)

func getGVRFromResource(disco *restmapper.DeferredDiscoveryRESTMapper, resource string) (schema.GroupVersionResource, error) {
	var gvr schema.GroupVersionResource

	if strings.Count(resource, "/") >= 2 {
		s := strings.SplitN(resource, "/", 3)
		gvr = schema.GroupVersionResource{Group: s[0], Version: s[1], Resource: s[2]}
	} else if strings.Count(resource, "/") == 1 {
		s := strings.SplitN(resource, "/", 2)
		gvr = schema.GroupVersionResource{Group: "", Version: s[0], Resource: s[1]}
	}

	if _, err := disco.ResourcesFor(gvr); err != nil {
		return schema.GroupVersionResource{}, err
	}
	return gvr, nil
}

func getNamespace(ns string) string {

	switch ns {
	case "all":
		return metav1.NamespaceAll
	default:
		return ns
	}
}
