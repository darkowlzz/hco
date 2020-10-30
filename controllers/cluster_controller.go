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
	"context"

	"github.com/go-logr/logr"
	conditions "github.com/openshift/custom-resource-status/conditions/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	darkowlzzspacev1 "github.com/darkowlzz/hco/api/v1"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=darkowlzz.space,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=darkowlzz.space,resources=clusters/status,verbs=get;update;patch

func (r *ClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("cluster", req.NamespacedName)

	// your logic here
	log.Info("Cluster reconciling...")

	var cluster darkowlzzspacev1.Cluster
	if err := r.Get(ctx, req.NamespacedName, &cluster); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	init := true
	if conditions.IsStatusConditionTrue(cluster.Status.Conditions, conditions.ConditionAvailable) {
		init = false
	}

	// Initialize the cluster if uninitialized.
	if init {
		// Create an owner reference using the cluster.
		ownerRef := metav1.OwnerReference{
			APIVersion: cluster.APIVersion,
			Kind:       cluster.Kind,
			Name:       cluster.Name,
			UID:        cluster.UID,
		}

		// Create App, SidecarA and SidecarB with cluster as the owner
		// reference.
		appInstance := &darkowlzzspacev1.App{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "app-" + cluster.Name,
				Labels:          cluster.Labels,
				Namespace:       cluster.Namespace,
				OwnerReferences: []metav1.OwnerReference{ownerRef},
			},
			Spec: darkowlzzspacev1.AppSpec{
				Image: cluster.Spec.Images.App,
			},
		}
		if err := r.Create(ctx, appInstance); err != nil {
			log.Info("failed to create app", "error", err)
		}

		sidecarA := &darkowlzzspacev1.SidecarA{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "sidecara-" + cluster.Name,
				Labels:          cluster.Labels,
				Namespace:       cluster.Namespace,
				OwnerReferences: []metav1.OwnerReference{ownerRef},
			},
			Spec: darkowlzzspacev1.SidecarASpec{
				Image: cluster.Spec.Images.SidecarA,
			},
		}
		if err := r.Create(ctx, sidecarA); err != nil {
			log.Info("failed to create sidecarA", "error", err)
		}

		sidecarB := &darkowlzzspacev1.SidecarB{
			ObjectMeta: metav1.ObjectMeta{
				Name:            "sidecarb-" + cluster.Name,
				Labels:          cluster.Labels,
				Namespace:       cluster.Namespace,
				OwnerReferences: []metav1.OwnerReference{ownerRef},
			},
			Spec: darkowlzzspacev1.SidecarBSpec{
				Image: cluster.Spec.Images.SidecarB,
			},
		}
		if err := r.Create(ctx, sidecarB); err != nil {
			log.Info("failed to create sidecarB", "error", err)
		}

		conditions.SetStatusCondition(&cluster.Status.Conditions, conditions.Condition{
			Type:    conditions.ConditionAvailable,
			Status:  corev1.ConditionTrue,
			Reason:  "Initialized",
			Message: "Initialized",
		})
		if err := r.Status().Update(ctx, &cluster); err != nil {
			return ctrl.Result{Requeue: true}, err
		}
		log.Info("Cluster initialised")
	}

	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&darkowlzzspacev1.Cluster{}).
		Complete(r)
}
