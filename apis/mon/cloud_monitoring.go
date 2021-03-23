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
	"context"
	"fmt"
	"strings"
	"time"

	gce "cloud.google.com/go/compute/metadata"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/monitoring/v3"
)

const (
	maxRetries     = 3
	retrialBackOff = 350
)

// ExportMetrics Export the timeseries list to cloud monitoring
func ExportMetrics(tsList []*monitoring.TimeSeries) error {
	service, err := monitoring.New(oauth2.NewClient(context.Background(), google.ComputeTokenSource("")))
	if err != nil {
		return err
	}

	targetSize := len(tsList)
	if targetSize > 0 {
		tsChunks := timeSeriesToChunks(tsList, 200) // break into chunks once API accept only 200 per time
		for _, tsChunk := range tsChunks {
			log.Infof("Exporting chunk with %v metrics", len(tsChunk))
			err := exportChunk(service, tsChunk, 1)
			if err != nil {
				log.Errorf(fmt.Sprintf("Failed to export chunk with %v metrics. Cause: %+v", len(tsChunk), err))
				logTimeSeriesList(tsChunk)
			} else {
				log.Infof("Success exporting chunk with %v metrics", len(tsChunk))
			}
		}

	}
	return nil
}

func exportChunk(service *monitoring.Service, tsChunk []*monitoring.TimeSeries, retries int) error {
	projectID, _ := gce.ProjectID()
	project := "projects/" + projectID
	request := &monitoring.CreateTimeSeriesRequest{TimeSeries: tsChunk}
	_, err := service.Projects.TimeSeries.Create(project, request).Do()
	if err != nil {
		log.Warnf(fmt.Sprintf("[Trial %v] Failed to export chunk with %v metrics. Cause: %+v", retries, len(tsChunk), err))
		if retries <= maxRetries && strings.Contains(err.Error(), "Error 500") {
			// incremental back off
			duration := time.Duration(retrialBackOff*retries) * time.Millisecond
			time.Sleep(duration)
			return exportChunk(service, tsChunk, retries+1)
		} else {
			return err
		}
	}
	return nil
}

func timeSeriesToChunks(tsList []*monitoring.TimeSeries, chunkSize int) [][]*monitoring.TimeSeries {
	chunks := [][]*monitoring.TimeSeries{}
	for len(tsList) > chunkSize {
		chunks = append(chunks, tsList[:chunkSize-1])
		tsList = tsList[chunkSize:]
	}
	if len(tsList) > 0 {
		chunks = append(chunks, tsList)
	}
	return chunks
}

func logTimeSeriesList(tsList []*monitoring.TimeSeries) {
	for tsIndex, v := range tsList {
		points := make([]float64, len(v.Points))
		for i, p := range v.Points {
			v := *p.Value
			if v.DoubleValue != nil {
				points[i] = *v.DoubleValue
			} else {
				points[i] = float64(*v.Int64Value)
			}
		}
		log.Errorf("TimeSeries[%v] => Metric: %s | Resource Labels: %v | Metric Labels: %v | Points: %+v",
			tsIndex, v.Metric.Type, v.Resource.Labels, v.Metric.Labels, points)
	}
}
