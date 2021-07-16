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

	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	webappv1 "github.com/toumorokoshi/k8s-object-refs/api/v1"
	"github.com/toumorokoshi/k8s-object-refs/pkg/refs"
)

var HOTEL_REF = refs.GVK{
	Group:   "webapp.tsutsumi.io",
	Version: "v1",
	Kind:    "Hotel",
}

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
	qc := refs.QueueContext{
		Context:    ctx,
		Req:        req,
		Reconciler: r,
	}
	var guestbook webappv1.Guestbook
	if err := r.Get(ctx, req.NamespacedName, &guestbook); err != nil {
		// If we're unable to retrieve the object, clear the
		// subscriptions.
		r.RefManager.UpdateSubscriptions(qc, []refs.RefSubscription{})
		logger.Error(err, "unable to fetch GuestBook")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// ref checking logic happens after subscriptions are updated,
	// to ensure that subscriptions are up-to-date even if the
	// target object is not yet available.
	nn := types.NamespacedName{
		Name:      guestbook.Spec.FooRef.Name,
		Namespace: guestbook.Spec.FooRef.Namespace,
	}
	r.RefManager.UpdateSubscriptions(qc, []refs.RefSubscription{
		{Gvk: HOTEL_REF, NamespacedName: nn},
	})

	var hotel webappv1.Hotel
	if err := r.Get(ctx, nn, &hotel); err != nil {
		logger.Error(err, "unable to find dependent hotel.")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	readyStatus := v1.ConditionFalse
	readyReason := "DependencyNotReady"

	// TODO: enable using conditions. This is not possible due to the
	// cached object not having the status updates from the other controller.
	// see https://github.com/kubernetes-sigs/controller-runtime/issues/585 for a related issue.
	// hotelReadyCondition := meta.FindStatusCondition(hotel.Status.Conditions, "Ready")
	// if hotelReadyCondition != nil && hotelReadyCondition.Status == v1.ConditionTrue {
	if hotel.Spec.Ok {
		logger.Info("hotel is ready")
		readyStatus = v1.ConditionTrue
		readyReason = "DependencyReady"
	}

	meta.SetStatusCondition(&guestbook.Status.Conditions, v1.Condition{
		Type:   "Ready",
		Status: readyStatus,
		Reason: readyReason,
	})

	if err := r.Status().Update(ctx, &guestbook); err != nil {
		logger.Error(err, "unable to update guestbook")
		return ctrl.Result{}, err
	}

	logger.Info("end of controller")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GuestbookReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1.Guestbook{}).Complete(r)
}
