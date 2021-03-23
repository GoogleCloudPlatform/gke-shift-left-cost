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
	"strings"
	"testing"
)

func TestVPAV1beta1NotImplemented(t *testing.T) {
	yaml := `
apiVersion: autoscaling.k8s.io/v1beta1
kind: VerticalPodAutoscaler
metadata:
  name: redis-vpa
spec:
  selector:
    matchLabels:
      label: vpa-label
  updatePolicy:
    updateMode: "Off"`

	_, err := decodeVPA([]byte(yaml))
	if err == nil || !strings.HasPrefix(err.Error(), "APIVersion and Kind not Supported") {
		t.Error(fmt.Errorf("Should have return an APIVersion error, but returned '%+v'", err))
	}
}

func TestVPAV1WithoutStatus(t *testing.T) {
	yaml := `
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: redis-vpa
spec:
  targetRef:
    apiVersion: "apps/v1"
    kind: Deployment
    name: redis-master`

	vpa, err := decodeVPA([]byte(yaml))
	if err != nil {
		t.Error(err)
		return
	}

	if vpa.IsInRecomendationMode {
		t.Error("VPA object is in Auto mode")
	}

	if len(vpa.Recomendations) > 0 {
		t.Error("VPA object has no recomendation.")
	}
}

func TestVPAV1WithStatus(t *testing.T) {
	yaml := `
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: redis-vpa
  namespace: otherns
spec:
  targetRef:
    apiVersion: "apps/v1"
    kind: Deployment
    name: redis-master
  updatePolicy:
    updateMode: "Off"    
status:
  conditions:
  - lastTransitionTime: "2021-02-10T14:20:46Z"
    message: Fetching history complete
    status: "False"
    type: FetchingHistory
  - lastTransitionTime: "2021-02-10T14:19:46Z"
    status: "False"
    type: LowConfidence
  - lastTransitionTime: "2021-02-10T14:19:46Z"
    status: "True"
    type: RecommendationProvided
  recommendation:
    containerRecommendations:
    - containerName: server
      lowerBound:
        cpu: 25m
        memory: 262144k
      target:
        cpu: 35m
        memory: 262144k
      uncappedTarget:
        cpu: 35m
        memory: 262144k
      upperBound:
        cpu: 39m
        memory: 262144k`

	vpa, err := decodeVPA([]byte(yaml))
	if err != nil {
		t.Error(err)
		return
	}

	expectedNS := "otherns"
	if got := vpa.Namespace; got != expectedNS {
		t.Errorf("Expected Namespace %+v, got %+v", expectedNS, got)
	}

	expectedName := "redis-vpa"
	if got := vpa.Name; got != expectedName {
		t.Errorf("Expected Name %+v, got %+v", expectedName, got)
	}

	expectedTargetRef := TargetRef{APIVersion: "apps/v1", Kind: "Deployment", Name: "redis-master"}
	if got := vpa.TargetRef; got != expectedTargetRef {
		t.Errorf("Expected TargetRef %+v, got %+v", expectedTargetRef, got)
	}

	if !vpa.IsInRecomendationMode {
		t.Error("VPA object is in recomendation mode")
	}

	recomendation := vpa.Recomendations[0]
	expectedCPU := float32(0.035)
	expectedMemory := int64(262144000)
	if recomendation.Target.CPU != expectedCPU {
		t.Errorf("Expected Target Recomended CPU %+v, got %+v", expectedCPU, recomendation.Target.CPU)
	}
	if recomendation.Target.Memory != expectedMemory {
		t.Errorf("Expected Target Recomended Memory %+v, got %+v", expectedMemory, recomendation.Target.Memory)
	}
	expectedCPU = float32(0.039)
	expectedMemory = int64(262144000)
	if recomendation.UpperBound.CPU != expectedCPU {
		t.Errorf("Expected UpperBound Recomended CPU %+v, got %+v", expectedCPU, recomendation.UpperBound.CPU)
	}
	if recomendation.UpperBound.Memory != expectedMemory {
		t.Errorf("Expected UpperBound Recomended Memory %+v, got %+v", expectedMemory, recomendation.UpperBound.Memory)
	}
}

