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
	"sync"

	"github.com/mlycore/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Client is built upon the real Kubernetes client-go
type Client struct {
	Config *rest.Config
	kubernetes.Interface
}

//DefaultClient is global Kubernetes rest client
var DefaultClient *Client
var once sync.Once

func NewClientInCluster() *Client {
	once.Do(func() {
		config, err := rest.InClusterConfig()
		if err != nil {
			log.Fatalf("read incluster config error: %s", err)
		}
		// creates the clientset
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			log.Fatalf("new client from incluster config error: %s", err)
		}
		DefaultClient = &Client{
			Config:    config,
			Interface: clientset,
		}
	})
	return DefaultClient
}
