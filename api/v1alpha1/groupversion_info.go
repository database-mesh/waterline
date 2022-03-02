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

|
import (                                                                  | func init() {
	"k8s.io/apimachinery/pkg/runtime/schema"                              |     utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	"sigs.k8s.io/controller-runtime/pkg/scheme"                           |
)                                                                         |     utilruntime.Must(databasemeshiov1alpha1.AddToScheme(scheme))
																		  |     //+kubebuilder:scaffold:scheme
var (                                                                     | }
	// GroupVersion is group version used to register these objects       |
	GroupVersion = schema.GroupVersion{Group: "database-mesh.io",         | func main() {
Version: "v1alpha1"}                                                      |     var metricsAddr string
																		  |     var enableLeaderElection bool
	// SchemeBuilder is used to add go types to the GroupVersionKind      |     var probeAddr string
scheme                                                                    |     flag.StringVar(&metricsAddr, "metrics-bind-address", ":8180", "The
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}           | address the metric endpoint binds to.")
																		  |     flag.StringVar(&probeAddr, "health-probe-bind-address", ":8181", "The
	// AddToScheme adds the types in this group-version to the given      | address the probe endpoint binds to.")
scheme.                                                                   |     flag.BoolVar(&enableLeaderElection, "leader-elect", false,
	AddToScheme = SchemeBuilder.AddToScheme                               |         "Enable leader election for controller manager. "+
)
