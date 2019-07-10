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
	"fmt"

	databricksv1 "github.com/microsoft/azure-databricks-operator/api/v1"
	dbmodels "github.com/xinsnake/databricks-sdk-golang/azure/models"
)

func (r *SecretScopeReconciler) submitSecrets(instance *databricksv1.SecretScope) error {
	scope := instance.ObjectMeta.Name
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
		err = r.APIClient.Secrets().PutSecretString(secret.Value, scope, secret.Key)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *SecretScopeReconciler) submitACLs(instance *databricksv1.SecretScope) error {
	scope := instance.ObjectMeta.Name
	scopeSecretAcls, err := r.APIClient.Secrets().ListSecretACLs(scope)
	if err != nil {
		return err
	}

	if len(scopeSecretAcls) > 0 {
		for _, existingAcl := range scopeSecretAcls {
			err = r.APIClient.Secrets().DeleteSecretACL(scope, existingAcl.Principal)
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

func (r *SecretScopeReconciler) submit(instance *databricksv1.SecretScope) error {
	scope := instance.ObjectMeta.Name
	initialManagePrincipal := instance.Spec.InitialManagePrincipal

	err := r.APIClient.Secrets().CreateSecretScope(scope, initialManagePrincipal)
	if err != nil {
		return err
	}

	err = r.submitSecrets(instance)
	if err != nil {
		return err
	}

	err = r.submitACLs(instance)
	if err != nil {
		return err
	}

	return nil
}
