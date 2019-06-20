/*
Copyright 2019 microsoft.

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

package v1beta1

func (nj *NotebookJob) LoadDefaultConfig() *NotebookJob {
	if nj.Spec.ClusterSpec.NodeTypeId == "" {
		nj.Spec.ClusterSpec.NodeTypeId = "Standard_DS3_v2"
	}
	if nj.Spec.ClusterSpec.SparkVersion == "" {
		nj.Spec.ClusterSpec.SparkVersion = "5.2.x-scala2.11"
	}
	if nj.Spec.ClusterSpec.NumWorkers == 0 {
		nj.Spec.ClusterSpec.NumWorkers = 3
	}
	return nj
}
