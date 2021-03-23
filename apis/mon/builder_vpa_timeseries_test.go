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

package mon

import (
	"metrics-exporter/apis/k8s"
	"testing"
	"time"
)

func TestBuildVPACPURecomendation(t *testing.T) {
	vpa := k8s.VPA{
		Namespace:             "default",
		Name:                  "currencyservice",
		IsInRecomendationMode: true,
		TargetRef: k8s.TargetRef{
			APIVersion: "v1",
			Kind:       "Deployment",
			Name:       "currencyservice",
		},
		Recomendations: []k8s.VPARecomendation{{
			ContainerName: "server",
			Target:        k8s.Resource{CPU: 400, Memory: 1024},
		}, {
			ContainerName: "sidecar",
			Target:        k8s.Resource{CPU: 200, Memory: 500},
		},
		},
	}

	labels := map[string]string{"name": "a"}
	now := time.Now().Format(time.RFC3339)
	values := []float64{float64(400), float64(200)}

	for i, rec := range vpa.Recomendations {
		ts := buildVPACPURecomendation(vpa, rec, labels, now)

		expected := labels["name"]
		if got := ts.Resource.Labels["name"]; got != expected {
			t.Errorf("Expected label %+v, got %+v", expected, got)
		}

		expected = vpaCpuMetricType
		if got := ts.Metric.Type; got != expected {
			t.Errorf("Expected Metric %+v, got %+v", expected, got)
		}

		metricLabels := map[string]string{
			"targetef_apiversion": vpa.TargetRef.APIVersion,
			"targetref_kind":      vpa.TargetRef.Kind,
			"targetref_name":      vpa.TargetRef.Name,
			"object_name":         vpa.Name,
		}
		for key, expected := range metricLabels {
			if got := ts.Metric.Labels[key]; got != expected {
				t.Errorf("Expected Label %+v, got %+v", expected, got)
			}
		}

		expected = now
		if got := ts.Points[0].Interval.EndTime; got != expected {
			t.Errorf("Expected EndTime %+v, got %+v", expected, got)
		}

		expectedf := values[i]
		if got := ts.Points[0].Value.DoubleValue; *got != expectedf {
			t.Errorf("Expected Value %+v, got %+v", expectedf, *got)
		}
	}
}

func TestBuildVPAMemoryRecomendation(t *testing.T) {
	vpa := k8s.VPA{
		Namespace:             "default",
		Name:                  "currencyservice",
		IsInRecomendationMode: true,
		TargetRef: k8s.TargetRef{
			APIVersion: "v1",
			Kind:       "Deployment",
			Name:       "currencyservice",
		},
		Recomendations: []k8s.VPARecomendation{{
			ContainerName: "server",
			Target:        k8s.Resource{CPU: 400, Memory: 1024},
		}, {
			ContainerName: "sidecar",
			Target:        k8s.Resource{CPU: 200, Memory: 500},
		},
		},
	}

	labels := map[string]string{"name": "a"}
	now := time.Now().Format(time.RFC3339)
	values := []int64{1024, 500}

	for i, rec := range vpa.Recomendations {
		ts := buildVPAMemoryRecomendation(vpa, rec, labels, now)

		expected := labels["name"]
		if got := ts.Resource.Labels["name"]; got != expected {
			t.Errorf("Expected label %+v, got %+v", expected, got)
		}

		expected = vpaMemoryMetricType
		if got := ts.Metric.Type; got != expected {
			t.Errorf("Expected Metric %+v, got %+v", expected, got)
		}

		metricLabels := map[string]string{
			"targetef_apiversion": vpa.TargetRef.APIVersion,
			"targetref_kind":      vpa.TargetRef.Kind,
			"targetref_name":      vpa.TargetRef.Name,
			"object_name":         vpa.Name,
		}
		for key, expected := range metricLabels {
			if got := ts.Metric.Labels[key]; got != expected {
				t.Errorf("Expected Label %+v, got %+v", expected, got)
			}
		}

		expected = now
		if got := ts.Points[0].Interval.EndTime; got != expected {
			t.Errorf("Expected EndTime %+v, got %+v", expected, got)
		}

		expectedf := values[i]
		if got := ts.Points[0].Value.Int64Value; *got != expectedf {
			t.Errorf("Expected Value %+v, got %+v", expectedf, *got)
		}
	}
}

