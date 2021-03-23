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

const (
	// VPAKind just to avoid mispeling
	VPAKind = "VerticalPodAutoscaler"
	// HPAKind just to avoid mispeling
	HPAKind = "HorizontalPodAutoscaler"
)

// GroupVersionKind is the reprsentation of k8s type
// This object is used to to avoid sprawl of dependent library (eg. apimachinary) across the code
// This will allow easy migration to others library (eg. kyaml) in the future once the dependency is all encapulated into k8s_decoder.go
type GroupVersionKind struct {
	Group   string
	Version string
	Kind    string
}

//Resource represents the VPA Resource recomendation
type Resource struct {
	CPU    float32
	Memory int64
}

// VPARecomendation represents the VPA Container Recomendations
type VPARecomendation struct {
	ContainerName string
	Target        Resource
	UpperBound    Resource
}

// TargetRef references the object which VPA applies to
type TargetRef struct {
	APIVersion string
	Kind       string
	Name       string
}

// VPA is the simplified reprsentation of k8s VPA
// Client doesn't need to handle different version and the complexity of k8s.io package
type VPA struct {
	Namespace             string
	Name                  string
	TargetRef             TargetRef
	IsInRecomendationMode bool
	Recomendations        []VPARecomendation
}

// HPA is the simplified reprsentation of k8s HPA
// Client doesn't need to handle different version and the complexity of k8s.io package
type HPA struct {
	Namespace           string
	Name                string
	TargetRef           TargetRef
	MinReplicas         int32
	MaxReplicas         int32
	TargetCPUPercentage int32
}
