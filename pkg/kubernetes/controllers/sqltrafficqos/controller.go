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

package kubernetes

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	// "sigs.k8s.io/controller-runtime/pkg/log"
	"github.com/mlycore/log"

	"github.com/database-mesh/waterline/api/v1alpha1"
	"github.com/database-mesh/waterline/pkg/tc"
)

// SQLTrafficQoSReconciler reconciles a SQLTrafficQoS object
type SQLTrafficQoSReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=database-mesh.io.my.domain,resources=sqltrafficqos,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=database-mesh.io.my.domain,resources=sqltrafficqos/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=database-mesh.io.my.domain,resources=sqltrafficqos/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SQLTrafficQoS object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *SQLTrafficQoSReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// _ = log.FromContext(ctx)

	// TODO(user): your logic here
	obj := &v1alpha1.SQLTrafficQoS{}

	if err := r.Client.Get(ctx, req.NamespacedName, obj); err != nil {
		log.Errorf("get resources error: %s", err)
		return ctrl.Result{}, nil
	}

	// TODO: sync SQLTrafficQoSStatus
	defer func() {

	}()

	// TODO: add logic, remove VirtualDatabase.
	// Read SQLTrafficQoS for basic QoS class up.
	// Read VirtualDatabase for application-level QoS after a Pod was scheduled on this Node

	// err := r.SetTcs(ctx, obj)
	// if err != nil {
	// return ctrl.Result{Requeue: true}, nil
	// }
	log.Infof("SQLTrafficQoS: %#v", obj)

	return ctrl.Result{}, nil
}

func (r *SQLTrafficQoSReconciler) SetTcs(ctx context.Context, qos *v1alpha1.SQLTrafficQoS) error {
	cmds := []string{}
	for _, r := range qos.Spec.Groups {
		cmd := makeTcCmd(r.Parent, r.NetworkDevice, r.ClassId, r.Rate, r.Ceil)
		cmds = append(cmds, cmd)
	}

	return tc.ExecuteCmd(cmds)
}

// tc class add dev ${NIC} parent  1:0 classid 1:1 htb rate "100mbit"
// tc class add dev ${NIC} parent  1:1 classid 1:30 htb rate "500kbit" ceil "500kbit"
func makeTcCmd(networkDeviceId, parent, classId, rate, ceil string) string {
	return fmt.Sprintf("tc class add dev %s parent %s classid %s htb rate %s ceil %s", networkDeviceId, parent, classId, rate, ceil)
}

// SetupWithManager sets up the controller with the Manager.
func (r *SQLTrafficQoSReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.SQLTrafficQoS{}).
		Complete(r)
}
