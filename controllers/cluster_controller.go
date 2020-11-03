/*


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
	compositev1 "github.com/darkowlzz/composite-reconciler/controller/v1"
	"github.com/go-logr/logr"
	conditions "github.com/openshift/custom-resource-status/conditions/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	darkowlzzspacev1 "github.com/darkowlzz/hco/api/v1"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	compositev1.CompositeReconciler
	Log    logr.Logger
	Scheme *runtime.Scheme
	req    ClusterRequest
}

// +kubebuilder:rbac:groups=darkowlzz.space,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=darkowlzz.space,resources=clusters/status,verbs=get;update;patch

// Ensure the Composite Controller interface is satisfied.
var _ compositev1.Controller = &ClusterReconciler{}

func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.CompositeReconciler = compositev1.CompositeReconciler{
		Log: r.Log.WithValues("component", "composite-reconciler"),
		C:   r,
		InitCondition: conditions.Condition{
			Type:    conditions.ConditionAvailable,
			Status:  corev1.ConditionTrue,
			Reason:  "Initialized",
			Message: "Initialized",
		},
		FinalizerName: "cluster-controller-finalizer",
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&darkowlzzspacev1.Cluster{}).
		WithEventFilter(ChangePredicate{}).
		Complete(r)
}

type ChangePredicate struct {
	predicate.Funcs
}

func (ChangePredicate) Update(e event.UpdateEvent) bool {
	if e.MetaOld == nil || e.MetaNew == nil {
		// ignore objects without metadata
		return false
	}

	if e.MetaNew.GetGeneration() != e.MetaOld.GetGeneration() {
		// reconcile on spec changes
		return true
	}

	// TODO: Provide option to handle forced sync using annotations.

	return false
}