func TestVPAV1WithManyContainers(t *testing.T) {
	yaml := `
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: redis-vpa
  namespace: otherns
spec:
  targetRef:
    apiVersion: "apps/v1"
    kind: Deployment
    name: redis-master
  updatePolicy:
    updateMode: "Off"    
status:
  conditions:
  - lastTransitionTime: "2021-02-10T14:20:46Z"
    message: Fetching history complete
    status: "False"
    type: FetchingHistory
  - lastTransitionTime: "2021-02-10T14:19:46Z"
    status: "False"
    type: LowConfidence
  - lastTransitionTime: "2021-02-10T14:19:46Z"
    status: "True"
    type: RecommendationProvided
  recommendation:
    containerRecommendations:
    - containerName: server
      lowerBound:
        cpu: 25m
        memory: 262144k
      target:
        cpu: 35m
        memory: 262144k
      uncappedTarget:
        cpu: 35m
        memory: 262144k
      upperBound:
        cpu: 39m
        memory: 262144k
    - containerName: server2
      lowerBound:
        cpu: 25m
        memory: 262144k
      target:
        cpu: 33m
        memory: 262144k
      uncappedTarget:
        cpu: 35m
        memory: 262144k
      upperBound:
        cpu: 40m
        memory: 262145k`

	vpa, err := decodeVPA([]byte(yaml))
	if err != nil {
		t.Error(err)
		return
	}

	if len(vpa.Recomendations) != 2 {
		t.Errorf("VPA object must have two recomendations")
	}

	recomendation := vpa.Recomendations[0]
	expectedContainerName := "server"
	if recomendation.ContainerName != expectedContainerName {
		t.Errorf("Expected Container Name %+v, got %+v", expectedContainerName, recomendation.ContainerName)
	}
	expectedCPU := float32(0.035)
	expectedMemory := int64(262144000)
	if recomendation.Target.CPU != expectedCPU {
		t.Errorf("Expected Target Recomended CPU %+v, got %+v", expectedCPU, recomendation.Target.CPU)
	}
	if recomendation.Target.Memory != expectedMemory {
		t.Errorf("Expected Target Recomended Memory %+v, got %+v", expectedMemory, recomendation.Target.Memory)
	}
	expectedCPU = float32(0.039)
	expectedMemory = int64(262144000)
	if recomendation.UpperBound.CPU != expectedCPU {
		t.Errorf("Expected UpperBound Recomended CPU %+v, got %+v", expectedCPU, recomendation.UpperBound.CPU)
	}
	if recomendation.UpperBound.Memory != expectedMemory {
		t.Errorf("Expected UpperBound Recomended Memory %+v, got %+v", expectedMemory, recomendation.UpperBound.Memory)
	}

	recomendation = vpa.Recomendations[1]
	expectedContainerName = "server2"
	if recomendation.ContainerName != expectedContainerName {
		t.Errorf("Expected Container Name %+v, got %+v", expectedContainerName, recomendation.ContainerName)
	}
	expectedCPU = float32(0.033)
	expectedMemory = int64(262144000)
	if recomendation.Target.CPU != expectedCPU {
		t.Errorf("Expected Target Recomended CPU %+v, got %+v", expectedCPU, recomendation.Target.CPU)
	}
	if recomendation.Target.Memory != expectedMemory {
		t.Errorf("Expected Target Recomended Memory %+v, got %+v", expectedMemory, recomendation.Target.Memory)
	}
	expectedCPU = float32(0.040)
	expectedMemory = int64(262145000)
	if recomendation.UpperBound.CPU != expectedCPU {
		t.Errorf("Expected UpperBound Recomended CPU %+v, got %+v", expectedCPU, recomendation.UpperBound.CPU)
	}
	if recomendation.UpperBound.Memory != expectedMemory {
		t.Errorf("Expected UpperBound Recomended Memory %+v, got %+v", expectedMemory, recomendation.UpperBound.Memory)
	}
}

