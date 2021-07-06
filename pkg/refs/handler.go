package refs

import (
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

// EventHandler functions
// TODO: break out eventHandlder into it's own struct.
type EnqueueRequestInRefererrControllers struct {
	refManager *RefManager
}

func (r *RefManager) Create(e event.CreateEvent, q workqueue.RateLimitingInterface) {
	r.fireReconcile(e.Object)
}

func (r *RefManager) Delete(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
	r.fireReconcile(e.Object)
}

func (r *RefManager) Update(e event.UpdateEvent, q workqueue.RateLimitingInterface) {
	r.fireReconcile(e.ObjectNew)
}

func (r *RefManager) Generic(e event.GenericEvent, q workqueue.RateLimitingInterface) {
	r.fireReconcile(e.Object)
}

func (r *RefManager) fireReconcile(object client.Object) {
	gvk := GVKFromClientObject(object)
	subscribers := r.controllerMapping[gvk]
	if subscribers == nil {
		return
	}
	namespacedName := types.NamespacedName{
		Namespace: object.GetNamespace(),
		Name:      object.GetName(),
	}
	queueContexts := subscribers[namespacedName]
	for _, qc := range queueContexts {
		qc.reconciler.Reconcile(qc.ctx, qc.req)
	}
}
