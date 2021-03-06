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
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	webappv1 "github.com/toumorokoshi/k8s-object-refs/api/v1"
)

// HotelReconciler reconciles a Hotel object
type HotelReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=webapp.tsutsumi.io,resources=hotels,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=webapp.tsutsumi.io,resources=hotels/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=webapp.tsutsumi.io,resources=hotels/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Hotel object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *HotelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var hotel webappv1.Hotel
	if err := r.Get(ctx, req.NamespacedName, &hotel); err != nil {
		logger.Error(err, "unable to fetch Hotel")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	readyStatus := v1.ConditionFalse
	readyReason := "OkSetToFalse"
	if hotel.Spec.Ok {
		readyStatus = v1.ConditionTrue
		readyReason = "OkSetToTrue"
	}

	meta.SetStatusCondition(&hotel.Status.Conditions, v1.Condition{
		Type:   "Ready",
		Status: readyStatus,
		Reason: readyReason,
	})

	if err := r.Status().Update(ctx, &hotel); err != nil {
		logger.Error(err, "unable to update Hotel")
		return ctrl.Result{}, err
	}

	logger.Info(fmt.Sprintf("set status to %v", readyStatus))

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HotelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1.Hotel{}).
		Complete(r)
}
