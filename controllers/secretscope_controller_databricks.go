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

package controllers

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	databricksv1alpha1 "github.com/microsoft/azure-databricks-operator/api/v1alpha1"
	dbmodels "github.com/xinsnake/databricks-sdk-golang/azure/models"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *SecretScopeReconciler) get(scope string) (*dbmodels.SecretScope, error) {
	execution := NewExecution("secretscopes", "list_secret_scops")
	scopes, err := r.APIClient.Secrets().ListSecretScopes()
	execution.Finish(err)
	if err != nil {
		return nil, err
	}

	matchingScope := dbmodels.SecretScope{}
	for _, existingScope := range scopes {
		if existingScope.Name == scope {
			matchingScope = existingScope
		}
	}

	if (dbmodels.SecretScope{}) == matchingScope {
		return nil, fmt.Errorf("get for secret scope failed. scope not found: %s", scope)
	}

	return &matchingScope, nil
}

func (r *SecretScopeReconciler) submitSecrets(instance *databricksv1alpha1.SecretScope) error {
	scope := instance.ObjectMeta.Name
	namespace := instance.Namespace
	execution := NewExecution("secretscopes", "list_secrets")
	scopeSecrets, err := r.APIClient.Secrets().ListSecrets(scope)
	execution.Finish(err)
	if err != nil {
		return err
	}

	// delete any existing secrets. We cannot update since we cannot inspect a secrets values
	// therefore, we delete all then create all
	if len(scopeSecrets) > 0 {
		for _, existingSecret := range scopeSecrets {
			execution := NewExecution("secretscopes", "delete_secret")
			err = r.APIClient.Secrets().DeleteSecret(scope, existingSecret.Key)
			execution.Finish(err)
			if err != nil {
				return err
			}
		}
	}

	for _, secret := range instance.Spec.SecretScopeSecrets {
		if secret.StringValue != "" {
			execution := NewExecution("secretscopes", "put_secret_string")
			err = r.APIClient.Secrets().PutSecretString(secret.StringValue, scope, secret.Key)
			execution.Finish(err)
			if err != nil {
				return err
			}
		} else if secret.ByteValue != "" {
			v, err := base64.StdEncoding.DecodeString(secret.ByteValue)
			if err != nil {
				return err
			}
			execution := NewExecution("secretscopes", "put_secret")
			err = r.APIClient.Secrets().PutSecret(v, scope, secret.Key)
			execution.Finish(err)
			if err != nil {
				return err
			}
		} else if secret.ValueFrom != nil {
			value, err := r.getSecretValueFrom(namespace, secret)
			if err != nil {
				return err
			}

			execution := NewExecution("secretscopes", "put_secret_string")
			err = r.APIClient.Secrets().PutSecretString(value, scope, secret.Key)
			execution.Finish(err)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *SecretScopeReconciler) getSecretValueFrom(namespace string, scopeSecret databricksv1alpha1.SecretScopeSecret) (string, error) {
	if &scopeSecret.ValueFrom != nil {
		namespacedName := types.NamespacedName{Namespace: namespace, Name: scopeSecret.ValueFrom.SecretKeyRef.Name}
		secret := &v1.Secret{}
		err := r.Get(context.Background(), namespacedName, secret)
		if err != nil {
			return "", err
		}

		value := string(secret.Data[scopeSecret.ValueFrom.SecretKeyRef.Key])
		return value, nil
	}

	return "", fmt.Errorf("No ValueFrom present to extract secret")
}

func (r *SecretScopeReconciler) submitACLs(instance *databricksv1alpha1.SecretScope) error {
	scope := instance.ObjectMeta.Name
	execution := NewExecution("secretscopes", "list_secret_acls")
	scopeSecretACLs, err := r.APIClient.Secrets().ListSecretACLs(scope)
	execution.Finish(err)
	if err != nil {
		return err
	}

	if len(scopeSecretACLs) > 0 {
		for _, existingACL := range scopeSecretACLs {
			execution := NewExecution("secretscopes", "delete_secret_acl")
			err = r.APIClient.Secrets().DeleteSecretACL(scope, existingACL.Principal)
			execution.Finish(err)
			if err != nil {
				return err
			}
		}
	}

	for _, acl := range instance.Spec.SecretScopeACLs {
		var permission dbmodels.AclPermission
		if acl.Permission == "READ" {
			permission = dbmodels.AclPermissionRead
		} else if acl.Permission == "WRITE" {
			permission = dbmodels.AclPermissionWrite
		} else if acl.Permission == "MANAGE" {
			permission = dbmodels.AclPermissionManage
		} else {
			err = fmt.Errorf("Bad Permission")
		}

		if err != nil {
			return err
		}

		execution := NewExecution("secretscopes", "put_secret_acl")
		err = r.APIClient.Secrets().PutSecretACL(scope, acl.Principal, permission)
		execution.Finish(err)
	}

	return nil
}

// checkSecrets checks if referenced secret is present in k8s or not.
func (r *SecretScopeReconciler) checkSecrets(instance *databricksv1alpha1.SecretScope) error {
	namespace := instance.Namespace

	// if secret in cluster is reference, see if secret exists.
	for _, secret := range instance.Spec.SecretScopeSecrets {
		if secret.ValueFrom != nil {
			if _, err := r.getSecretValueFrom(namespace, secret); err != nil {
				return err
			}
		}
	}

	instance.Status.SecretInClusterAvailable = true
	return r.Update(context.Background(), instance)
}

func (r *SecretScopeReconciler) submit(instance *databricksv1alpha1.SecretScope) (requeue bool, err error) {
	scope := instance.ObjectMeta.Name
	initialManagePrincipal := instance.Spec.InitialManagePrincipal

	execution := NewExecution("secretscopes", "create_secret_scope")
	err = r.APIClient.Secrets().CreateSecretScope(scope, initialManagePrincipal)
	execution.Finish(err)
	if err != nil {
		return
	}

	err = r.submitSecrets(instance)
	if err != nil {
		requeue = true
		return
	}

	if instance.Spec.SecretScopeACLs != nil {
		err = r.submitACLs(instance)
		if err != nil {
			return
		}
	}

	remoteScope, err := r.get(scope)
	if err != nil {
		requeue = true
		return
	}

	instance.Status.SecretScope = remoteScope
	return true, r.Update(context.Background(), instance)
}

func (r *SecretScopeReconciler) delete(instance *databricksv1alpha1.SecretScope) error {

	if instance.Status.SecretScope != nil {
		scope := instance.Status.SecretScope.Name
		execution := NewExecution("secretscopes", "delete_secret_scope")
		err := r.APIClient.Secrets().DeleteSecretScope(scope)
		execution.Finish(err)
		if err != nil && !strings.Contains(err.Error(), "does not exist") {
			return err
		}
	}
	return nil
}
