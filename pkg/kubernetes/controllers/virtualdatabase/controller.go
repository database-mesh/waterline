// Copyright 2022 Database Mesh Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controllers

import (
	"context"

	"github.com/mlycore/log"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/database-mesh/waterline/api/v1alpha1"
)

// VirtualDatabaseReconciler reconciles a VirtualDatabase object
type VirtualDatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=database-mesh.io.my.domain,resources=virtualdatabases,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=database-mesh.io.my.domain,resources=virtualdatabases/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=database-mesh.io.my.domain,resources=virtualdatabases/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the VirtualDatabase object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *VirtualDatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// TODO(user): your logic here
	obj := &v1alpha1.VirtualDatabase{}

	if err := r.Client.Get(ctx, req.NamespacedName, obj); err != nil {
		log.Errorf("get resources error: %s", err)
		return ctrl.Result{}, nil
	}

	defer func() {

	}()

	// pod := &corev1.Pod{}
	// err := r.Client.Get(ctx, types.NamespacedName{
	// 	Name:      obj.Name,
	// 	Namespace: obj.Namespace,
	// }, pod)
	// if err != nil {
	// 	log.Errorf("get pod error: %s", err)
	// 	return ctrl.Result{}, err
	// }
	log.Infof("VirtualDatabase: %#v", obj)

	// err = r.SetQoSClass(ctx, pod, object)

	return ctrl.Result{}, nil
}

// func (r *VirtualDatabaseReconciler) SetQoSClass(ctx context.Context, pod corev1.Pod) {
// 	for _, c := range object.QoS {

// 	}
// }

func makeQoSCmd() []string {
	return []string{}
}

// SetupWithManager sets up the controller with the Manager.
func (r *VirtualDatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.VirtualDatabase{}).
		Complete(r)
}
