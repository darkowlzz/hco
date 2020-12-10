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
	"fmt"

	compositev1 "github.com/darkowlzz/composite-reconciler/controller/v1"
	operatorv1 "github.com/darkowlzz/composite-reconciler/operator/v1"
	"github.com/darkowlzz/composite-reconciler/operator/v1/executor"
	"github.com/darkowlzz/composite-reconciler/operator/v1/operand"
	"github.com/go-logr/logr"
	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	darkowlzzspacev1 "github.com/darkowlzz/hco/api/v1"
	"github.com/darkowlzz/hco/controllers/cluster"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	compositev1.CompositeReconciler
	Log    logr.Logger
	Scheme *runtime.Scheme
	// req      ClusterRequest
	req      cluster.ClusterRequest
	Recorder record.EventRecorder

	Operator operatorv1.Operator
}

// +kubebuilder:rbac:groups=darkowlzz.space,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=darkowlzz.space,resources=clusters/status,verbs=get;update;patch

// Ensure the Composite Controller interface is satisfied.
var _ compositev1.Controller = &ClusterReconciler{}

func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Setup a composite reconciler with an initial status.
	initCondn := conditionsv1.Condition{
		Type:    conditionsv1.ConditionProgressing,
		Status:  corev1.ConditionTrue,
		Reason:  "Initializing",
		Message: "Initializing",
	}
	cr, err := compositev1.NewCompositeReconciler(
		compositev1.WithCleanupStrategy(compositev1.OwnerReferenceCleanup),
		compositev1.WithLogger(r.Log.WithValues("component", "composite-reconciler")),
		compositev1.WithController(r),
		compositev1.WithInitCondition(initCondn),
	)
	if err != nil {
		return err
	}
	r.CompositeReconciler = *cr

	// Create operands for various components and add them to an operator.

	appOp := cluster.NewAppOperand("app-operand", r.Client, []string{}, operand.RequeueOnError)
	sidecarAOp := cluster.NewSidecarAOperand("sidecarA-operand", r.Client, []string{"app-operand"}, operand.RequeueOnError)
	sidecarBOp := cluster.NewSidecarBOperand("sidecarB-operand", r.Client, []string{"app-operand"}, operand.RequeueOnError)

	co, err := operatorv1.NewCompositeOperator(
		operatorv1.WithOperands(appOp, sidecarAOp, sidecarBOp),
		operatorv1.WithEventRecorder(r.Recorder),
		operatorv1.WithExecutionStrategy(executor.Serial),
	)
	if err != nil {
		return err
	}
	r.Operator = co
	fmt.Println("Operator order:", co.Order())

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
