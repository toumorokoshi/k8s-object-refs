package refs

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// TODO: handle removing watches when there are no more objects that we're waiting on for that watch.

// TODO: when a queue object is exposed, use that instead.
type QueueContext struct {
	ctx        context.Context
	req        ctrl.Request
	reconciler reconcile.Reconciler
}

// TODO: split to types

type RefManager struct {
	// eventMap          EventMap
	controllerMapping map[GVK]map[types.NamespacedName][]QueueContext
	manager           ctrl.Manager
	// watchers stores handlers to the watchers RefManager has
	// started, to enable proper garbage collection as they are no
	// longer used.
	// watchers map[GVK]Watch
}

func NewRefManager() RefManager {
	return RefManager{
		controllerMapping: make(map[GVK]map[types.NamespacedName][]QueueContext),
	}
}

func (r *RefManager) SetupWithManager(mgr ctrl.Manager) error {
	r.manager = mgr
	return nil
}

// UpdateSubscriptions consumes all namespaces that are
func (r *RefManager) UpdateSubscriptions(gvk GVK, namespaceName types.NamespacedName, queue QueueContext) error {
	source := source.Kind{
		Type: gvk.ToClientObject(),
	}
	if err := r.manager.SetFields(source); err != nil {
		return err
	}
	source.Start()
	r.manager.Add
	// 1. update map
	// 2.
}
