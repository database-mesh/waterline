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

package server

import (
	"github.com/database-mesh/waterline/pkg/cri"
	kube "github.com/database-mesh/waterline/pkg/kubernetes"
	"github.com/database-mesh/waterline/pkg/server/config"
	"k8s.io/client-go/kubernetes"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	KubernetesClient       kubernetes.Interface
	ContainerRuntimeClient cri.ContainerRuntimeInterfaceClient
}

func New(conf *config.Config) (*Server, error) {
	cr, err := cri.NewContainerRuntimeInterfaceClient(conf.CRI)
	if err != nil {
		return nil, err
	}

	kc, err := kube.NewClientInCluster()
	if err != nil {
		return nil, err
	}

	return &Server{
		ContainerRuntimeClient: cr
		KubernetesClient: kc,
		}, nil
}

func (s Server) Run() {
	var eg errgroup.Group
	eg.Go(func() error {
		log.Infof("starting watching kubernetes")
		// TODO: add kubernetes crd watching
		return nil
	})
	return eg.Wait()
}
