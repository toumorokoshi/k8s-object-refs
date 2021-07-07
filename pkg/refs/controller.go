package refs

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RefReconciler struct {
	client.Client
	manager *RefManager
	gvk     GVK
}

func (r *RefReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	object := &unstructured.Unstructured{}
	if err := r.Get(ctx, req.NamespacedName, object); err != nil {
		subscribers := r.manager.EventMapping[r.gvk]
		if subscribers == nil {
			return ctrl.Result{}, nil
		}
		namespacedName := types.NamespacedName{
			Namespace: object.GetNamespace(),
			Name:      object.GetName(),
		}
		queueContexts := subscribers[namespacedName]
		for _, qc := range queueContexts {
			qc.Reconciler.Reconcile(qc.Context, qc.Req)
		}
	}
	return ctrl.Result{}, nil
}
