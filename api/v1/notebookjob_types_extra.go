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

package v1

type NotebookTask struct {
	NotebookPath string `json:"notebookPath,omitempty"`
}

type NotebookStream struct {
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
}

type KeyMapping struct {
	SecretKey string `json:"secretKey"`
	OutputKey string `json:"outputKey"`
}
type NotebookSpecSecret struct {
	SecretName string       `json:"secretName"`
	Mapping    []KeyMapping `json:"mapping"`
}

type ClusterSpec struct {
	SparkVersion string `json:"sparkVersion,omitempty"`
	NodeTypeId   string `json:"nodeTypeId,omitempty"`
	NumWorkers   int    `json:"numWorkers,omitempty"`
}

type NotebookAdditionalLibrary struct {
	Type       string            `json:"type"`
	Properties map[string]string `json:"properties"`
}
