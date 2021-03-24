# Copyright 2021 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

displayName: GKE - App Right Sizing (CLUSTER_TO_REPLACE:{{- (index .items 0).metadata.namespace -}})
mosaicLayout:
  columns: 12
  tiles:

  - widget:
      title: 'CPU: Namespace over-provisioning'
      xyChart:
        chartOptions:
          mode: COLOR
        dataSets:
        - plotType: LINE
          timeSeriesQuery:
            timeSeriesQueryLanguage: |-
                { namespace_request_cores:
                    fetch k8s_container
                    | metric 'kubernetes.io/container/cpu/request_cores'
                    | filter
                        (resource.cluster_name == 'CLUSTER_TO_REPLACE'
                        && resource.namespace_name == '{{- (index .items 0).metadata.namespace -}}')
                    | group_by 1m, [value_request_cores_mean: mean(value.request_cores)]
                    | every 1m
                    | group_by [cluster_name: resource.cluster_name],
                        [value: sum(value_request_cores_mean)]

                ; namespace_recommended_cores:
                    { recommendation:
                        { vpa:
                          fetch k8s_container
                          | metric 'custom.googleapis.com/podautoscaler/vpa/cpu/target_recommendation'
                          | filter 
                              (resource.cluster_name == 'CLUSTER_TO_REPLACE'
                              && resource.namespace_name == '{{- (index .items 0).metadata.namespace -}}')
                          | group_by 1m, [value_cpu_mean: mean(value.target_recommendation)]
                          | every 1m
                          | group_by [cluster_name: resource.cluster_name, kind:metric.targetref_kind, controller_name:metric.targetref_name],
                          		[recommendation: sum(value_cpu_mean)]
                        ; hpa:
                          fetch k8s_pod
                          | metric 'custom.googleapis.com/podautoscaler/hpa/cpu/target_utilization'
                          | filter 
                              (resource.cluster_name == 'CLUSTER_TO_REPLACE'
                              && resource.namespace_name == '{{- (index .items 0).metadata.namespace -}}')
                          | group_by 1m, [value_cpu_mean: mean(value.target_utilization)]
                          | every 1m
                          | group_by [cluster_name: resource.cluster_name, kind:metric.targetref_kind, controller_name:metric.targetref_name],
                          		[target: mean(value_cpu_mean)]
                        }
                        | outer_join [0],[0]
                        | value [recommendation: vpa.recommendation + (if( gt(hpa.target,0),  vpa.recommendation * (100 - hpa.target) / 100,  cast_double(0) )) ]
                        | group_by [cluster_name, controller_name], 
                            [per_controler_name: cast_units(mean(recommendation), "{cpu}")]
                    
                    ; number_of_pods:
                        fetch k8s_pod
                        | metric 'kubernetes.io/pod/volume/total_bytes'
                        | filter
                            (resource.cluster_name == 'CLUSTER_TO_REPLACE'
                            && resource.namespace_name == '{{- (index .items 0).metadata.namespace -}}')
                        | group_by 1m, [mean(val())]
                        | every 1m
                        | group_by [cluster_name:resource.cluster_name, controller_name:metadata.system_labels.top_level_controller_name],
                                [per_controller_name:count(val())]
                    }
                    | join
                    | group_by [cluster_name, controller_name],
                        [recommendation:sum(recommendation.per_controler_name * number_of_pods.per_controller_name)]
                    | group_by [cluster_name],
                        [value:sum(recommendation)]
                }
                | join
                | value [apps_overprovisioned_perc:cast_units(100 - (100 * namespace_recommended_cores.value / namespace_request_cores.value), '%')]
        timeshiftDuration: 0s
        yAxis:
          label: y1Axis
          scale: LINEAR
    height: 4    
    width: 6


  - widget:
      title: 'Memory: Namespace over-provisioning'
      xyChart:
        chartOptions:
          mode: COLOR
        dataSets:
        - plotType: LINE
          timeSeriesQuery:
            timeSeriesQueryLanguage: |-
                { namespace_request_bytes:
                    fetch k8s_container
                    | metric 'kubernetes.io/container/memory/request_bytes'
                    | filter
                        (resource.cluster_name == 'CLUSTER_TO_REPLACE'
                        && resource.namespace_name == '{{- (index .items 0).metadata.namespace -}}')
                    | group_by 1m, [value_request_bytes_mean: mean(value.request_bytes)]
                    | every 1m
                    | group_by [cluster_name: resource.cluster_name],
                        [value: sum(value_request_bytes_mean)]

                ; namespace_recommended_bytes:
                
                    { recommendation:
                        fetch k8s_container
                        | metric 'custom.googleapis.com/podautoscaler/vpa/memory/target_recommendation'
                        | filter
                            (resource.cluster_name == 'CLUSTER_TO_REPLACE'
                            && resource.namespace_name == '{{- (index .items 0).metadata.namespace -}}')
                        | group_by 1m,
                            [value_vpa_recommendation_mean: mean(value.target_recommendation)]
                        | every 1m
                        | group_by [cluster_name: resource.cluster_name, controller_name:resource.pod_name],
                            [per_controller_name: cast_units(sum(value_vpa_recommendation_mean), "By")]
                    
                    ; number_of_pods:
                        fetch k8s_pod
                        | metric 'kubernetes.io/pod/volume/total_bytes'
                        | filter
                            (resource.cluster_name == 'CLUSTER_TO_REPLACE'
                            && resource.namespace_name == '{{- (index .items 0).metadata.namespace -}}')
                        | group_by 1m, [mean(val())]
                        | every 1m
                        | group_by [cluster_name:resource.cluster_name, controller_name:metadata.system_labels.top_level_controller_name],
                            [per_controller_name:count(val())]
                    }
                    | join
                    | group_by [cluster_name, controller_name],
                        [recommendation:sum(recommendation.per_controller_name * number_of_pods.per_controller_name)]
                    | group_by [cluster_name],
                        [value:sum(recommendation)]
                        
                }
                | join
                | value [apps_overprovisioned_perc:cast_units(100 - (100 * namespace_recommended_bytes.value / namespace_request_bytes.value), '%')]
        timeshiftDuration: 0s
        yAxis:
          label: y1Axis
          scale: LINEAR
    width: 6
    height: 4
    xPos: 6


  - widget:
      title: 'CPU: Top 5 over-provisioned apps (in cores)'
      xyChart:
        chartOptions:
          mode: COLOR
        dataSets:
        - plotType: LINE
          timeSeriesQuery:
            timeSeriesQueryLanguage: |-
                { namespace_request_cores:
                    fetch k8s_container
                    | metric 'kubernetes.io/container/cpu/request_cores'
                    | filter
                        (resource.cluster_name == 'CLUSTER_TO_REPLACE'
                        && resource.namespace_name == '{{- (index .items 0).metadata.namespace -}}')
                    | group_by 1m, [value_request_cores_mean: mean(value.request_cores)]
                    | every 1m
                    | group_by [cluster_name: resource.cluster_name, controller_name:metadata.system_labels.top_level_controller_name],
                        [value: sum(value_request_cores_mean)]

                ; namespace_recommended_cores:

                    { recommendation:
                        { vpa:
                          fetch k8s_container
                          | metric 'custom.googleapis.com/podautoscaler/vpa/cpu/target_recommendation'
                          | filter 
                              (resource.cluster_name == 'CLUSTER_TO_REPLACE'
                              && resource.namespace_name == '{{- (index .items 0).metadata.namespace -}}')
                          | group_by 1m, [value_cpu_mean: mean(value.target_recommendation)]
                          | every 1m
                          | group_by [cluster_name: resource.cluster_name, kind:metric.targetref_kind, controller_name:metric.targetref_name],
                          		[recommendation: sum(value_cpu_mean)]
                        ; hpa:
                          fetch k8s_pod
                          | metric 'custom.googleapis.com/podautoscaler/hpa/cpu/target_utilization'
                          | filter 
                              (resource.cluster_name == 'CLUSTER_TO_REPLACE'
                              && resource.namespace_name == '{{- (index .items 0).metadata.namespace -}}')
                          | group_by 1m, [value_cpu_mean: mean(value.target_utilization)]
                          | every 1m
                          | group_by [cluster_name: resource.cluster_name, kind:metric.targetref_kind, controller_name:metric.targetref_name],
                          		[target: mean(value_cpu_mean)]
                        }
                        | outer_join [0],[0]
                        | value [recommendation: vpa.recommendation + (if( gt(hpa.target,0),  vpa.recommendation * (100 - hpa.target) / 100,  cast_double(0) )) ]
                        | group_by [cluster_name, controller_name], 
                            [per_controler_name: cast_units(mean(recommendation), "{cpu}")]
                    
                    ; number_of_pods:
                        fetch k8s_pod
                        | metric 'kubernetes.io/pod/volume/total_bytes'
                        | filter
                            (resource.cluster_name == 'CLUSTER_TO_REPLACE'
                            && resource.namespace_name == '{{- (index .items 0).metadata.namespace -}}')
                        | group_by 1m, [mean(val())]
                        | every 1m
                        | group_by [cluster_name:resource.cluster_name, controller_name:metadata.system_labels.top_level_controller_name],
                                [per_controller_name:count(val())]
                    }
                    | join
                    | group_by [cluster_name, controller_name],
                        [value:sum(recommendation.per_controler_name * number_of_pods.per_controller_name)]

                }
                | join
                #| value [app_overprovisioned_perc:cast_units(100 - (100 * namespace_recommended_cores.value / namespace_request_cores.value), '%')]
                | value [app_overprovisioned_in_num_cores:namespace_request_cores.value - namespace_recommended_cores.value]
                | top 5
        timeshiftDuration: 0s
        yAxis:
          label: y1Axis
          scale: LINEAR
    width: 6
    height: 4    
    yPos: 4

  - widget:
      title: 'Memory: Top 5 over-provisioned apps (in bytes)'
      xyChart:
        chartOptions:
          mode: COLOR
        dataSets:
        - plotType: LINE
          timeSeriesQuery:
            timeSeriesQueryLanguage: |-
                { namespace_request_bytes:
                    fetch k8s_container
                    | metric 'kubernetes.io/container/memory/request_bytes'
                    | filter
                        (resource.cluster_name == 'CLUSTER_TO_REPLACE'
                        && resource.namespace_name == '{{- (index .items 0).metadata.namespace -}}')
                    | group_by 1m, [value_request_bytes_mean: mean(value.request_bytes)]
                    | every 1m
                    | group_by [cluster_name: resource.cluster_name, controller_name:metadata.system_labels.top_level_controller_name],
                        [value: sum(value_request_bytes_mean)]

                ; namespace_recommended_bytes:
                
                    { recommendation:
                        fetch k8s_container
                        | metric 'custom.googleapis.com/podautoscaler/vpa/memory/target_recommendation'
                        | filter
                            (resource.cluster_name == 'CLUSTER_TO_REPLACE'
                            && resource.namespace_name == '{{- (index .items 0).metadata.namespace -}}')
                        | group_by 1m,
                            [value_vpa_recommendation_mean: mean(value.target_recommendation)]
                        | every 1m
                        | group_by [cluster_name: resource.cluster_name, controller_name:resource.pod_name],
                            [per_controller_name: cast_units(sum(value_vpa_recommendation_mean), "By")]
                    
                    ; number_of_pods:
                        fetch k8s_pod
                        | metric 'kubernetes.io/pod/volume/total_bytes'
                        | filter
                            (resource.cluster_name == 'CLUSTER_TO_REPLACE'
                            && resource.namespace_name == '{{- (index .items 0).metadata.namespace -}}')
                        | group_by 1m, [mean(val())]
                        | every 1m
                        | group_by [cluster_name:resource.cluster_name, controller_name:metadata.system_labels.top_level_controller_name],
                            [per_controller_name:count(val())]
                    }
                    | join
                    | group_by [cluster_name, controller_name],
                        [value:sum(recommendation.per_controller_name * number_of_pods.per_controller_name)]

                }
                | join
                #| value [app_overprovisioned_perc:cast_units(100 - (100 * namespace_recommended_bytes.value / namespace_request_bytes.value), '%')]
                | value [app_overprovisioned_in_bytes:namespace_request_bytes.value - namespace_recommended_bytes.value]
                | top 5
        timeshiftDuration: 0s
        yAxis:
          label: y1Axis
          scale: LINEAR
    width: 6
    height: 4    
    xPos: 6
    yPos: 4


  - widget:
      title: 'CPU: Top 5 under-provisioned apps (%)'
      xyChart:
        chartOptions:
          mode: COLOR
        dataSets:
        - plotType: LINE
          timeSeriesQuery:
            timeSeriesQueryLanguage: |-
                { namespace_request_cores:
                    fetch k8s_container
                    | metric 'kubernetes.io/container/cpu/request_cores'
                    | filter
                        (resource.cluster_name == 'CLUSTER_TO_REPLACE'
                        && resource.namespace_name == '{{- (index .items 0).metadata.namespace -}}')
                    | group_by 1m, [value_request_cores_mean: mean(value.request_cores)]
                    | every 1m
                    | group_by [cluster_name: resource.cluster_name, controller_name:metadata.system_labels.top_level_controller_name],
                        [value: sum(value_request_cores_mean)]

                ; namespace_recommended_cores:

                    { recommendation:
                        { vpa:
                          fetch k8s_container
                          | metric 'custom.googleapis.com/podautoscaler/vpa/cpu/target_recommendation'
                          | filter 
                              (resource.cluster_name == 'CLUSTER_TO_REPLACE'
                              && resource.namespace_name == '{{- (index .items 0).metadata.namespace -}}')
                          | group_by 1m, [value_cpu_mean: mean(value.target_recommendation)]
                          | every 1m
                          | group_by [cluster_name: resource.cluster_name, kind:metric.targetref_kind, controller_name:metric.targetref_name],
                          		[recommendation: sum(value_cpu_mean)]
                        ; hpa:
                          fetch k8s_pod
                          | metric 'custom.googleapis.com/podautoscaler/hpa/cpu/target_utilization'
                          | filter 
                              (resource.cluster_name == 'CLUSTER_TO_REPLACE'
                              && resource.namespace_name == '{{- (index .items 0).metadata.namespace -}}')
                          | group_by 1m, [value_cpu_mean: mean(value.target_utilization)]
                          | every 1m
                          | group_by [cluster_name: resource.cluster_name, kind:metric.targetref_kind, controller_name:metric.targetref_name],
                          		[target: mean(value_cpu_mean)]
                        }
                        | outer_join [0],[0]
                        | value [recommendation: vpa.recommendation + (if( gt(hpa.target,0),  vpa.recommendation * (100 - hpa.target) / 100,  cast_double(0) )) ]
                        | group_by [cluster_name, controller_name], 
                            [per_controler_name: cast_units(mean(recommendation), "{cpu}")]
                    
                    ; number_of_pods:
                        fetch k8s_pod
                        | metric 'kubernetes.io/pod/volume/total_bytes'
                        | filter
                            (resource.cluster_name == 'CLUSTER_TO_REPLACE'
                            && resource.namespace_name == '{{- (index .items 0).metadata.namespace -}}')
                        | group_by 1m, [mean(val())]
                        | every 1m
                        | group_by [cluster_name:resource.cluster_name, controller_name:metadata.system_labels.top_level_controller_name],
                                [per_controller_name:count(val())]
                    }
                    | join
                    | group_by [cluster_name, controller_name],
                        [value:sum(recommendation.per_controler_name * number_of_pods.per_controller_name)]

                }
                | join
                | value [app_overprovisioned_perc:cast_units(100 - (100 * namespace_recommended_cores.value / namespace_request_cores.value), '%')]
                | filter (app_overprovisioned_perc < 0)
                | bottom 5
        timeshiftDuration: 0s
        yAxis:
          label: y1Axis
          scale: LINEAR
    width: 6
    height: 4    
    yPos: 8

  - widget:
      title: 'Memory: Top 5 under-provisioned apps (%)'
      xyChart:
        chartOptions:
          mode: COLOR
        dataSets:
        - plotType: LINE
          timeSeriesQuery:
            timeSeriesQueryLanguage: |-
                { namespace_request_bytes:
                    fetch k8s_container
                    | metric 'kubernetes.io/container/memory/request_bytes'
                    | filter
                        (resource.cluster_name == 'CLUSTER_TO_REPLACE'
                        && resource.namespace_name == '{{- (index .items 0).metadata.namespace -}}')
                    | group_by 1m, [value_request_bytes_mean: mean(value.request_bytes)]
                    | every 1m
                    | group_by [cluster_name: resource.cluster_name, controller_name:metadata.system_labels.top_level_controller_name],
                        [value: sum(value_request_bytes_mean)]

                ; namespace_recommended_bytes:
                
                    { recommendation:
                        fetch k8s_container
                        | metric 'custom.googleapis.com/podautoscaler/vpa/memory/target_recommendation'
                        | filter
                            (resource.cluster_name == 'CLUSTER_TO_REPLACE'
                            && resource.namespace_name == '{{- (index .items 0).metadata.namespace -}}')
                        | group_by 1m,
                            [value_vpa_recommendation_mean: mean(value.target_recommendation)]
                        | every 1m
                        | group_by [cluster_name: resource.cluster_name, controller_name:resource.pod_name],
                            [per_controller_name: cast_units(sum(value_vpa_recommendation_mean), "By")]
                    
                    ; number_of_pods:
                        fetch k8s_pod
                        | metric 'kubernetes.io/pod/volume/total_bytes'
                        | filter
                            (resource.cluster_name == 'CLUSTER_TO_REPLACE'
                            && resource.namespace_name == '{{- (index .items 0).metadata.namespace -}}')
                        | group_by 1m, [mean(val())]
                        | every 1m
                        | group_by [cluster_name:resource.cluster_name, controller_name:metadata.system_labels.top_level_controller_name],
                            [per_controller_name:count(val())]
                    }
                    | join
                    | group_by [cluster_name, controller_name],
                        [value:sum(recommendation.per_controller_name * number_of_pods.per_controller_name)]

                }
                | join
                | value [app_overprovisioned_perc:cast_units(100 - (100 * namespace_recommended_bytes.value / namespace_request_bytes.value), '%')]                
                | filter (app_overprovisioned_perc < 0)
                | bottom 5
        timeshiftDuration: 0s
        yAxis:
          label: y1Axis
          scale: LINEAR
    width: 6
    height: 4    
    xPos: 6
    yPos: 8


