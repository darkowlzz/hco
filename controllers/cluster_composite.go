package controllers

import (
	"context"

	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	darkowlzzspacev1 "github.com/darkowlzz/hco/api/v1"
	cc "github.com/darkowlzz/hco/compositecontroller"
)

func (r *ClusterReconciler) InitReconcile(ctx context.Context, req reconcile.Request) {
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

func (r *ClusterReconciler) IsUninitialized() bool {
	return cc.DefaultIsUninitialized(r.req.Instance.Status.Conditions)
}

func (r *ClusterReconciler) Initialize(condn conditionsv1.Condition) error {
	// if err := r.RunOperation(); err != nil {
	if err := r.RunInit(); err != nil {
		return err
	}
	if err := r.UpdateConditions([]conditionsv1.Condition{condn}); err != nil {
		return err
	}
	return nil
}

func (r *ClusterReconciler) RunInit() error {
	cluster := r.req.Instance

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
	if err := r.Create(r.req.Ctx, appInstance); err != nil {
		if apierrors.IsAlreadyExists(err) {
			nsName := types.NamespacedName{
				Name:      appInstance.Name,
				Namespace: appInstance.Namespace,
			}
			if err := r.Get(r.req.Ctx, nsName, appInstance); err != nil {
				r.Log.Info("failed to get app", "error", err)
				return err
			}
			if updateErr := r.Update(r.req.Ctx, appInstance); updateErr != nil {
				r.Log.Info("failed to update app", "error", err)
			}
		} else {
			r.Log.Info("failed to create app", "error", err)
		}
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
	if err := r.Create(r.req.Ctx, sidecarA); err != nil {
		if apierrors.IsAlreadyExists(err) {
			if updateErr := r.Update(r.req.Ctx, sidecarA); updateErr != nil {
				r.Log.Info("failed to update sidecarA", "error", err)
			}
		} else {
			r.Log.Info("failed to create sidecarA", "error", err)
		}
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
	if err := r.Create(r.req.Ctx, sidecarB); err != nil {
		if apierrors.IsAlreadyExists(err) {
			if updateErr := r.Update(r.req.Ctx, sidecarB); updateErr != nil {
				r.Log.Info("failed to update sidecarB", "error", err)
			}
		} else {
			r.Log.Info("failed to create sidecarB", "error", err)
		}
	}

	return nil
}

func (r *ClusterReconciler) UpdateConditions(condns []conditionsv1.Condition) error {
	for _, condn := range condns {
		conditionsv1.SetStatusCondition(&r.req.Instance.Status.Conditions, condn)
	}

	// Update the object in API.
	if err := r.Status().Update(r.req.Ctx, r.req.Instance); err != nil {
		return err
	}

	return nil
}

func (r *ClusterReconciler) StatusUpdate() error {
	return nil
}

func (r *ClusterReconciler) EmitEvent(eventType, reason, message string) {
}

func (r *ClusterReconciler) CheckForDrift() (bool, error) {
	cluster := r.req.Instance

	// Check if the app config drifted.
	var app darkowlzzspacev1.App
	nsName := types.NamespacedName{
		Name:      "app-" + cluster.Name,
		Namespace: cluster.Namespace,
	}
	if err := r.Get(r.req.Ctx, nsName, &app); err != nil {
		if apierrors.IsNotFound(err) {
			// Resource doesn't exist. Drift.
			return true, nil
		}
		return false, err
	}
	if app.Spec.Image != cluster.Spec.Images.App {
		return true, nil
	}

	var sidecarA darkowlzzspacev1.SidecarA
	nsName = types.NamespacedName{
		Name:      "sidecara-" + cluster.Name,
		Namespace: cluster.Namespace,
	}
	if err := r.Get(r.req.Ctx, nsName, &sidecarA); err != nil {
		if apierrors.IsNotFound(err) {
			// Resource doesn't exist. Drift.
			return true, nil
		}
		return false, err
	}
	if sidecarA.Spec.Image != cluster.Spec.Images.SidecarA {
		return true, nil
	}

	var sidecarB darkowlzzspacev1.SidecarB
	nsName = types.NamespacedName{
		Name:      "sidecarb-" + cluster.Name,
		Namespace: cluster.Namespace,
	}
	if err := r.Get(r.req.Ctx, nsName, &sidecarB); err != nil {
		if apierrors.IsNotFound(err) {
			// Resource doesn't exist. Drift.
			return true, nil
		}
		return false, err
	}
	if sidecarB.Spec.Image != cluster.Spec.Images.SidecarB {
		return true, nil
	}

	return false, nil
}

func (r *ClusterReconciler) OwnedResourceIsUpdatable() bool {
	return true
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

func (r *ClusterReconciler) RunOperation() error {
	cluster := r.req.Instance

	var app darkowlzzspacev1.App
	nsName := types.NamespacedName{
		Name:      "app-" + cluster.Name,
		Namespace: cluster.Namespace,
	}
	if err := r.Get(r.req.Ctx, nsName, &app); err != nil {
		r.Log.Info("failed to get app", "error", err)
		return err
	}
	app.Spec.Image = cluster.Spec.Images.App
	if err := r.Update(r.req.Ctx, &app); err != nil {
		r.Log.Info("failed to update app", "error", err)
	}

	var sidecarA darkowlzzspacev1.SidecarA
	nsName.Name = "sidecara-" + cluster.Name
	if err := r.Get(r.req.Ctx, nsName, &sidecarA); err != nil {
		r.Log.Info("failed to get sidecarA", "error", err)
		return err
	}
	sidecarA.Spec.Image = cluster.Spec.Images.SidecarA
	if err := r.Update(r.req.Ctx, &sidecarA); err != nil {
		r.Log.Info("failed to update sidecarA", "error", err)
	}

	var sidecarB darkowlzzspacev1.SidecarB
	nsName.Name = "sidecarb-" + cluster.Name
	if err := r.Get(r.req.Ctx, nsName, &sidecarB); err != nil {
		r.Log.Info("failed to get sidecarB", "error", err)
		return err
	}
	sidecarB.Spec.Image = cluster.Spec.Images.SidecarB
	if err := r.Update(r.req.Ctx, &sidecarB); err != nil {
		r.Log.Info("failed to update sidecarB", "error", err)
	}

	return nil
}
