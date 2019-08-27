/*
Copyright 2019 The Crossplane Authors.

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

package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	databasev1alpha1 "github.com/crossplaneio/crossplane/apis/database/v1alpha1"

	"github.com/crossplaneio/stack-gcp/gcp/apis/database/v1alpha2"
)

// ConfigurePostgreSQLCloudsqlInstance configures the supplied instance (presumed
// to be a CloudsqlInstance) using the supplied instance claim (presumed to be a
// PostgreSQLInstance) and instance class.
func ConfigurePostgreSQLCloudsqlInstance(_ context.Context, cm resource.Claim, cs resource.Class, mg resource.Managed) error {
	pg, cmok := cm.(*databasev1alpha1.PostgreSQLInstance)
	if !cmok {
		return errors.Errorf("expected resource claim %s to be %s", cm.GetName(), databasev1alpha1.PostgreSQLInstanceGroupVersionKind)
	}

	rs, csok := cs.(*v1alpha2.CloudsqlInstanceClass)
	if !csok {
		return errors.Errorf("expected resource class %s to be %s", cs.GetName(), v1alpha2.CloudsqlInstanceClassGroupVersionKind)
	}

	i, mgok := mg.(*v1alpha2.CloudsqlInstance)
	if !mgok {
		return errors.Errorf("expected managed instance %s to be %s", mg.GetName(), v1alpha2.CloudsqlInstanceGroupVersionKind)
	}

	spec := &v1alpha2.CloudsqlInstanceSpec{
		ResourceSpec: runtimev1alpha1.ResourceSpec{
			ReclaimPolicy: runtimev1alpha1.ReclaimRetain,
		},
		CloudsqlInstanceParameters: rs.SpecTemplate.CloudsqlInstanceParameters,
	}
	translated := translateVersion(pg.Spec.EngineVersion, v1alpha2.PostgresqlDBVersionPrefix)
	v, err := resource.ResolveClassClaimValues(spec.DatabaseVersion, translated)
	if err != nil {
		return err
	}
	spec.DatabaseVersion = v

	// NOTE(hasheddan): consider moving defaulting to either CRD or managed reconciler level
	checkEmptySpec(spec)

	spec.WriteConnectionSecretToReference = corev1.LocalObjectReference{Name: string(cm.GetUID())}
	spec.ProviderReference = rs.SpecTemplate.ProviderReference
	spec.ReclaimPolicy = rs.SpecTemplate.ReclaimPolicy

	i.Spec = *spec

	return nil
}

// ConfigureMyCloudsqlInstance configures the supplied instance (presumed to be
// a CloudsqlInstance) using the supplied instance claim (presumed to be a
// MySQLInstance) and instance class.
func ConfigureMyCloudsqlInstance(_ context.Context, cm resource.Claim, cs resource.Class, mg resource.Managed) error {
	my, cmok := cm.(*databasev1alpha1.MySQLInstance)
	if !cmok {
		return errors.Errorf("expected instance claim %s to be %s", cm.GetName(), databasev1alpha1.MySQLInstanceGroupVersionKind)
	}

	rs, csok := cs.(*v1alpha2.CloudsqlInstanceClass)
	if !csok {
		return errors.Errorf("expected resource class %s to be %s", cs.GetName(), v1alpha2.CloudsqlInstanceClassGroupVersionKind)
	}

	i, mgok := mg.(*v1alpha2.CloudsqlInstance)
	if !mgok {
		return errors.Errorf("expected managed resource %s to be %s", mg.GetName(), v1alpha2.CloudsqlInstanceGroupVersionKind)
	}

	spec := &v1alpha2.CloudsqlInstanceSpec{
		ResourceSpec: runtimev1alpha1.ResourceSpec{
			ReclaimPolicy: runtimev1alpha1.ReclaimRetain,
		},
		CloudsqlInstanceParameters: rs.SpecTemplate.CloudsqlInstanceParameters,
	}

	translated := translateVersion(my.Spec.EngineVersion, v1alpha2.MysqlDBVersionPrefix)
	v, err := resource.ResolveClassClaimValues(spec.DatabaseVersion, translated)
	if err != nil {
		return err
	}
	spec.DatabaseVersion = v

	// NOTE(hasheddan): consider moving defaulting to either CRD or managed reconciler level
	checkEmptySpec(spec)

	spec.WriteConnectionSecretToReference = corev1.LocalObjectReference{Name: string(cm.GetUID())}
	spec.ProviderReference = rs.SpecTemplate.ProviderReference
	spec.ReclaimPolicy = rs.SpecTemplate.ReclaimPolicy

	i.Spec = *spec

	return nil
}

func translateVersion(version, versionPrefix string) string {
	if version == "" {
		return ""
	}
	return fmt.Sprintf("%s_%s", versionPrefix, strings.Replace(version, ".", "_", -1))
}

func checkEmptySpec(spec *v1alpha2.CloudsqlInstanceSpec) {
	if spec.Labels == nil {
		spec.Labels = map[string]string{}
	}
	if spec.AuthorizedNetworks == nil {
		spec.AuthorizedNetworks = []string{}
	}
	if spec.StorageGB == 0 {
		spec.StorageGB = v1alpha2.DefaultStorageGB
	}
}
