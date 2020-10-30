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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	darkowlzzspacev1 "github.com/darkowlzz/hco/api/v1"
)

// SidecarBReconciler reconciles a SidecarB object
type SidecarBReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=darkowlzz.space,resources=sidecarbs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=darkowlzz.space,resources=sidecarbs/status,verbs=get;update;patch

func (r *SidecarBReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("sidecarb", req.NamespacedName)

	// your logic here

	return ctrl.Result{}, nil
}

func (r *SidecarBReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&darkowlzzspacev1.SidecarB{}).
		Complete(r)
}
