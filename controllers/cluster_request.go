package controllers

import (
	"context"
	"fmt"

	eventv1 "github.com/darkowlzz/composite-reconciler/event/v1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	darkowlzzspacev1 "github.com/darkowlzz/hco/api/v1"
)

// ClusterRequest contains infomation related to the current cluster
// reconciliation request.
type ClusterRequest struct {
	reconcile.Request
	Instance         *darkowlzzspacev1.Cluster
	OriginalInstance *darkowlzzspacev1.Cluster
	Ctx              context.Context
	Log              logr.Logger
}

type ClusterReadyEvent struct {
	Object      runtime.Object
	ClusterName string
}

func (c *ClusterReadyEvent) Record(recorder record.EventRecorder) {
	recorder.Event(c.Object,
		eventv1.K8sEventTypeNormal,
		"ClusterReady",
		fmt.Sprintf("Cluster %s ready", c.ClusterName),
	)
}

type AppCreatedEvent struct {
	Object  runtime.Object
	AppName string
}

func (c *AppCreatedEvent) Record(recorder record.EventRecorder) {
	recorder.Event(c.Object,
		eventv1.K8sEventTypeNormal,
		"AppReady",
		fmt.Sprintf("Created App with name %s", c.AppName),
	)
}

type SidecarACreatedEvent struct {
	Object       runtime.Object
	SidecarAName string
}

func (c *SidecarACreatedEvent) Record(recorder record.EventRecorder) {
	recorder.Event(c.Object,
		eventv1.K8sEventTypeNormal,
		"SidecarAReady",
		fmt.Sprintf("Created SidecarA with name %s", c.SidecarAName),
	)
}

type SidecarBCreatedEvent struct {
	Object       runtime.Object
	SidecarBName string
}

func (c *SidecarBCreatedEvent) Record(recorder record.EventRecorder) {
	recorder.Event(c.Object,
		eventv1.K8sEventTypeNormal,
		"SidecarBReady",
		fmt.Sprintf("Created SidecarB with name %s", c.SidecarBName),
	)
}
