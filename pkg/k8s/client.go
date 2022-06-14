package k8s

import (
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

type Client struct {
	discoveryClient *discovery.DiscoveryClient
	DiscoveryMapper *restmapper.DeferredDiscoveryRESTMapper
	DynamicClient   dynamic.Interface
}

func NewClient(config *rest.Config) (*Client, error) {

	// 1. Prepare a RESTMapper to find GVR
	// DiscoveryClient queries API server about the resources
	disC, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}
	cacheC := memory.NewMemCacheClient(disC)
	cacheC.Invalidate()

	dm := restmapper.NewDeferredDiscoveryRESTMapper(cacheC)

	// 2. Prepare the dynamic client
	dynC, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		discoveryClient: disC,
		DiscoveryMapper: dm,
		DynamicClient:   dynC,
	}, nil
}
