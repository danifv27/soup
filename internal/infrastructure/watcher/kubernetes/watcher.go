package kubernetes

import (
	"fmt"
	"net/url"
	"time"

	"github.com/danifv27/soup/internal/application/audit"
	"github.com/danifv27/soup/internal/application/logger"
	"github.com/danifv27/soup/internal/application/watcher"
	"github.com/danifv27/soup/pkg/k8s"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

type WatcherHandler struct {
	logger logger.Logger
	// info     *soup.K8sInfo
	client   *k8s.Client
	informer *multiResourceInformer
}

type Resource struct {
	Kind string
}

func newMultiResourceInformer(res []Resource, resyncPeriod time.Duration, namespaces []string, client *k8s.Client) (*multiResourceInformer, error) {

	informers := make(map[string]map[string]cache.SharedIndexInformer)

	resources := make(map[string]schema.GroupVersionResource)
	for _, r := range res {
		gvr, err := getGVRFromResource(client.DiscoveryMapper, r.Kind)
		if err != nil {
			return nil, err
		}
		resources[r.Kind] = gvr
	}

	dynamicInformers := make([]dynamicinformer.DynamicSharedInformerFactory, 0, len(namespaces))

	for _, ns := range namespaces {

		namespace := getNamespace(ns)
		di := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
			client.DynamicClient,
			resyncPeriod,
			namespace,
			nil,
		)

		for r, gvr := range resources {
			if _, ok := informers[ns]; !ok {
				informers[ns] = make(map[string]cache.SharedIndexInformer)
			}
			informers[ns][r] = di.ForResource(gvr).Informer()
		}

		dynamicInformers = append(dynamicInformers, di)
	}

	return &multiResourceInformer{
		resourceToGVR:      resources,
		resourceToInformer: informers,
		informerFactory:    dynamicInformers,
	}, nil
}

func validateURISchema(uri string) (*url.URL, error) {

	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "informer" {
		return nil, fmt.Errorf("ParseURI: invalid scheme %s", u.Scheme)
	}

	return u, nil
}

// func ParseTimeInterval(source string) (time.Duration, error) {
// 	var intervalRegex = regexp.MustCompile(`([0-9]+)([smhdw])`)
// 	var err error
// 	var val int

// 	matches := intervalRegex.FindStringSubmatch(strings.ToLower(source))

// 	if matches == nil {
// 		return 0, fmt.Errorf("parseTimeInterval: no matches %s", source)
// 	}

// 	if val, err = strconv.Atoi(matches[1]); err != nil {
// 		return 0, err
// 	}

// 	switch matches[2] {
// 	case "s":
// 		return time.Duration(val) * time.Second, nil
// 	case "m":
// 		return time.Duration(val) * time.Second * 60, nil
// 	case "h":
// 		return time.Duration(val) * time.Second * 60 * 60, nil
// 	case "d":
// 		return time.Duration(val) * time.Second * 60 * 60 * 24, nil
// 	case "w":
// 		return time.Duration(val) * time.Second * 60 * 60 * 24 * 7, nil
// 	default:
// 		return 0, fmt.Errorf("parseTimeInterval: unknown unit %s", source)
// 	}
// }

func NewWatcher(uri string, resources []Resource, namespaces []string, l logger.Logger, a audit.Auditer) (*WatcherHandler, error) {
	var err error
	var u *url.URL
	var path, ctx, mode string
	var context *string
	var resync time.Duration

	if u, err = validateURISchema(uri); err != nil {
		return nil, err
	}

	switch u.Opaque {
	case "k8s":
		path = u.Query().Get("path")
		ctx = u.Query().Get("context")
		if ctx == "" {
			context = nil
		}
		context = &ctx
		if resync, err = time.ParseDuration(u.Query().Get("resync")); err != nil {
			return nil, err
		}
		mode = u.Query().Get("mode")
		if mode == "" {
			mode = string(WatchMode)
		}
	default:
		return nil, fmt.Errorf("NewWatcher: unsupported watcher implementation %q", u.Opaque)
	}

	l.WithFields(logger.Fields{
		"path":    path,
		"context": context,
		"mode":    mode,
	}).Debug("Creating k8s config")
	kubeconfig, err := k8s.NewClusterConfig(path, context)
	if err != nil {
		return nil, err
	}

	c, err := k8s.NewClient(kubeconfig)
	if err != nil {
		return nil, err
	}
	l.WithFields(logger.Fields{
		"resources":  resources,
		"resync":     resync,
		"namespaces": namespaces,
	}).Debug("Preparing informer")
	i, err := newMultiResourceInformer(resources, resync, namespaces, c)
	if err != nil {
		return nil, err
	}

	watcher := WatcherHandler{
		logger:   l,
		client:   c,
		informer: i,
	}

	watcher.AddEventHandler(getEventHandler(DiffMode), a)

	return &watcher, nil
}

// AddEventHandler adds the handler to each namespaced informer
func (w WatcherHandler) AddEventHandler(handler watcher.EventHandler, audit audit.Auditer) {

	w.informer.addEventHandler(handler, audit)
}

// HasSynced checks if each namespaced informer has synced
func (w WatcherHandler) HasSynced() bool {

	return w.informer.hasSynced()
}

func (w WatcherHandler) Start(stopCh <-chan struct{}) {

	w.informer.start(stopCh)
}