func TestVPAV1beta2WithStatus(t *testing.T) {
	yaml := `
apiVersion: autoscaling.k8s.io/v1beta2
kind: VerticalPodAutoscaler
metadata:
  name: redis-vpa
spec:
  targetRef:
    apiVersion: "apps/v1"
    kind: Deployment
    name: redis-master
  updatePolicy:
    updateMode: "Off"    
status:
  conditions:
  - lastTransitionTime: "2021-02-10T14:20:46Z"
    message: Fetching history complete
    status: "False"
    type: FetchingHistory
  - lastTransitionTime: "2021-02-10T14:19:46Z"
    status: "False"
    type: LowConfidence
  - lastTransitionTime: "2021-02-10T14:19:46Z"
    status: "True"
    type: RecommendationProvided
  recommendation:
    containerRecommendations:
    - containerName: server
      lowerBound:
        cpu: 25m
        memory: 262144k
      target:
        cpu: 35m
        memory: 262144k
      uncappedTarget:
        cpu: 35m
        memory: 262144k
      upperBound:
        cpu: 39m
        memory: 262144k`

	vpa, err := decodeVPA([]byte(yaml))
	if err != nil {
		t.Error(err)
		return
	}

	expectedNS := "default"
	if got := vpa.Namespace; got != expectedNS {
		t.Errorf("Expected Namespace %+v, got %+v", expectedNS, got)
	}

	expectedName := "redis-vpa"
	if got := vpa.Name; got != expectedName {
		t.Errorf("Expected Name %+v, got %+v", expectedName, got)
	}

	expectedTargetRef := TargetRef{APIVersion: "apps/v1", Kind: "Deployment", Name: "redis-master"}
	if got := vpa.TargetRef; got != expectedTargetRef {
		t.Errorf("Expected TargetRef %+v, got %+v", expectedTargetRef, got)
	}

	if !vpa.IsInRecomendationMode {
		t.Error("VPA object is in recomendation mode")
	}

	recomendation := vpa.Recomendations[0]
	expectedCPU := float32(0.035)
	expectedMemory := int64(262144000)
	if recomendation.Target.CPU != expectedCPU {
		t.Errorf("Expected Target Recomended CPU %+v, got %+v", expectedCPU, recomendation.Target.CPU)
	}
	if recomendation.Target.Memory != expectedMemory {
		t.Errorf("Expected Target Recomended Memory %+v, got %+v", expectedMemory, recomendation.Target.Memory)
	}
	expectedCPU = float32(0.039)
	expectedMemory = int64(262144000)
	if recomendation.UpperBound.CPU != expectedCPU {
		t.Errorf("Expected UpperBound Recomended CPU %+v, got %+v", expectedCPU, recomendation.UpperBound.CPU)
	}
	if recomendation.Target.Memory != expectedMemory {
		t.Errorf("Expected UpperBound Recomended Memory %+v, got %+v", expectedMemory, recomendation.UpperBound.Memory)
	}
}

