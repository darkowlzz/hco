package controllers

import (
	"context"
	"fmt"
	"reflect"

	controllerv1 "github.com/darkowlzz/composite-reconciler/controller/v1"
	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	darkowlzzspacev1 "github.com/darkowlzz/hco/api/v1"
	"github.com/darkowlzz/hco/controllers/cluster"
)

func (r *ClusterReconciler) InitReconcile(ctx context.Context, req ctrl.Request) {
	// Initialize cluster request.
	r.req = cluster.ClusterRequest{
		Request:  req,
		Instance: &darkowlzzspacev1.Cluster{},
		Ctx:      ctx,
		Log:      r.Log.WithValues("Cluster", req.NamespacedName),
	}
}

func (r *ClusterReconciler) FetchInstance(ctx context.Context) error {
	if err := r.Get(ctx, r.req.NamespacedName, r.req.Instance); err != nil {
		return err
	}

	// Create an owner reference from the instance.
	cluster := r.req.Instance
	r.req.OwnerRef = metav1.OwnerReference{
		APIVersion: cluster.APIVersion,
		Kind:       cluster.Kind,
		Name:       cluster.Name,
		UID:        cluster.UID,
	}

	return nil
}

func (r *ClusterReconciler) Default() {

}

func (r *ClusterReconciler) Validate() error {
	return nil
}

func (r *ClusterReconciler) SaveClone() {
	r.req.OriginalInstance = r.req.Instance.DeepCopy()
}

func (r *ClusterReconciler) IsUninitialized() bool {
	return controllerv1.DefaultIsUninitialized(r.req.Instance.Status.Conditions)
}

func (r *ClusterReconciler) Initialize(condn conditionsv1.Condition) error {
	r.UpdateConditions([]conditionsv1.Condition{condn})
	// TODO: Maybe update the object status via API.
	return nil
}

func (r *ClusterReconciler) UpdateStatus(ctx context.Context) error {
	r.Log.Info("fetching status updates", "instance", r.req.Instance.Name)

	var appReady, sidecarAReady, sidecarBReady bool

	// Get an app instance and set the status condition as ready.
	appNsName := types.NamespacedName{
		Name:      "app-" + r.req.Instance.Name,
		Namespace: r.req.Instance.Namespace,
	}
	var app darkowlzzspacev1.App
	if getErr := r.Get(ctx, appNsName, &app); getErr != nil {
		if !apierrors.IsNotFound(getErr) {
			return fmt.Errorf("failed to get app %v for Cluster status update: %w", appNsName, getErr)
		}
	} else {
		appReady = true
	}
	r.Log.Info("fetched app status", "ready", appReady)

	// Get a sidecarA instance and set the status condition as ready.
	sidecarANsName := types.NamespacedName{
		Name:      "sidecara-" + r.req.Instance.Name,
		Namespace: r.req.Instance.Namespace,
	}
	var sidecarA darkowlzzspacev1.SidecarA
	if getErr := r.Get(ctx, sidecarANsName, &sidecarA); getErr != nil {
		if !apierrors.IsNotFound(getErr) {
			return fmt.Errorf("failed to get sidecarA %v for Cluster status update: %w", sidecarANsName, getErr)
		}
	} else {
		sidecarAReady = true
	}
	r.Log.Info("fetched sidecarA status", "ready", sidecarAReady)

	// Get a sidecarB instance and set the status condition as ready.
	sidecarBNsName := types.NamespacedName{
		Name:      "sidecarb-" + r.req.Instance.Name,
		Namespace: r.req.Instance.Namespace,
	}
	var sidecarB darkowlzzspacev1.SidecarB
	if getErr := r.Get(ctx, sidecarBNsName, &sidecarB); getErr != nil {
		if !apierrors.IsNotFound(getErr) {
			return fmt.Errorf("failed to get sidecarB %v for Cluster status update: %w", sidecarBNsName, getErr)
		}
	} else {
		sidecarBReady = true
	}
	r.Log.Info("fetched sidecarB status", "ready", sidecarBReady)

	fmt.Println("APP READY:", appReady)
	fmt.Println("SIDECARA READY:", sidecarAReady)
	fmt.Println("SIDECARB READY:", sidecarBReady)

	// If all the components are ready, mark the cluster ready.
	if appReady && sidecarAReady && sidecarBReady {
		// Remove pending condition and add available condition.
		conditionsv1.RemoveStatusCondition(&r.req.Instance.Status.Conditions, conditionsv1.ConditionProgressing)
		readyCondition := conditionsv1.Condition{
			Type:    conditionsv1.ConditionAvailable,
			Status:  corev1.ConditionTrue,
			Reason:  "Ready",
			Message: "Ready",
		}
		conditionsv1.SetStatusCondition(&r.req.Instance.Status.Conditions, readyCondition)
	}

	// TODO: Add separate status conditions for all the components similar to
	// the conditions in a k8s node object.

	return nil
}

func (r *ClusterReconciler) UpdateConditions(condns []conditionsv1.Condition) {
	for _, condn := range condns {
		conditionsv1.SetStatusCondition(&r.req.Instance.Status.Conditions, condn)
	}
}

func (r *ClusterReconciler) PatchStatus(ctx context.Context) error {
	r.Log.Info("updating status", "instance", r.req.Instance.Name)
	if reflect.DeepEqual(r.req.OriginalInstance.Status, r.req.Instance.Status) {
		r.Log.Info("no diff found to update")
		return nil
	}
	return r.Status().Patch(ctx, r.req.Instance, client.MergeFrom(r.req.OriginalInstance))
}

func (r *ClusterReconciler) GetObjectMetadata() metav1.ObjectMeta {
	return r.req.Instance.ObjectMeta
}

func (r *ClusterReconciler) AddFinalizer(finalizer string) error {
	return nil
}

func (r *ClusterReconciler) Cleanup(ctx context.Context) (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

func (r *ClusterReconciler) Operate(ctx context.Context) (res ctrl.Result, err error) {
	return r.Operator.Ensure(ctx, r.req.Instance, r.req.OwnerRef)
}