{{- range $i, $d := .items -}}
  {{- if le $i  8 -}}
    {{- /*This condition is required because Dashboard widget maximum allowed is 40*/ -}}
    {{- $namespace := $d.metadata.namespace -}}
    {{- $controllerName := $d.metadata.name }}

  - widget:
      title: 'CPU: {{$controllerName}} (p/ Pod)'
      xyChart:
        chartOptions:
          mode: COLOR
        dataSets:
        - plotType: LINE
          timeSeriesQuery:
            timeSeriesQueryLanguage: |-
              { t_0:
                  fetch k8s_container
                  | metric 'kubernetes.io/container/cpu/request_cores'
                  | filter
                      (resource.cluster_name = 'CLUSTER_TO_REPLACE'
                      && resource.namespace_name = '{{$namespace}}'
                      && metadata.system_labels.top_level_controller_name = '{{$controllerName}}')
                  | group_by 1m, [value_request_cores_mean: mean(value.request_cores)]
                  | every 1m
                  | group_by [pod_name: resource.pod_name],
                      [value_request_cores_mean_sum: sum(value_request_cores_mean)]
                  | group_by [metric: 'request_cores'],
                      [avg_value: mean(value_request_cores_mean_sum)]
              ; t_1:
                  fetch k8s_container
                  | metric 'kubernetes.io/container/cpu/limit_cores'
                  | filter
                      (resource.cluster_name = 'CLUSTER_TO_REPLACE'
                      && resource.namespace_name = '{{$namespace}}'
                      && metadata.system_labels.top_level_controller_name = '{{$controllerName}}')
                  | group_by 1m, [value_limit_cores_mean: mean(value.limit_cores)]
                  | every 1m
                  | group_by [pod_name: resource.pod_name],
                      [value_limit_cores_mean_sum: sum(value_limit_cores_mean)]
                  | group_by [metric: 'limit_cores'],
                      [avg_value: mean(value_limit_cores_mean_sum)]
              ; t_2:
                  fetch k8s_container
                  | metric 'kubernetes.io/container/cpu/core_usage_time'
                  | filter
                      (resource.cluster_name = 'CLUSTER_TO_REPLACE'
                      && resource.namespace_name = '{{$namespace}}'
                      && metadata.system_labels.top_level_controller_name = '{{$controllerName}}')
                  | align rate(1m)
                  | every 1m
                  | group_by [pod_name: resource.pod_name],
                      [value_core_usage_time_sum: sum(value.core_usage_time)]
                  | group_by [metric: 'used_cores'],
                      [avg_value: mean(value_core_usage_time_sum)]
              ; recommendation:
                  { vpa:
                    fetch k8s_container
                    | metric 'custom.googleapis.com/podautoscaler/vpa/cpu/target_recommendation'
                    | filter 
                        (resource.cluster_name == 'CLUSTER_TO_REPLACE'
                        && resource.namespace_name == '{{$namespace}}'
                		&& metric.targetref_name = '{{$controllerName}}')
                    | group_by 1m, [value_cpu_mean: mean(value.target_recommendation)]
                    | every 1m
                    | group_by [cluster_name: resource.cluster_name, kind:metric.targetref_kind, controller_name:metric.targetref_name],
                    		[recommendation: sum(value_cpu_mean)]
                  ; hpa:
                    fetch k8s_pod
                    | metric 'custom.googleapis.com/podautoscaler/hpa/cpu/target_utilization'
                    | filter 
                        (resource.cluster_name == 'CLUSTER_TO_REPLACE'
                        && resource.namespace_name == '{{$namespace}}'
                		&& metric.targetref_name = '{{$controllerName}}')                              
                    | group_by 1m, [value_cpu_mean: mean(value.target_utilization)]
                    | every 1m
                    | group_by [cluster_name: resource.cluster_name, kind:metric.targetref_kind, controller_name:metric.targetref_name],
                    		[target: mean(value_cpu_mean)]
                  }
                  | outer_join [0],[0]
                  | value [recommendation: vpa.recommendation + (if( gt(hpa.target,0),  vpa.recommendation * (100 - hpa.target) / 100,  cast_double(0) )) ]
                  | group_by [metric: 'recommended_cores'], 
                      [avg_value: mean(recommendation)]
              ;
                  { requested:
                    fetch k8s_container
                    | metric 'kubernetes.io/container/cpu/request_cores'
                    | filter
                        (resource.cluster_name = 'CLUSTER_TO_REPLACE'
                        && resource.namespace_name = '{{$namespace}}'
                        && metadata.system_labels.top_level_controller_name = '{{$controllerName}}')
                    | group_by 1m, [value_request_cores_mean: mean(value.request_cores)]
                    | every 1m
                    | group_by [pod_name: resource.pod_name],
                        [value_request_cores_mean_sum: sum(value_request_cores_mean)]
                    | group_by [metric: 'request_cores'],
                        [value: mean(value_request_cores_mean_sum)]
                  ; hpa:
                    fetch k8s_pod
                    | metric 'custom.googleapis.com/podautoscaler/hpa/cpu/target_utilization'
                    | filter 
                        (resource.cluster_name == 'CLUSTER_TO_REPLACE'
                        && resource.namespace_name == '{{$namespace}}'
                				&& metric.targetref_name = '{{$controllerName}}')                              
                    | group_by 1m, [value_cpu_mean: mean(value.target_utilization)]
                    | every 1m
                    | group_by [cluster_name: resource.cluster_name, controller_name:metric.targetref_name],
                    		[target: cast_units(mean(value_cpu_mean), '{cpu}')]
                }
                | join 
                | value [hpa_target: requested.value * (hpa.target/100)]
                | group_by [metric: 'hpa_target_utilizaiton'],
                        [avg_value: mean(hpa_target)]                     
              }
              | union
        timeshiftDuration: 0s
        yAxis:
            label: y1Axis
            scale: LINEAR
    width: 4
    height: 4    
    yPos: Y_POS_TO_REPLACE_{{$i}}


  - widget:
      title: 'Mem: {{$controllerName}} (p/ Pod)'
      xyChart:
        chartOptions:
          mode: COLOR
        dataSets:
        - plotType: LINE
          timeSeriesQuery:
            timeSeriesQueryLanguage: |-
              { t_0:
                  fetch k8s_container
                  | metric 'kubernetes.io/container/memory/request_bytes'
                  | filter
                      (resource.cluster_name = 'CLUSTER_TO_REPLACE'
                      && resource.namespace_name = '{{$namespace}}'
                      && metadata.system_labels.top_level_controller_name = '{{$controllerName}}')
                  | group_by 1m, [value_request_bytes_mean: mean(value.request_bytes)]
                  | every 1m
                  | group_by [pod_name: resource.pod_name],
                      [value_request_bytes_mean_sum: sum(value_request_bytes_mean)]
                  | group_by [metric: 'request_bytes'],
                      [avg_value: mean(value_request_bytes_mean_sum)]
              ; t_1:
                  fetch k8s_container
                  | metric 'kubernetes.io/container/memory/limit_bytes'
                  | filter
                      (resource.cluster_name = 'CLUSTER_TO_REPLACE'
                      && resource.namespace_name = '{{$namespace}}'
                      && metadata.system_labels.top_level_controller_name = '{{$controllerName}}')
                  | group_by 1m, [value_limit_bytes_mean: mean(value.limit_bytes)]
                  | every 1m
                  | group_by [pod_name: resource.pod_name],
                      [value_limit_bytes_mean_sum: sum(value_limit_bytes_mean)]
                  | group_by [metric: 'limit_bytes'],
                      [avg_value: mean(value_limit_bytes_mean_sum)]
              ; t_2:
                  fetch k8s_container
                  | metric 'kubernetes.io/container/memory/used_bytes'
                  | filter
                      (resource.cluster_name = 'CLUSTER_TO_REPLACE'
                      && resource.namespace_name = '{{$namespace}}'
                      && metadata.system_labels.top_level_controller_name = '{{$controllerName}}'
                      && metric.memory_type = 'non-evictable')
                  | group_by 1m, [value_used_bytes_mean: mean(value.used_bytes)]
                  | every 1m
                  | group_by [pod_name: resource.pod_name],
                      [value_used_bytes_mean_sum: sum(value_used_bytes_mean)]
                  | group_by [metric: 'used_bytes'],
                      [avg_value: max(value_used_bytes_mean_sum)]
              ; recommendation:
                  fetch k8s_container
                  | metric 'custom.googleapis.com/podautoscaler/vpa/memory/target_recommendation'
                  | filter
                      (resource.cluster_name = 'CLUSTER_TO_REPLACE'
                      && resource.namespace_name = '{{$namespace}}'
                      && metric.targetref_name = '{{$controllerName}}')
                  | group_by 1m,
                      [value_vpa_recommendation_mean: mean(value.target_recommendation)]
                  | every 1m
                  | group_by [metric: 'recommended_bytes'],
                      [avg_value: cast_double(sum(value_vpa_recommendation_mean))] }
              | union
        timeshiftDuration: 0s
        yAxis:
            label: y1Axis
            scale: LINEAR
    width: 4
    height: 4    
    yPos: Y_POS_TO_REPLACE_{{$i}}
    xPos: 4

  - widget:
      title: 'Replicas: {{$controllerName}}'
      xyChart:
        chartOptions:
          mode: COLOR
        dataSets:
        - plotType: LINE
          timeSeriesQuery:
              timeSeriesQueryLanguage: |-
                  fetch k8s_pod
                  | metric 'kubernetes.io/pod/volume/total_bytes'
                  | filter
                          (resource.cluster_name == 'CLUSTER_TO_REPLACE'
                          && resource.namespace_name == '{{$namespace}}')
                  | filter (metadata.system_labels.top_level_controller_name == '{{$controllerName}}')
                  | group_by 1m, [mean(val())]
                  | every 1m
                  | group_by [resource.cluster_name], count(val())
        timeshiftDuration: 0s
        yAxis:
          label: y1Axis
          scale: LINEAR
    width: 4
    height: 4    
    yPos: Y_POS_TO_REPLACE_{{$i}}
    xPos: 8

  {{- end -}}

{{- end -}}

