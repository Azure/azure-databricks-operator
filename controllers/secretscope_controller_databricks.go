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
	scopes, err := r.APIClient.Secrets().ListSecretScopes()
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
	scopeSecrets, err := r.APIClient.Secrets().ListSecrets(scope)
	if err != nil {
		return err
	}

	// delete any existing secrets. We cannot update since we cannot inspect a secrets values
	// therefore, we delete all then create all
	if len(scopeSecrets) > 0 {
		for _, existingSecret := range scopeSecrets {
			err = r.APIClient.Secrets().DeleteSecret(scope, existingSecret.Key)
			if err != nil {
				return err
			}
		}
	}

	for _, secret := range instance.Spec.SecretScopeSecrets {
		if secret.StringValue != "" {
			err = r.APIClient.Secrets().PutSecretString(secret.StringValue, scope, secret.Key)
			if err != nil {
				return err
			}
		} else if secret.ByteValue != "" {
			v, err := base64.StdEncoding.DecodeString(secret.ByteValue)
			if err != nil {
				return err
			}
			err = r.APIClient.Secrets().PutSecret(v, scope, secret.Key)
			if err != nil {
				return err
			}
		} else if secret.ValueFrom != nil {
			value, err := r.getSecretValueFrom(namespace, secret)
			if err != nil {
				return err
			}

			err = r.APIClient.Secrets().PutSecretString(value, scope, secret.Key)
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
	scopeSecretAcls, err := r.APIClient.Secrets().ListSecretACLs(scope)
	if err != nil {
		return err
	}

	if len(scopeSecretAcls) > 0 {
		for _, existingACL := range scopeSecretAcls {
			err = r.APIClient.Secrets().DeleteSecretACL(scope, existingACL.Principal)
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

		err = r.APIClient.Secrets().PutSecretACL(scope, acl.Principal, permission)
	}

	return nil
}

// checkCluster checks if Databricks cluster supports ACLs, and checks if secret scope exists.
func (r *SecretScopeReconciler) checkCluster(instance *databricksv1alpha1.SecretScope) error {
	scope := instance.ObjectMeta.Name
	initialManagePrincipal := instance.Spec.InitialManagePrincipal

	// try create secret scope to see if exists or not.
	err := r.APIClient.Secrets().CreateSecretScope(scope, initialManagePrincipal)
	if err != nil {
		return err
	}

	// try to list ACLs to see if cluster supports ACL.
	if instance.Spec.SecretScopeACLs != nil {
		if _, err = r.APIClient.Secrets().ListSecretACLs(scope); err != nil {
			if er := r.APIClient.Secrets().DeleteSecretScope(scope); er != nil {
				return fmt.Errorf("%v\n%v", err, er)
			}
			return err
		}
	}

	return nil
}

// checkSecrets checks if referenced secret is present in k8s or not.
func (r *SecretScopeReconciler) checkSecrets(instance *databricksv1alpha1.SecretScope) error {
	scope := instance.ObjectMeta.Name
	namespace := instance.Namespace

	// if secret in cluster is reference, see if secret exists.
	// if secret does not exist, try delete secret scope.
	for _, secret := range instance.Spec.SecretScopeSecrets {
		if secret.ValueFrom != nil {
			if _, err := r.getSecretValueFrom(namespace, secret); err != nil {
				if er := r.APIClient.Secrets().DeleteSecretScope(scope); er != nil {
					return fmt.Errorf("%v\n%v", err, er)
				}
				return err
			}
		}
	}

	return nil
}

func (r *SecretScopeReconciler) submit(instance *databricksv1alpha1.SecretScope) error {
	scope := instance.ObjectMeta.Name

	err := r.submitSecrets(instance)
	if err != nil {
		return err
	}

	if instance.Spec.SecretScopeACLs != nil {
		err := r.submitACLs(instance)
		if err != nil {
			return err
		}
	}

	remoteScope, err := r.get(scope)
	if err != nil {
		return err
	}

	instance.Status.SecretScope = remoteScope
	return r.Update(context.Background(), instance)
}

func (r *SecretScopeReconciler) update(instance *databricksv1alpha1.SecretScope) error {
	err := r.submitSecrets(instance)
	if err != nil {
		return err
	}

	return r.submitACLs(instance)
}

func (r *SecretScopeReconciler) delete(instance *databricksv1alpha1.SecretScope) error {

	if instance.Status.SecretScope != nil {
		scope := instance.Status.SecretScope.Name
		err := r.APIClient.Secrets().DeleteSecretScope(scope)
		if err != nil && !strings.Contains(err.Error(), "does not exist") {
			return err
		}
	}
	return nil
}
