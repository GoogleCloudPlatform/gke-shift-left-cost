// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8s

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	vpaV1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpaV1beta1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta1"

	hpaV1 "k8s.io/api/autoscaling/v1"
	hpaV2beta1 "k8s.io/api/autoscaling/v2beta1"
	hpaV2beta2 "k8s.io/api/autoscaling/v2beta2"
)

var gvkList schema.GroupVersionKind = schema.GroupVersionKind{Group: "", Version: "v1", Kind: "List"}
var gvkVPAV1 schema.GroupVersionKind = schema.GroupVersionKind{Group: "autoscaling.k8s.io", Version: "v1", Kind: VPAKind}
var gvkVPAV1beta1 schema.GroupVersionKind = schema.GroupVersionKind{Group: "autoscaling.k8s.io", Version: "v1beta1", Kind: VPAKind}
var gvkVPAV1beta2 schema.GroupVersionKind = schema.GroupVersionKind{Group: "autoscaling.k8s.io", Version: "v1beta2", Kind: VPAKind}
var gvkHPAV1 schema.GroupVersionKind = schema.GroupVersionKind{Group: "autoscaling", Version: "v1", Kind: HPAKind}
var gvkHPAV2beta1 schema.GroupVersionKind = schema.GroupVersionKind{Group: "autoscaling", Version: "v2beta1", Kind: HPAKind}
var gvkHPAV2beta2 schema.GroupVersionKind = schema.GroupVersionKind{Group: "autoscaling", Version: "v2beta2", Kind: HPAKind}

func decode(scheme *runtime.Scheme, data []byte) (runtime.Object, GroupVersionKind, error) {
	decoder := serializer.NewCodecFactory(scheme).UniversalDeserializer()
	obj, gvk, err := decoder.Decode(data, nil, nil)
	if err != nil {
		return (runtime.Object)(nil), GroupVersionKind{}, err
	}
	groupVersionKind := GroupVersionKind{
		Group:   gvk.Group,
		Version: gvk.Version,
		Kind:    gvk.Kind,
	}
	return obj, groupVersionKind, err
}

func buildVPAScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypeWithName(gvkList, &vpaV1.VerticalPodAutoscalerList{})
	scheme.AddKnownTypeWithName(gvkVPAV1, &vpaV1.VerticalPodAutoscaler{})
	scheme.AddKnownTypeWithName(gvkVPAV1beta1, &vpaV1beta1.VerticalPodAutoscaler{})
	scheme.AddKnownTypeWithName(gvkVPAV1beta2, &vpaV1.VerticalPodAutoscaler{})
	return scheme
}

func buildHPAScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypeWithName(gvkList, &hpaV1.HorizontalPodAutoscalerList{})
	scheme.AddKnownTypeWithName(gvkHPAV1, &hpaV1.HorizontalPodAutoscaler{})
	scheme.AddKnownTypeWithName(gvkHPAV2beta1, &hpaV2beta1.HorizontalPodAutoscaler{})
	scheme.AddKnownTypeWithName(gvkHPAV2beta2, &hpaV2beta2.HorizontalPodAutoscaler{})
	return scheme
}
