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

package controllers

import (
	"context"
	"fmt"
	"strings"

	databricksv1alpha1 "github.com/microsoft/azure-databricks-operator/api/v1alpha1"
	dbhttpmodels "github.com/polar-rams/databricks-sdk-golang/azure/secrets/httpmodels"
	dbmodels "github.com/polar-rams/databricks-sdk-golang/azure/secrets/models"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *SecretScopeReconciler) get(scope string) (*dbmodels.SecretScope, error) {
	execution := NewExecution("secretscopes", "list_secret_scops")
	listSecretScopesRes, err := r.APIClient.Secrets().ListSecretScopes()
	execution.Finish(err)
	if err != nil {
		return nil, err
	}

	matchingScope := dbmodels.SecretScope{}
	if listSecretScopesRes.Scopes != nil {
		for _, existingScope := range *listSecretScopesRes.Scopes {
			if existingScope.Name == scope {
				matchingScope = existingScope
			}
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
	listSecretsReq := dbhttpmodels.ListSecretsReq{
		Scope: scope,
	}
	listSecretsRes, err := r.APIClient.Secrets().ListSecrets(listSecretsReq)
	execution.Finish(err)
	if err != nil {
		return err
	}

	// delete any existing secrets. We cannot update since we cannot inspect a secrets values
	// therefore, we delete all then create all
	if len(listSecretsRes.Secrets) > 0 {
		for _, existingSecret := range listSecretsRes.Secrets {
			execution := NewExecution("secretscopes", "delete_secret")
			deleteSecretReq := dbhttpmodels.DeleteSecretReq{
				Scope: scope,
				Key:   existingSecret.Key,
			}
			err = r.APIClient.Secrets().DeleteSecret(deleteSecretReq)
			execution.Finish(err)
			if err != nil {
				return err
			}
		}
	}

	for _, secret := range instance.Spec.SecretScopeSecrets {
		putSecretReq := dbhttpmodels.PutSecretReq{
			Scope: scope,
			Key:   secret.Key,
		}
		if secret.StringValue != "" {
			putSecretReq.StringValue = secret.StringValue
		} else if secret.ByteValue != "" {
			putSecretReq.BytesValue = secret.ByteValue
		} else if secret.ValueFrom != nil {
			value, err := r.getSecretValueFrom(namespace, secret)
			if err != nil {
				return err
			}
			putSecretReq.StringValue = value
		}

		if putSecretReq.StringValue != "" || putSecretReq.BytesValue != "" {
			execution := NewExecution("secretscopes", "put_secret")
			err = r.APIClient.Secrets().PutSecret(putSecretReq)
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
	listSecretACLsReq := dbhttpmodels.ListSecretACLsReq{
		Scope: scope,
	}
	listSecretACLsRes, err := r.APIClient.Secrets().ListSecretACLs(listSecretACLsReq)
	execution.Finish(err)
	if err != nil {
		return err
	}

	if len(listSecretACLsRes.Items) > 0 {
		for _, existingACL := range listSecretACLsRes.Items {
			execution := NewExecution("secretscopes", "delete_secret_acl")
			deleteSecretACLReq := dbhttpmodels.DeleteSecretACLReq{
				Scope:     scope,
				Principal: existingACL.Principal,
			}
			err = r.APIClient.Secrets().DeleteSecretACL(deleteSecretACLReq)
			execution.Finish(err)
			if err != nil {
				return err
			}
		}
	}

	for _, acl := range instance.Spec.SecretScopeACLs {
		var permission dbmodels.ACLPermission
		if acl.Permission == "READ" {
			permission = dbmodels.ACLPermissionRead
		} else if acl.Permission == "WRITE" {
			permission = dbmodels.ACLPermissionWrite
		} else if acl.Permission == "MANAGE" {
			permission = dbmodels.ACLPermissionManage
		} else {
			err = fmt.Errorf("Bad Permission")
		}

		if err != nil {
			return err
		}

		execution := NewExecution("secretscopes", "put_secret_acl")
		putSecretACLReq := dbhttpmodels.PutSecretACLReq{
			Scope:      scope,
			Principal:  acl.Principal,
			Permission: permission,
		}
		err = r.APIClient.Secrets().PutSecretACL(putSecretACLReq)
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
	createSecretScopeReq := dbhttpmodels.CreateSecretScopeReq{
		Scope:                  scope,
		InitialManagePrincipal: initialManagePrincipal,
	}
	err = r.APIClient.Secrets().CreateSecretScope(createSecretScopeReq)
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
		deleteSecretScopeReq := dbhttpmodels.DeleteSecretScopeReq{
			Scope: scope,
		}
		err := r.APIClient.Secrets().DeleteSecretScope(deleteSecretScopeReq)
		execution.Finish(err)
		if err != nil && !strings.Contains(err.Error(), "does not exist") {
			return err
		}
	}
	return nil
}
