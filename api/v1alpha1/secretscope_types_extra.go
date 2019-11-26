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

package v1alpha1

// SecretScopeSecret represents a secret in a secret scope
type SecretScopeSecret struct {
	Key         string                `json:"key,omitempty"`
	StringValue string                `json:"string_value,omitempty"`
	ByteValue   string                `json:"byte_value,omitempty"`
	ValueFrom   *SecretScopeValueFrom `json:"value_from,omitempty"`
}

// SecretScopeACL represents ACLs for a secret scope
type SecretScopeACL struct {
	Principal  string `json:"principal,omitempty"`
	Permission string `json:"permission,omitempty"`
}

// SecretScopeValueFrom references a secret scope
type SecretScopeValueFrom struct {
	SecretKeyRef SecretScopeKeyRef `json:"secret_key_ref,omitempty"`
}

// SecretScopeKeyRef refers to a secret scope Key
type SecretScopeKeyRef struct {
	Name string `json:"name,omitempty"`
	Key  string `json:"key,omitempty"`
}
