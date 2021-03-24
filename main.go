/**
 * Copyright 2020 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package main

import (
	"flag"
	"fmt"
	"metrics-exporter/apis/k8s"
	"metrics-exporter/apis/mon"
	"os"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
)

const version = "v0.0.1"

var (
	verbosity = flag.String("v", "info", "[Optional] Verbosity: panic|fatal|error|warn|info|debug|trace")
)

func init() {
	flag.Parse()

	level, err := log.ParseLevel(*verbosity)
	exitOnError("Invalid 'verbose' parameter", err)
	log.SetFormatter(&log.JSONFormatter{
		DisableTimestamp: true,
		FieldMap: log.FieldMap{
			log.FieldKeyLevel: "severity",
		},
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(level)
}

func main() {
	log.Infof("************** METRICS EXPORTER STARTED (version %s) **************", version)

	now := time.Now().Format(time.RFC3339)

	vpas := retrieveVPAs()
	tsList := mon.BuildVPARecommendationTimeSeries(vpas, now)

	hpas := retrieveHPAs()
	tsList = append(tsList, mon.BuildHPACPUTargetUtilizationTimeSeries(hpas, now)...)

	err := mon.ExportMetrics(tsList)
	exitOnError("Failed to instantiate cloud monitoring object", err)

	log.Infof("************** METRICS EXPORTER FINISHED *************")
}

func retrieveVPAs() []k8s.VPA {
	cmd := "kubectl get vpa --all-namespaces -o yaml"
	out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	data := string(out)
	exitOnError(fmt.Sprintf("Failed to execute command: %s\nRoot Cause: %+v", cmd, data), err)

	vpas, err := k8s.DecodeVPAList([]byte(data))
	exitOnError(fmt.Sprintf("Failed to decode VPA list from command: %s", cmd), err)
	return vpas
}

func retrieveHPAs() []k8s.HPA {
	cmd := "kubectl get hpa --all-namespaces -o yaml"
	out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	data := string(out)
	exitOnError(fmt.Sprintf("Failed to execute command: %s\nRoot Cause: %+v", cmd, data), err)

	hpas, err := k8s.DecodeHPAList([]byte(data))
	exitOnError(fmt.Sprintf("Failed to decode HPA list from command: %s", cmd), err)
	return hpas
}

func exitOnError(message string, err error) {
	if err != nil {
		exitWithError(message, err)
	}
}

func exitWithError(message string, err error) {
	fmt.Printf("\nError: %s\nCause: %+v\n\nSee parameters options below:\n", err, message)
	flag.PrintDefaults()
	os.Exit(-1)
}