func TestBuildVPARecomendations(t *testing.T) {
	vpa := k8s.VPA{
		Namespace:             "default",
		Name:                  "currencyservice",
		IsInRecomendationMode: true,
		TargetRef: k8s.TargetRef{
			APIVersion: "v1",
			Kind:       "Deployment",
			Name:       "currencyservice",
		},
		Recomendations: []k8s.VPARecomendation{{
			ContainerName: "server",
			Target:        k8s.Resource{CPU: 400, Memory: 1024},
		}, {
			ContainerName: "sidecar",
			Target:        k8s.Resource{CPU: 200, Memory: 500},
		},
		},
	}

	now := time.Now().Format(time.RFC3339)
	tsList := buildVPARecomendations(vpa, now)

	expected := 4
	if got := len(tsList); got != expected {
		t.Errorf("Expected # %+v, got %+v", expected, got)
	}

	metrics := []string{vpaCpuMetricType, vpaMemoryMetricType}
	for i, ts := range tsList {
		expected := metrics[i%2]
		if got := ts.Metric.Type; got != expected {
			t.Errorf("Expected Metric %+v, got %+v", expected, got)
		}
	}
}

func TestBuildVPARecommendationTimeSeries(t *testing.T) {
	vpas := []k8s.VPA{
		{
			Namespace:             "default",
			Name:                  "currencyservice-1",
			IsInRecomendationMode: true,
			TargetRef: k8s.TargetRef{
				APIVersion: "v1",
				Kind:       "Deployment",
				Name:       "currencyservice",
			},
			Recomendations: []k8s.VPARecomendation{{
				ContainerName: "server",
				Target:        k8s.Resource{CPU: 400, Memory: 1024},
			}, {
				ContainerName: "sidecar",
				Target:        k8s.Resource{CPU: 200, Memory: 500},
			},
			},
		},
		{
			Namespace:             "default",
			Name:                  "currencyservice-2",
			IsInRecomendationMode: true,
			TargetRef: k8s.TargetRef{
				APIVersion: "v1",
				Kind:       "Deployment",
				Name:       "currencyservice",
			},
			Recomendations: []k8s.VPARecomendation{{
				ContainerName: "server",
				Target:        k8s.Resource{CPU: 400, Memory: 1024},
			}, {
				ContainerName: "sidecar",
				Target:        k8s.Resource{CPU: 200, Memory: 500},
			},
			},
		},
		{
			Namespace:             "default",
			Name:                  "other",
			IsInRecomendationMode: false,
			TargetRef: k8s.TargetRef{
				APIVersion: "v1",
				Kind:       "Deployment",
				Name:       "other",
			},
			Recomendations: []k8s.VPARecomendation{{
				ContainerName: "server",
				Target:        k8s.Resource{CPU: 400, Memory: 1024},
			}, {
				ContainerName: "sidecar",
				Target:        k8s.Resource{CPU: 200, Memory: 500},
			},
			},
		},
	}

	now := time.Now().Format(time.RFC3339)
	tsList := BuildVPARecommendationTimeSeries(vpas, now)

	expected := 4
	if got := len(tsList); got != expected {
		t.Errorf("Expected # %+v, got %+v", expected, got)
	}

	metrics := []string{vpaCpuMetricType, vpaMemoryMetricType}
	for i, ts := range tsList {
		expected := metrics[i%2]
		if got := ts.Metric.Type; got != expected {
			t.Errorf("Expected Metric %+v, got %+v", expected, got)
		}

		expected = "currencyservice-1"
		if got := ts.Metric.Labels["object_name"]; got != expected {
			t.Errorf("Expected Label %+v, got %+v", expected, got)
		}
	}
}
