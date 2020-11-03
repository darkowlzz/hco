package controllers

import (
	"context"

	darkowlzzspacev1 "github.com/darkowlzz/hco/api/v1"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ClusterRequest contains infomation related to the current cluster
// reconciliation request.
type ClusterRequest struct {
	reconcile.Request
	Instance *darkowlzzspacev1.Cluster
	Ctx      context.Context
	Log      logr.Logger
}
