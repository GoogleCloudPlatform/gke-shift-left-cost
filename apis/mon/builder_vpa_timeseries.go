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
	"fmt"
	"metrics-exporter/apis/k8s"
	"strings"

	log "github.com/sirupsen/logrus"

	gce "cloud.google.com/go/compute/metadata"
	"google.golang.org/api/monitoring/v3"
)

const (
	vpaCpuMetricType    = "custom.googleapis.com/podautoscaler/vpa/cpu/target_recommendation"
	vpaMemoryMetricType = "custom.googleapis.com/podautoscaler/vpa/memory/target_recommendation"
)

// BuildVPARecommendationTimeSeries buid Timeseries objects for cpu and memory recommendations
func BuildVPARecommendationTimeSeries(vpas []k8s.VPA, now string) []*monitoring.TimeSeries {
	var vpaInRecomendationMode map[string]k8s.VPA = make(map[string]k8s.VPA)
	tsList := []*monitoring.TimeSeries{}
	for _, vpa := range vpas {
		if vpa.IsInRecomendationMode {
			targetKey := fmt.Sprintf("%s|%s|%s", vpa.TargetRef.Kind, vpa.Namespace, vpa.TargetRef.Name)
			if _, found := vpaInRecomendationMode[targetKey]; !found {
				vpaInRecomendationMode[targetKey] = vpa
				tsList = append(tsList, buildVPARecomendations(vpa, now)...)
			} else {
				// Skip VPA object once we alreay had one in recommendation mode for the same target object
				log.Infof("Skipping VPA '%s.%s' on recommendation mode once '%s.%s' was already loaded",
					vpa.Namespace, vpa.Name, vpaInRecomendationMode[targetKey].Namespace, vpaInRecomendationMode[targetKey].Name)
			}
		} else {
			log.Infof("Skipping VPA '%s.%s' once it is not on recommendation mode", vpa.Namespace, vpa.Name)
		}
	}
	return tsList
}

func buildVPARecomendations(vpa k8s.VPA, now string) []*monitoring.TimeSeries {
	tsList := []*monitoring.TimeSeries{}
	for _, rec := range vpa.Recomendations {
		resourceLabels := buildVPAResourceLabels(vpa, rec.ContainerName)
		// vpa cpu recommendation
		cpuTs := buildVPACPURecomendation(vpa, rec, resourceLabels, now)
		tsList = append(tsList, &cpuTs)
		// vpa memory recommendation
		memoryTs := buildVPAMemoryRecomendation(vpa, rec, resourceLabels, now)
		tsList = append(tsList, &memoryTs)
	}
	return tsList
}

func buildVPACPURecomendation(vpa k8s.VPA, rec k8s.VPARecomendation, resourceLabels map[string]string, now string) monitoring.TimeSeries {
	numberOfCores := float64(rec.Target.CPU)
	return monitoring.TimeSeries{
		Resource: &monitoring.MonitoredResource{
			Type:   "k8s_container",
			Labels: resourceLabels,
		},
		Metric: &monitoring.Metric{
			Type: vpaCpuMetricType,
			Labels: map[string]string{
				"targetef_apiversion": vpa.TargetRef.APIVersion,
				"targetref_kind":      vpa.TargetRef.Kind,
				"targetref_name":      vpa.TargetRef.Name,
				"object_name":         vpa.Name,
			},
		},
		Points: []*monitoring.Point{{
			Interval: &monitoring.TimeInterval{
				EndTime: now,
			},
			Value: &monitoring.TypedValue{
				DoubleValue: &numberOfCores,
			},
		}},
		Unit: "{cpu}", // TODO: understand why it is not being used
	}
}

func buildVPAMemoryRecomendation(vpa k8s.VPA, rec k8s.VPARecomendation, resourceLabels map[string]string, now string) monitoring.TimeSeries {
	memoryBytes := rec.Target.Memory
	return monitoring.TimeSeries{
		Resource: &monitoring.MonitoredResource{
			Type:   "k8s_container",
			Labels: resourceLabels,
		},
		Metric: &monitoring.Metric{
			Type: vpaMemoryMetricType,
			Labels: map[string]string{
				"targetef_apiversion": vpa.TargetRef.APIVersion,
				"targetref_kind":      vpa.TargetRef.Kind,
				"targetref_name":      vpa.TargetRef.Name,
				"object_name":         vpa.Name,
			},
		},
		Points: []*monitoring.Point{{
			Interval: &monitoring.TimeInterval{
				EndTime: now,
			},
			Value: &monitoring.TypedValue{
				Int64Value: &memoryBytes,
			},
		}},
		Unit: "By", // TODO: understand why it is not being used
	}
}

func buildVPAResourceLabels(vpa k8s.VPA, containerName string) map[string]string {
	projectID, _ := gce.ProjectID()
	location, _ := gce.InstanceAttributeValue("cluster-location")
	location = strings.TrimSpace(location)
	clusterName, _ := gce.InstanceAttributeValue("cluster-name")
	clusterName = strings.TrimSpace(clusterName)
	return map[string]string{
		"project_id":     projectID,
		"location":       location,
		"cluster_name":   clusterName,
		"namespace_name": vpa.Namespace,
		"pod_name":       vpa.TargetRef.Name,
		"container_name": containerName,
	}
}
