/*
The MIT License (MIT)

Copyright (c) 2019  Microsoft

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
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
