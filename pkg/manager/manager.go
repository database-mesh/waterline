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

package manager

import (
	sqltrafficqos "github.com/database-mesh/waterline/pkg/kubernetes/controllers/sqltrafficqos"
	virtualdatabase "github.com/database-mesh/waterline/pkg/kubernetes/controllers/virtualdatabase"
	"github.com/database-mesh/waterline/pkg/kubernetes/watcher"
	"github.com/mlycore/log"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	ctrlmgr "sigs.k8s.io/controller-runtime/pkg/manager"
)

type Manager struct {
	Pod *watcher.PodWatcher
	Mgr ctrlmgr.Manager
}

func (m *Manager) WatchAndHandle() error {
	for {
		select {
		case event := <-m.Pod.Core.ResultChan():
			{
				log.Infof("[%s] pod event: %#v", event.Type, event.Object.(*corev1.Pod).Name)
				//TODO: Handle different types of events
			}
		}
	}
	return nil
}

func (m *Manager) Bootstrap() error {
	if err := (&sqltrafficqos.SQLTrafficQoSReconciler{
		Client: m.Mgr.GetClient(),
		Scheme: m.Mgr.GetScheme(),
	}).SetupWithManager(m.Mgr); err != nil {
		log.Errorf("sqltrafficqos setupWithManager error: %s", err)
		return err
	}

	if err := (&virtualdatabase.VirtualDatabaseReconciler{
		Client: m.Mgr.GetClient(),
		Scheme: m.Mgr.GetScheme(),
	}).SetupWithManager(m.Mgr); err != nil {
		log.Errorf("virtualdatabase setupWithManager error: %s", err)
		return err
	}

	if err := m.Mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return err
	}
	if err := m.Mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return err
	}

	if err := m.Mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return err
	}
	return nil
}
