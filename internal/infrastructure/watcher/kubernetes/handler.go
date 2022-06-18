package kubernetes

import (
	"fmt"

	"github.com/go-test/deep"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/cache"

	"github.com/danifv27/soup/internal/application/audit"
	"github.com/danifv27/soup/internal/application/watcher"
)

const (
	EventAdd    string = "EventAdd"
	EventUpdate string = "EventUpdate"
	EventDelete string = "EventDelete"
	EventDiff   string = "EventDiff"
)

const (
	WatchMode runMode = "watch"
	DiffMode  runMode = "diff"
)

type runMode string

func watchHandler(resourceType string, auditer audit.Auditer) cache.ResourceEventHandlerFuncs {
	var handler cache.ResourceEventHandlerFuncs

	handler.AddFunc = func(obj interface{}) {
		o := obj.(*unstructured.Unstructured)
		e := audit.Event{
			Action:  EventAdd,
			Actor:   "TODO find k8s serviceAccount",
			Message: fmt.Sprintf("%s (%s): %v", o.Object["metadata"].(map[string]interface{})["name"], resourceType, obj),
		}
		auditer.Audit(&e)
	}

	handler.UpdateFunc = func(old, new interface{}) {
		o := new.(*unstructured.Unstructured)
		e := audit.Event{
			Action:  EventUpdate,
			Actor:   "TODO find k8s serviceAccount",
			Message: fmt.Sprintf("%s (%s): %v", o.Object["metadata"].(map[string]interface{})["name"], resourceType, new),
		}
		auditer.Audit(&e)
	}

	handler.DeleteFunc = func(obj interface{}) {
		o := obj.(*unstructured.Unstructured)
		e := audit.Event{
			Action:  EventDelete,
			Actor:   "TODO find k8s serviceAccount",
			Message: fmt.Sprintf("%s (%s): %v", o.Object["metadata"].(map[string]interface{})["name"], resourceType, obj),
		}
		auditer.Audit(&e)
	}

	return handler
}

func diffHandler(resourceType string, auditer audit.Auditer) cache.ResourceEventHandlerFuncs {
	var handler cache.ResourceEventHandlerFuncs

	// handler.AddFunc = func(obj interface{}) {
	// 	o := obj.(*unstructured.Unstructured)
	// 	e := audit.Event{
	// 		Action:  EventAdd,
	// 		Actor:   "TODO find k8s serviceAccount",
	// 		Message: fmt.Sprintf("%s (%s): %v", o.Object["metadata"].(map[string]interface{})["name"], resourceType, obj),
	// 	}
	// 	auditer.Audit(&e)
	// }

	handler.UpdateFunc = func(old, new interface{}) {
		oldObj := old.(*unstructured.Unstructured)
		newObj := new.(*unstructured.Unstructured)

		if !equality.Semantic.DeepEqual(old, new) {
			diff := deep.Equal(oldObj, newObj)
			e := audit.Event{
				Action:  EventDiff,
				Actor:   "TODO find k8s serviceAccount",
				Message: fmt.Sprintf("%s (%s): %v ", newObj.Object["metadata"].(map[string]interface{})["name"], resourceType, diff),
			}
			auditer.Audit(&e)
		}
	}

	// handler.DeleteFunc = func(obj interface{}) {
	// 	o := obj.(*unstructured.Unstructured)
	// 	e := audit.Event{
	// 		Action:  EventDelete,
	// 		Actor:   "TODO find k8s serviceAccount",
	// 		Message: fmt.Sprintf("%s (%s): %v", o.Object["metadata"].(map[string]interface{})["name"], resourceType, obj),
	// 	}
	// 	auditer.Audit(&e)
	// }

	return handler
}

func getEventHandler(mode runMode) watcher.EventHandler {

	switch mode {
	case DiffMode:
		return diffHandler
	default:
		return watchHandler
	}
}
