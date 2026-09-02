// Copyright 2025 MongoDB Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package generator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestValidateCRD(t *testing.T) {
	tests := map[string]struct {
		crd     *apiextensions.CustomResourceDefinition
		wantErr string
	}{
		"valid CRD": {
			crd: &apiextensions.CustomResourceDefinition{
				ObjectMeta: v1.ObjectMeta{
					Name: "examples.test.com",
				},
				Spec: apiextensions.CustomResourceDefinitionSpec{
					Group: "test.com",
					Names: apiextensions.CustomResourceDefinitionNames{
						Plural:   "examples",
						Singular: "example",
						Kind:     "Example",
						ListKind: "ExampleList",
					},
					Versions: []apiextensions.CustomResourceDefinitionVersion{
						{
							Name:    "v1",
							Served:  true,
							Storage: true,
						},
					},
					Validation: &apiextensions.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
							Type: "object",
						},
					},
					Scope: apiextensions.NamespaceScoped,
				},
				Status: apiextensions.CustomResourceDefinitionStatus{
					StoredVersions: []string{"v1"},
				},
			},
		},
		"invalid CRD": {
			wantErr: "must be spec.names.plural",
			crd: &apiextensions.CustomResourceDefinition{
				ObjectMeta: v1.ObjectMeta{
					Name: "examples.test.com",
				},
				Spec: apiextensions.CustomResourceDefinitionSpec{
					Group: "test.com",
					Names: apiextensions.CustomResourceDefinitionNames{
						Plural:   "wrongs",
						Singular: "wrong",
						Kind:     "Wrong",
						ListKind: "WrongList",
					},
					Versions: []apiextensions.CustomResourceDefinitionVersion{
						{
							Name:    "v1",
							Served:  true,
							Storage: true,
						},
					},
					Validation: &apiextensions.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
							Type: "object",
						},
					},
					Scope: apiextensions.NamespaceScoped,
				},
				Status: apiextensions.CustomResourceDefinitionStatus{
					StoredVersions: []string{"v1"},
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := ValidateCRD(context.Background(), tt.crd)
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}
