package controllers

import (
	"context"
	"reflect"

	controllerv1 "github.com/darkowlzz/composite-reconciler/controller/v1"
	eventv1 "github.com/darkowlzz/composite-reconciler/event/v1"
	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	darkowlzzspacev1 "github.com/darkowlzz/hco/api/v1"
)

func (r *ClusterReconciler) InitReconcile(ctx context.Context, req ctrl.Request) {
	// Initialize cluster request.
	r.req = ClusterRequest{
		Request:  req,
		Instance: &darkowlzzspacev1.Cluster{},
		Ctx:      ctx,
		Log:      r.Log.WithValues("Cluster", req.NamespacedName),
	}
}

func (r *ClusterReconciler) FetchInstance() error {
	if err := r.Get(r.req.Ctx, r.req.NamespacedName, r.req.Instance); err != nil {
		return err
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
	return nil
}

func (r *ClusterReconciler) FetchStatus() error {
	// Remove pending condition and add available condition.
	conditionsv1.RemoveStatusCondition(&r.req.Instance.Status.Conditions, conditionsv1.ConditionProgressing)
	readyCondition := conditionsv1.Condition{
		Type:    conditionsv1.ConditionAvailable,
		Status:  corev1.ConditionTrue,
		Reason:  "Ready",
		Message: "Ready",
	}
	conditionsv1.SetStatusCondition(&r.req.Instance.Status.Conditions, readyCondition)
	return nil
}

func (r *ClusterReconciler) UpdateConditions(condns []conditionsv1.Condition) {
	for _, condn := range condns {
		conditionsv1.SetStatusCondition(&r.req.Instance.Status.Conditions, condn)
	}
}

func (r *ClusterReconciler) UpdateStatus() error {
	r.Log.Info("updating status")
	if reflect.DeepEqual(r.req.OriginalInstance, r.req.Instance) {
		r.Log.Info("no diff found to update")
		return nil
	}
	return r.Status().Update(r.req.Ctx, r.req.Instance)
}

func (r *ClusterReconciler) GetObjectMetadata() metav1.ObjectMeta {
	return r.req.Instance.ObjectMeta
}

func (r *ClusterReconciler) AddFinalizer(finalizer string) error {
	return nil
}

func (r *ClusterReconciler) Cleanup() error {
	return nil
}

func (r *ClusterReconciler) hasApp() (bool, error) {
	cluster := r.req.Instance
	var app darkowlzzspacev1.App
	nsName := types.NamespacedName{
		Name:      "app-" + cluster.Name,
		Namespace: cluster.Namespace,
	}
	if getErr := r.Get(r.req.Ctx, nsName, &app); getErr != nil {
		if apierrors.IsNotFound(getErr) {
			return false, nil
		}
		return false, getErr
	}
	return true, nil
}

func (r *ClusterReconciler) hasSidecarA() (bool, error) {
	cluster := r.req.Instance
	var sidecarA darkowlzzspacev1.SidecarA
	nsName := types.NamespacedName{
		Name:      "sidecara-" + cluster.Name,
		Namespace: cluster.Namespace,
	}
	if getErr := r.Get(r.req.Ctx, nsName, &sidecarA); getErr != nil {
		if apierrors.IsNotFound(getErr) {
			return false, nil
		}
		return false, getErr
	}
	return true, nil
}

func (r *ClusterReconciler) hasSidecarB() (bool, error) {
	cluster := r.req.Instance
	var sidecarB darkowlzzspacev1.SidecarB
	nsName := types.NamespacedName{
		Name:      "sidecarb-" + cluster.Name,
		Namespace: cluster.Namespace,
	}
	if getErr := r.Get(r.req.Ctx, nsName, &sidecarB); getErr != nil {
		if apierrors.IsNotFound(getErr) {
			return false, nil
		}
		return false, getErr
	}
	return true, nil
}

func (r *ClusterReconciler) Operate() (res ctrl.Result, event eventv1.ReconcilerEvent, err error) {
	res = ctrl.Result{}
	event = nil
	err = nil

	cluster := r.req.Instance

	// Create an owner reference using the cluster.
	ownerRef := metav1.OwnerReference{
		APIVersion: cluster.APIVersion,
		Kind:       cluster.Kind,
		Name:       cluster.Name,
		UID:        cluster.UID,
	}

	appExists, hasErr := r.hasApp()
	if hasErr != nil {
		err = hasErr
		return
	}

	if !appExists {
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
		if createErr := r.Create(r.req.Ctx, appInstance); createErr != nil {
			r.Log.Info("failed to create app", "error", createErr)
			err = createErr
			return
		}
		// Create event and requeue.
		event = &AppCreatedEvent{Object: cluster, AppName: appInstance.Name}
		res = ctrl.Result{Requeue: true}
		return
	}

	sidecarAExists, hasErr := r.hasSidecarA()
	if hasErr != nil {
		err = hasErr
		return
	}

	if !sidecarAExists {
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
		if createErr := r.Create(r.req.Ctx, sidecarA); createErr != nil {
			r.Log.Info("failed to create sidecarA", "error", createErr)
			err = createErr
		}
		// Create event and requeue.
		event = &SidecarACreatedEvent{Object: cluster, SidecarAName: sidecarA.Name}
		res = ctrl.Result{Requeue: true}
		return
	}

	sidecarBExists, hasErr := r.hasSidecarB()
	if hasErr != nil {
		err = hasErr
		return
	}

	if !sidecarBExists {
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
		if createErr := r.Create(r.req.Ctx, sidecarB); createErr != nil {
			r.Log.Info("failed to create sidecarB", "error", createErr)
			err = createErr
		}
		// Create event and requeue.
		event = &SidecarBCreatedEvent{Object: cluster, SidecarBName: sidecarB.Name}
		res = ctrl.Result{Requeue: true}
		return
	}

	return
}
