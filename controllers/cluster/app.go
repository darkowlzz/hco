package cluster

import (
	"context"
	"fmt"

	eventv1 "github.com/darkowlzz/composite-reconciler/event/v1"
	"github.com/darkowlzz/composite-reconciler/operator/v1/operand"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	darkowlzzspacev1 "github.com/darkowlzz/hco/api/v1"
)

type AppOperand struct {
	name            string
	client          client.Client
	requires        []string
	requeueStrategy operand.RequeueStrategy
}

func (a *AppOperand) Name() string {
	return a.name
}

func (a *AppOperand) Requires() []string {
	return a.requires
}

func (a *AppOperand) RequeueStrategy() operand.RequeueStrategy {
	return a.requeueStrategy
}

func (a *AppOperand) Ensure(ctx context.Context, obj runtime.Object, ownerRef metav1.OwnerReference) (eventv1.ReconcilerEvent, error) {
	cluster, ok := obj.(*darkowlzzspacev1.Cluster)
	if !ok {
		return nil, fmt.Errorf("failed to convert %v to Cluster", obj)
	}

	// Check if app already exists.
	appExists, hasErr := a.hasApp(ctx, cluster)
	if hasErr != nil {
		return nil, hasErr
	}

	if !appExists {
		// Create App with cluster as the owner reference.
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
		if createErr := a.client.Create(ctx, appInstance); createErr != nil {
			// r.Log.Info("failed to create app", "error", createErr)
			return nil, createErr
		}
		// Create event and requeue.
		event := &AppCreatedEvent{Object: cluster, AppName: appInstance.Name}
		return event, nil
	}

	return nil, nil
}

func (a *AppOperand) hasApp(ctx context.Context, instance *darkowlzzspacev1.Cluster) (bool, error) {
	cluster := instance
	var app darkowlzzspacev1.App
	nsName := types.NamespacedName{
		Name:      "app-" + cluster.Name,
		Namespace: cluster.Namespace,
	}
	if getErr := a.client.Get(ctx, nsName, &app); getErr != nil {
		if apierrors.IsNotFound(getErr) {
			return false, nil
		}
		return false, getErr
	}
	return true, nil
}

func (a *AppOperand) ReadyCheck(ctx context.Context, obj runtime.Object) (bool, error) {
	return true, nil
}

func (a *AppOperand) Delete(ctx context.Context, obj runtime.Object) (eventv1.ReconcilerEvent, error) {
	return nil, nil
}

func NewAppOperand(name string, client client.Client, requires []string, requeueStrategy operand.RequeueStrategy) *AppOperand {
	return &AppOperand{
		name:            name,
		client:          client,
		requires:        requires,
		requeueStrategy: requeueStrategy,
	}
}
