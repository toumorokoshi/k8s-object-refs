package refs

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// TODO: handle removing watches when there are no more objects that we're waiting on for that watch.

// TODO: when a queue object is exposed, use that instead.
type QueueContext struct {
	Context    context.Context
	Req        ctrl.Request
	Reconciler reconcile.Reconciler
}

// TODO: split to types

type RefManager struct {
	EventMapping         map[GVK]map[types.NamespacedName][]QueueContext
	Manager              ctrl.Manager
	ContextsByController map[GVK]context.Context

	// watchers stores handlers to the watchers RefManager has
	// started, to enable proper garbage collection as they are no
	// longer used.
	// watchers map[GVK]Watch
}

func NewRefManager() RefManager {
	return RefManager{
		EventMapping:         make(map[GVK]map[types.NamespacedName][]QueueContext),
		ContextsByController: make(map[GVK]context.Context),
	}
}

func (r *RefManager) SetupWithManager(mgr ctrl.Manager) error {
	r.Manager = mgr
	return nil
}

// UpdateSubscriptions consumes all namespaces that are
func (r *RefManager) UpdateSubscriptions(gvk GVK, namespacedName types.NamespacedName, queue QueueContext) error {
	contextByNamespaceName := make(map[types.NamespacedName][]QueueContext)
	contextByNamespaceName[namespacedName] = []QueueContext{queue}
	r.EventMapping[gvk] = contextByNamespaceName
	r.startController(gvk)
	return nil
}

// TODO handle errors
func (r *RefManager) startController(gvk GVK) error {
	// if a context already exists, then we already spawned
	// a controller.
	if context := r.ContextsByController[gvk]; context != nil {
		return nil
	}
	c, err := controller.NewUnmanaged(fmt.Sprintf("%v", gvk), r.Manager, controller.Options{
		Reconciler: &RefReconciler{
			Client:  r.Manager.GetClient(),
			manager: r,
			gvk:     gvk,
		},
	})
	if err != nil {
		r.Manager.GetLogger().Error(err, "unable to create manager")
	}
	if err := c.Watch(&source.Kind{Type: gvk.ToClientObject()}, &handler.EnqueueRequestForObject{}); err != nil {
		r.Manager.GetLogger().Error(err, "unable to watch")
	}

	ctx := context.Background()
	go func() {
		if err := c.Start(ctx); err != nil {
			// TODO: error handling
		}
	}()
	r.ContextsByController[gvk] = ctx
	return nil
}
