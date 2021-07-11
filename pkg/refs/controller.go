package refs

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type RefReconciler struct {
	client.Client
	manager *RefManager
	gvk     GVK
}

func (r *RefReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info(fmt.Sprintf("eventMapping: %v", r.manager.EventMapping))
	subscribers := r.manager.EventMapping[r.gvk]
	if subscribers == nil {
		return ctrl.Result{}, nil
	}
	queueContexts := subscribers[req.NamespacedName]
	for _, qc := range queueContexts {
		logger.Info("enqueuing reconcile")
		qc.Reconciler.Reconcile(qc.Context, qc.Req)
	}
	return ctrl.Result{}, nil
}
