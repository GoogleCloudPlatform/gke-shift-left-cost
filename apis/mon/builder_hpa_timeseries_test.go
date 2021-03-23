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
	"strconv"
	"testing"
	"time"
)

func TestBuildHPACPUTargetUtilization(t *testing.T) {
	hpa := k8s.HPA{
		Namespace: "default",
		Name:      "currencyservice-hpa",
		TargetRef: k8s.TargetRef{
			APIVersion: "v1",
			Kind:       "Deployment",
			Name:       "currencyservice",
		},
		MinReplicas:         1,
		MaxReplicas:         10,
		TargetCPUPercentage: 70,
	}

	now := time.Now().Format(time.RFC3339)
	ts := buildHPACPUTargetUtilization(hpa, now)

	expected := "currencyservice"
	if got := ts.Resource.Labels["pod_name"]; got != expected {
		t.Errorf("Expected label %+v, got %+v", expected, got)
	}

	expected = hpaCPUTargetUtilizationMetricType
	if got := ts.Metric.Type; got != expected {
		t.Errorf("Expected Metric %+v, got %+v", expected, got)
	}

	metricLabels := map[string]string{
		"targetef_apiversion": hpa.TargetRef.APIVersion,
		"targetref_kind":      hpa.TargetRef.Kind,
		"targetref_name":      hpa.TargetRef.Name,
		"minReplicas":         strconv.Itoa(int(hpa.MinReplicas)),
		"maxReplicas":         strconv.Itoa(int(hpa.MaxReplicas)),
		"object_name":         hpa.Name,
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

	expectedf := int64(70)
	if got := ts.Points[0].Value.Int64Value; *got != expectedf {
		t.Errorf("Expected Value %+v, got %+v", expectedf, *got)
	}
}

func TestBuildHPACPUTargetUtilizationTimeSeriess(t *testing.T) {
	hpas := []k8s.HPA{
		{
			Namespace: "default",
			Name:      "currencyservice-hpa-1",
			TargetRef: k8s.TargetRef{
				APIVersion: "v1",
				Kind:       "Deployment",
				Name:       "currencyservice",
			},
			MinReplicas:         1,
			MaxReplicas:         10,
			TargetCPUPercentage: 70,
		},
		{
			Namespace: "default",
			Name:      "currencyservice-hpa-2",
			TargetRef: k8s.TargetRef{
				APIVersion: "v1",
				Kind:       "Deployment",
				Name:       "currencyservice",
			},
			MinReplicas:         1,
			MaxReplicas:         10,
			TargetCPUPercentage: 80,
		},
		{
			Namespace: "default",
			Name:      "currencyservice-hpa-other",
			TargetRef: k8s.TargetRef{
				APIVersion: "v1",
				Kind:       "Deployment",
				Name:       "currencyservice2",
			},
			MinReplicas:         1,
			MaxReplicas:         10,
			TargetCPUPercentage: 0,
		},
	}

	now := time.Now().Format(time.RFC3339)
	tsList := BuildHPACPUTargetUtilizationTimeSeries(hpas, now)

	expectedI := 1
	if got := len(tsList); got != expectedI {
		t.Errorf("Expected # %+v, got %+v", expectedI, got)
	}

	ts := tsList[0]
	expected := hpaCPUTargetUtilizationMetricType
	if got := ts.Metric.Type; got != expected {
		t.Errorf("Expected Metric %+v, got %+v", expected, got)
	}

	expected = "currencyservice-hpa-1"
	if got := ts.Metric.Labels["object_name"]; got != expected {
		t.Errorf("Expected Label %+v, got %+v", expected, got)
	}
}
