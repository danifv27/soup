package kubernetes

import (
	"github.com/danifv27/soup/internal/application/audit"
	"github.com/danifv27/soup/internal/application/watcher"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

type multiResourceInformer struct {
	//GroupVersionResource per kind of resource to watch
	resourceToGVR map[string]schema.GroupVersionResource
	//informer per namespace and resource
	resourceToInformer map[string]map[string]cache.SharedIndexInformer
	//one factory per namespace
	informerFactory []dynamicinformer.DynamicSharedInformerFactory
}

// addEventHandler adds the handler to each namespaced informer
func (i *multiResourceInformer) addEventHandler(handler watcher.EventHandler, audit audit.Auditer) {
	for _, ki := range i.resourceToInformer {
		for kind, informer := range ki {
			informer.AddEventHandler(handler(kind, audit))
		}
	}
}

// hasSynced checks if each namespaced informer has synced
func (i *multiResourceInformer) hasSynced() bool {

	for _, ki := range i.resourceToInformer {
		for _, informer := range ki {
			if ok := informer.HasSynced(); !ok {
				return ok
			}
		}
	}

	return true
}

func (i *multiResourceInformer) start(stopCh <-chan struct{}) {

	for _, informer := range i.informerFactory {
		informer.Start(stopCh)
	}
}
