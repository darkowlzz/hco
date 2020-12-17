package cluster

import (
	"context"
	"fmt"

	eventv1 "github.com/darkowlzz/composite-reconciler/event/v1"
	"github.com/darkowlzz/composite-reconciler/operator/v1/operand"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	darkowlzzspacev1 "github.com/darkowlzz/hco/api/v1"
)

type SidecarAOperand struct {
	name            string
	client          client.Client
	requires        []string
	requeueStrategy operand.RequeueStrategy
}

// Compile-time assert to verify interface compatibility.
var _ operand.Operand = &SidecarAOperand{}

func (s *SidecarAOperand) Name() string {
	return s.name
}

func (s *SidecarAOperand) Requires() []string {
	return s.requires
}

func (s *SidecarAOperand) RequeueStrategy() operand.RequeueStrategy {
	return s.requeueStrategy
}

func (s *SidecarAOperand) Ensure(ctx context.Context, obj client.Object, ownerRef metav1.OwnerReference) (eventv1.ReconcilerEvent, error) {
	cluster, ok := obj.(*darkowlzzspacev1.Cluster)
	if !ok {
		return nil, fmt.Errorf("failed to convert %v to Cluster", obj)
	}

	sidecarAExists, hasErr := s.hasSidecarA(ctx, cluster)
	if hasErr != nil {
		return nil, hasErr
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
		if createErr := s.client.Create(ctx, sidecarA); createErr != nil {
			// r.Log.Info("failed to create sidecarA", "error", createErr)
			return nil, createErr
		}
		// Create event and requeue.
		event := &SidecarACreatedEvent{Object: cluster, SidecarAName: sidecarA.Name}
		return event, nil
	}

	return nil, nil
}

func (s *SidecarAOperand) hasSidecarA(ctx context.Context, instance *darkowlzzspacev1.Cluster) (bool, error) {
	cluster := instance
	var sidecarA darkowlzzspacev1.SidecarA
	nsName := types.NamespacedName{
		Name:      "sidecara-" + cluster.Name,
		Namespace: cluster.Namespace,
	}
	if getErr := s.client.Get(ctx, nsName, &sidecarA); getErr != nil {
		if apierrors.IsNotFound(getErr) {
			return false, nil
		}
		return false, getErr
	}
	return true, nil
}

func (s *SidecarAOperand) ReadyCheck(ctx context.Context, obj client.Object) (bool, error) {
	return true, nil
}

func (s *SidecarAOperand) Delete(ctx context.Context, obj client.Object) (eventv1.ReconcilerEvent, error) {
	return nil, nil
}

func NewSidecarAOperand(name string, client client.Client, requires []string, requeueStrategy operand.RequeueStrategy) *SidecarAOperand {
	return &SidecarAOperand{
		name:            name,
		client:          client,
		requires:        requires,
		requeueStrategy: requeueStrategy,
	}
}
