/*
Copyright 2022 bizflycloud

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package basic

import (
	"fmt"
	"strings"
	"github.com/bizflycloud/clusterlint/checks"
	"github.com/bizflycloud/clusterlint/kube"
)

func init() {
	checks.Register(&hostPathCheck{})
}

type hostPathCheck struct{}

// Name returns a unique name for this check.
func (h *hostPathCheck) Name() string {
	return "hostpath-volume"
}

// Groups returns a list of group names this check should be part of.
func (h *hostPathCheck) Groups() []string {
	return []string{"basic", "bke"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (h *hostPathCheck) Description() string {
	return "Check if there are pods using hostpath volumes"
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (h *hostPathCheck) Run(objects *kube.Objects) ([]checks.Diagnostic, error) {
	bkePod := [...]string{"csi-bizflycloud-nodeplugin","kube-router"}
	var diagnostics []checks.Diagnostic
	for _, pod := range objects.Pods.Items {
		for _,podName := range bkePod {
			if strings.Contains(pod.Name, podName) {
				for _, volume := range pod.Spec.Volumes {
					pod := pod
					if volume.VolumeSource.HostPath != nil {
						d := checks.Diagnostic{
							Severity: checks.Warning,
							Message:  fmt.Sprintf("Avoid using hostpath for volume '%s'.", volume.Name),
							Kind:     checks.Pod,
							Object:   &pod.ObjectMeta,
							Owners:   pod.ObjectMeta.GetOwnerReferences(),
						}
						diagnostics = append(diagnostics, d)
					}
				}
			}
		}
	}

	return diagnostics, nil
}
