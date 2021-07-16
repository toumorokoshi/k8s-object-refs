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
	curElement := subscribers[req.NamespacedName]
	if curElement == nil {
		return ctrl.Result{}, nil
	}
	// iterate past the head, which is a dummy element
	curElement = curElement.nextInEventMap
	for curElement != nil {
		logger.Info("enqueuing reconcile")
		// TODO: error handling
		qc := curElement.queueContext
		qc.Reconciler.Reconcile(qc.Context, qc.Req)
		curElement = curElement.nextInEventMap
	}
	return ctrl.Result{}, nil
}
