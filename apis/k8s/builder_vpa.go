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
	"fmt"

	v1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

//DecodeVPAList reads k8s VerticalPodAutoScaler yaml and trasform to VPA object
func DecodeVPAList(data []byte) ([]VPA, error) {
	scheme := buildVPAScheme()
	obj, gvk, err := decode(scheme, data)
	if err != nil {
		return []VPA{}, fmt.Errorf("Error Decoding. Check if your GroupVersionKind is defined in api/k8s/decoder.go. Root cause %+v", err)
	}
	switch obj.(type) {
	case *v1.VerticalPodAutoscalerList:
		return buildVPAList(obj), nil
	default:
		return []VPA{}, fmt.Errorf("APIVersion and Kind not Supported: %+v", gvk)
	}
}

func buildVPAList(obj interface{}) []VPA {
	list := []VPA{}
	vpaList := obj.(*v1.VerticalPodAutoscalerList)
	for _, v := range vpaList.Items {
		gvk := v.GetObjectKind().GroupVersionKind()
		vpa, err := buildVPA(&v, GroupVersionKind{Group: gvk.Group, Version: gvk.Version, Kind: gvk.Kind})
		if err != nil {
			fmt.Printf("Unable to decode object %+v. Root cause: %+v", gvk, err)
		} else {
			list = append(list, vpa)
		}
	}
	return list
}

//decodeVPA reads k8s VerticalPodAutoScaler yaml and trasform to VPA object
func decodeVPA(data []byte) (VPA, error) {
	scheme := buildVPAScheme()
	obj, gvk, err := decode(scheme, data)
	if err != nil {
		return VPA{}, fmt.Errorf("Error Decoding. Check if your GroupVersionKind is defined in api/k8s_decoder.go. Root cause %+v", err)
	}
	return buildVPA(obj, gvk)
}

//buildVPA reads k8s VerticalPodAutoScaler object and trasform to VPA object
func buildVPA(obj interface{}, gvk GroupVersionKind) (VPA, error) {
	switch gvk.Version {
	case gvkVPAV1.Version:
		return buildVPAV1(obj.(*v1.VerticalPodAutoscaler)), nil
	case gvkVPAV1beta2.Version:
		return buildVPAV1(obj.(*v1.VerticalPodAutoscaler)), nil
	default:
		return VPA{}, fmt.Errorf("APIVersion and Kind not Supported: %+v", gvk)
	}
}

func buildVPAV1(vpa *v1.VerticalPodAutoscaler) VPA {
	recomendations := []VPARecomendation{}
	if vpa.Status.Recommendation != nil {
		cr := vpa.Status.Recommendation.ContainerRecommendations
		for _, r := range cr {
			recomendation := VPARecomendation{
				ContainerName: r.ContainerName,
				Target: Resource{
					CPU:    float32(r.Target.Cpu().MilliValue()) / 1000,
					Memory: r.Target.Memory().Value(),
				},
				UpperBound: Resource{
					CPU:    float32(r.UpperBound.Cpu().MilliValue()) / 1000,
					Memory: r.UpperBound.Memory().Value(),
				},
			}
			recomendations = append(recomendations, recomendation)
		}
	}

	isInRecomendationMode := false // default is Auto
	if vpa.Spec.UpdatePolicy != nil {
		isInRecomendationMode = *vpa.Spec.UpdatePolicy.UpdateMode == v1.UpdateModeOff
	}

	tr := vpa.Spec.TargetRef
	ns := "default"
	if vpa.GetNamespace() != "" {
		ns = vpa.GetNamespace()
	}
	return VPA{
		Namespace:             ns,
		Name:                  vpa.GetName(),
		TargetRef:             TargetRef{APIVersion: tr.APIVersion, Kind: tr.Kind, Name: tr.Name},
		IsInRecomendationMode: isInRecomendationMode,
		Recomendations:        recomendations,
	}
}
