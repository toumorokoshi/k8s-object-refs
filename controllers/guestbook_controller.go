/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	webappv1 "github.com/toumorokoshi/k8s-object-refs/api/v1"
	"github.com/toumorokoshi/k8s-object-refs/pkg/refs"
)

// GuestbookReconciler reconciles a Guestbook object
type GuestbookReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	RefManager refs.RefManager
}

//+kubebuilder:rbac:groups=webapp.tsutsumi.io,resources=guestbooks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=webapp.tsutsumi.io,resources=guestbooks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=webapp.tsutsumi.io,resources=guestbooks/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Guestbook object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *GuestbookReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	var guestBook webappv1.Guestbook
	if err := r.Get(ctx, req.NamespacedName, &guestBook); err != nil {
		logger.Error(err, "unable to fetch GuestBook")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	nn := types.NamespacedName{
		Name:      guestBook.Spec.FooRef.Name,
		Namespace: guestBook.Spec.FooRef.Namespace,
	}

	qc := refs.QueueContext{
		Context:    ctx,
		Req:        req,
		Reconciler: r,
	}

	r.RefManager.UpdateSubscriptions(refs.GVK{
		Group:   "core",
		Version: "v1",
		Kind:    "Pod",
	}, nn, qc)

	logger.Info("nn")
	// r.RefManager

	// CUSTOM: update entry

	// queue := workqueue.NewNamedRateLimitingQueue(
	// 	workqueue.DefaultControllerRateLimiter(),
	// 	"foo",
	// )
	// src := source.Kind{Type: &v1.Pod{}}
	// src.Start(ctx, &HandleUpdateEvent{ctx: ctx}, queue)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GuestbookReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1.Guestbook{}).Complete(r)
}
