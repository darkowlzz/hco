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

type SidecarBOperand struct {
	name            string
	client          client.Client
	requires        []string
	requeueStrategy operand.RequeueStrategy
}

// Compile-time assert to verify interface compatibility.
var _ operand.Operand = &SidecarBOperand{}

func (s *SidecarBOperand) Name() string {
	return s.name
}

func (s *SidecarBOperand) Requires() []string {
	return s.requires
}

func (s *SidecarBOperand) RequeueStrategy() operand.RequeueStrategy {
	return s.requeueStrategy
}

func (s *SidecarBOperand) Ensure(ctx context.Context, obj client.Object, ownerRef metav1.OwnerReference) (eventv1.ReconcilerEvent, error) {
	cluster, ok := obj.(*darkowlzzspacev1.Cluster)
	if !ok {
		return nil, fmt.Errorf("failed to convert %v to Cluster", obj)
	}

	sidecarBExists, hasErr := s.hasSidecarB(ctx, cluster)
	if hasErr != nil {
		return nil, hasErr
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
		if createErr := s.client.Create(ctx, sidecarB); createErr != nil {
			// r.Log.Info("failed to create sidecarB", "error", createErr)
			return nil, createErr
		}
		// Create event and requeue.
		event := &SidecarBCreatedEvent{Object: cluster, SidecarBName: sidecarB.Name}
		return event, nil
	}

	return nil, nil
}

func (s *SidecarBOperand) hasSidecarB(ctx context.Context, instance *darkowlzzspacev1.Cluster) (bool, error) {
	cluster := instance
	var sidecarB darkowlzzspacev1.SidecarB
	nsName := types.NamespacedName{
		Name:      "sidecarb-" + cluster.Name,
		Namespace: cluster.Namespace,
	}
	if getErr := s.client.Get(ctx, nsName, &sidecarB); getErr != nil {
		if apierrors.IsNotFound(getErr) {
			return false, nil
		}
		return false, getErr
	}
	return true, nil
}

func (s *SidecarBOperand) ReadyCheck(ctx context.Context, obj client.Object) (bool, error) {
	return true, nil
}

func (s *SidecarBOperand) Delete(ctx context.Context, obj client.Object) (eventv1.ReconcilerEvent, error) {
	return nil, nil
}

func NewSidecarBOperand(name string, client client.Client, requires []string, requeueStrategy operand.RequeueStrategy) *SidecarBOperand {
	return &SidecarBOperand{
		name:            name,
		client:          client,
		requires:        requires,
		requeueStrategy: requeueStrategy,
	}
}
