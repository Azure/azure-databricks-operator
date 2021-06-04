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

	databricksv1alpha2 "github.com/polar-rams/azure-databricks-operator/api/v1alpha2"
	"github.com/polar-rams/databricks-sdk-golang/azure/secrets/httpmodels"
	dbmodels "github.com/polar-rams/databricks-sdk-golang/azure/secrets/models"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *SecretScopeReconciler) get(scope string) (*dbmodels.SecretScope, error) {
	execution := NewExecution("secretscopes", "list_secret_scops")
	lscr, err := r.APIClient.Secrets().ListSecretScopes()
	execution.Finish(err)
	if err != nil {
		return nil, err
	}
	scopes := lscr.Scopes

	matchingScope := dbmodels.SecretScope{}
	for _, existingScope := range *scopes {
		if existingScope.Name == scope {
			matchingScope = existingScope
		}
	}

	if (dbmodels.SecretScope{}) == matchingScope {
		return nil, fmt.Errorf("get for secret scope failed. scope not found: %s", scope)
	}

	return &matchingScope, nil
}

func (r *SecretScopeReconciler) submitSecrets(instance *databricksv1alpha2.SecretScope) error {
	scope := instance.ObjectMeta.Name
	namespace := instance.Namespace
	execution := NewExecution("secretscopes", "list_secrets")
	lsr := httpmodels.ListSecretsReq{
		Scope: scope,
	}
	lsrsp, err := r.APIClient.Secrets().ListSecrets(lsr)
	scopeSecrets := lsrsp.Secrets
	execution.Finish(err)
	if err != nil {
		return err
	}

	// delete any existing secrets. We cannot update since we cannot inspect a secrets values
	// therefore, we delete all then create all
	if len(scopeSecrets) > 0 {
		for _, existingSecret := range scopeSecrets {
			execution := NewExecution("secretscopes", "delete_secret")
			dsr := httpmodels.DeleteSecretReq{
				Scope: scope,
				Key:   existingSecret.Key,
			}
			err = r.APIClient.Secrets().DeleteSecret(dsr)
			execution.Finish(err)
			if err != nil {
				return err
			}
		}
	}

	for _, secret := range instance.Spec.SecretScopeSecrets {
		if secret.StringValue != "" {
			execution := NewExecution("secretscopes", "put_secret_string")
			psr := httpmodels.PutSecretReq{
				StringValue: secret.StringValue,
				Scope:       scope,
				Key:         secret.Key,
			}
			err = r.APIClient.Secrets().PutSecret(psr)
			execution.Finish(err)
			if err != nil {
				return err
			}
		} else if secret.ByteValue != "" {
			// v, err := base64.StdEncoding.DecodeString(secret.ByteValue)
			// if err != nil {
			// 	return err
			// }
			execution := NewExecution("secretscopes", "put_secret")
			psr := httpmodels.PutSecretReq{
				BytesValue: secret.ByteValue,
				Scope:      scope,
				Key:        secret.Key,
			}
			err = r.APIClient.Secrets().PutSecret(psr)
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
			psr := httpmodels.PutSecretReq{
				StringValue: value,
				Scope:       scope,
				Key:         secret.Key,
			}
			err = r.APIClient.Secrets().PutSecret(psr)
			// err = r.APIClient.Secrets().PutSecretString(value, scope, secret.Key)
			execution.Finish(err)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *SecretScopeReconciler) getSecretValueFrom(namespace string, scopeSecret databricksv1alpha2.SecretScopeSecret) (string, error) {
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

func (r *SecretScopeReconciler) submitACLs(instance *databricksv1alpha2.SecretScope) error {
	scope := instance.ObjectMeta.Name
	execution := NewExecution("secretscopes", "list_secret_acls")
	lsr := httpmodels.ListSecretACLsReq{
		Scope: scope,
	}
	lsrsp, err := r.APIClient.Secrets().ListSecretACLs(lsr)
	scopeSecretACLs := lsrsp.Items
	execution.Finish(err)
	if err != nil {
		return err
	}

	if len(scopeSecretACLs) > 0 {
		for _, existingACL := range scopeSecretACLs {
			execution := NewExecution("secretscopes", "delete_secret_acl")
			dsr := httpmodels.DeleteSecretACLReq{
				Scope:     scope,
				Principal: existingACL.Principal,
			}
			err = r.APIClient.Secrets().DeleteSecretACL(dsr)
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
		psr := httpmodels.PutSecretACLReq{
			Scope:      scope,
			Principal:  acl.Principal,
			Permission: permission,
		}
		err = r.APIClient.Secrets().PutSecretACL(psr)
		execution.Finish(err)
	}

	return nil
}

// checkSecrets checks if referenced secret is present in k8s or not.
func (r *SecretScopeReconciler) checkSecrets(instance *databricksv1alpha2.SecretScope) error {
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

func (r *SecretScopeReconciler) submit(instance *databricksv1alpha2.SecretScope) (requeue bool, err error) {
	scope := instance.ObjectMeta.Name
	initialManagePrincipal := instance.Spec.InitialManagePrincipal

	execution := NewExecution("secretscopes", "create_secret_scope")
	csr := httpmodels.CreateSecretScopeReq{
		Scope:                  scope,
		InitialManagePrincipal: initialManagePrincipal,
	}
	err = r.APIClient.Secrets().CreateSecretScope(csr)
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

func (r *SecretScopeReconciler) delete(instance *databricksv1alpha2.SecretScope) error {

	if instance.Status.SecretScope != nil {
		scope := instance.Status.SecretScope.Name
		execution := NewExecution("secretscopes", "delete_secret_scope")
		dsr := httpmodels.DeleteSecretScopeReq{
			Scope: scope,
		}
		err := r.APIClient.Secrets().DeleteSecretScope(dsr)
		execution.Finish(err)
		if err != nil && !strings.Contains(err.Error(), "does not exist") {
			return err
		}
	}
	return nil
}