func TestVPAList(t *testing.T) {
	yaml := `
  apiVersion: v1
  items:
  - apiVersion: autoscaling.k8s.io/v1
    kind: VerticalPodAutoscaler
    metadata:
      name: adservice-vpa
      namespace: default
    spec:
      targetRef:
        apiVersion: apps/v1
        kind: Deployment
        name: adservice
        namespace: default
      updatePolicy:
        updateMode: "Off"
    status:
      conditions:
      - lastTransitionTime: "2021-02-10T14:20:46Z"
        message: Fetching history complete
        status: "False"
        type: FetchingHistory
      - lastTransitionTime: "2021-02-10T14:19:46Z"
        status: "False"
        type: LowConfidence
      - lastTransitionTime: "2021-02-10T14:19:46Z"
        status: "True"
        type: RecommendationProvided
      recommendation:
        containerRecommendations:
        - containerName: server
          lowerBound:
            cpu: 25m
            memory: "351112894"
          target:
            cpu: 35m
            memory: "351198544"
          uncappedTarget:
            cpu: 35m
            memory: "351198544"
          upperBound:
            cpu: 39m
            memory: "394031262"
  - apiVersion: autoscaling.k8s.io/v1
    kind: VerticalPodAutoscaler
    metadata:
      name: cartservice-vpa
      namespace: default
    spec:
      targetRef:
        apiVersion: apps/v1
        kind: Deployment
        name: cartservice
        namespace: default
      updatePolicy:
        updateMode: "Off"
    status:
      conditions:
      - lastTransitionTime: "2021-02-10T14:20:46Z"
        message: Fetching history complete
        status: "False"
        type: FetchingHistory
      - lastTransitionTime: "2021-02-10T14:19:46Z"
        status: "False"
        type: LowConfidence
      - lastTransitionTime: "2021-02-10T14:19:46Z"
        status: "True"
        type: RecommendationProvided
      recommendation:
        containerRecommendations:
        - containerName: server
          lowerBound:
            cpu: 25m
            memory: 262144k
          target:
            cpu: 25m
            memory: 262144k
          uncappedTarget:
            cpu: 25m
            memory: 262144k
          upperBound:
            cpu: 25m
            memory: 262144k
  kind: List
  metadata:
    resourceVersion: ""
    selfLink: ""`

	vpaList, err := DecodeVPAList([]byte(yaml))
	if err != nil {
		t.Error(err)
		return
	}

	if len(vpaList) != 2 {
		t.Error("VPA List has two objects")
	}
}

func TestVPAListWithUnsuportedVersion(t *testing.T) {
	yaml := `
  apiVersion: v1
  items:
  - apiVersion: autoscaling.k8s.io/v1
    kind: VerticalPodAutoscaler
    metadata:
      name: adservice-vpa
      namespace: default
    spec:
      targetRef:
        apiVersion: apps/v1
        kind: Deployment
        name: adservice
        namespace: default
      updatePolicy:
        updateMode: "Off"
    status:
      conditions:
      - lastTransitionTime: "2021-02-10T14:20:46Z"
        message: Fetching history complete
        status: "False"
        type: FetchingHistory
      - lastTransitionTime: "2021-02-10T14:19:46Z"
        status: "False"
        type: LowConfidence
      - lastTransitionTime: "2021-02-10T14:19:46Z"
        status: "True"
        type: RecommendationProvided
      recommendation:
        containerRecommendations:
        - containerName: server
          lowerBound:
            cpu: 25m
            memory: "351112894"
          target:
            cpu: 35m
            memory: "351198544"
          uncappedTarget:
            cpu: 35m
            memory: "351198544"
          upperBound:
            cpu: 39m
            memory: "394031262"
  - apiVersion: autoscaling.k8s.io/v1beta1
    kind: VerticalPodAutoscaler
    metadata:
      name: cartservice-vpa
      namespace: default
    spec:
      selector:
        matchLabels:
          label: vpa-label
      updatePolicy:
        updateMode: "Off"
    status:
      conditions:
      - lastTransitionTime: "2021-02-10T14:20:46Z"
        message: Fetching history complete
        status: "False"
        type: FetchingHistory
      - lastTransitionTime: "2021-02-10T14:19:46Z"
        status: "False"
        type: LowConfidence
      - lastTransitionTime: "2021-02-10T14:19:46Z"
        status: "True"
        type: RecommendationProvided
      recommendation:
        containerRecommendations:
        - containerName: server
          lowerBound:
            cpu: 25m
            memory: 262144k
          target:
            cpu: 25m
            memory: 262144k
          uncappedTarget:
            cpu: 25m
            memory: 262144k
          upperBound:
            cpu: 25m
            memory: 262144k
  kind: List
  metadata:
    resourceVersion: ""
    selfLink: ""`

	vpaList, err := DecodeVPAList([]byte(yaml))
	if err != nil {
		t.Error(err)
		return
	}

	if len(vpaList) != 1 {
		t.Error("VPA List has only one supported objects")
	}
}
