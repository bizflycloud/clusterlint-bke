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

package bke

import (
	"github.com/bizflycloud/clusterlint/checks"
	"github.com/bizflycloud/clusterlint/kube"
)

func init() {
	checks.Register(&invalidSnapshotCheck{})
}

type invalidSnapshotCheck struct{}

// Name returns a unique name for this check.
func (i *invalidSnapshotCheck) Name() string {
	return "invalid-volume-snapshot"
}

// Groups returns a list of group names this check should be part of.
func (i *invalidSnapshotCheck) Groups() []string {
	return []string{"bke"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (i *invalidSnapshotCheck) Description() string {
	return "Checks if there are invalid volume snapshots that would fail webhook validation"
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (i *invalidSnapshotCheck) Run(objects *kube.Objects) ([]checks.Diagnostic, error) {
	var diagnostics []checks.Diagnostic
	errMsg := "Snapshot has been marked as invalid by CSI validation - check persistentVolumeClaimName and volumeSnapshotContentName are not both set"
	ssLabelKey := "snapshot.storage.kubernetes.io/invalid-snapshot-resource"
	for _, snapshot := range objects.VolumeSnapshotsV1.Items {
		snapshotLabels := snapshot.Labels
		if _, ok := snapshotLabels[ssLabelKey]; ok {
			d := checks.Diagnostic{
				Severity: checks.Error,
				Message:  errMsg,
				Kind:     checks.VolumeSnapshot,
				Object:   &snapshot.ObjectMeta,
				Owners:   snapshot.ObjectMeta.GetOwnerReferences(),
			}
			diagnostics = append(diagnostics, d)
		}
	}
	for _, snapshot := range objects.VolumeSnapshotsBeta.Items {
		snapshotLabels := snapshot.Labels
		if _, ok := snapshotLabels[ssLabelKey]; ok {
			d := checks.Diagnostic{
				Severity: checks.Error,
				Message:  errMsg,
				Kind:     checks.VolumeSnapshot,
				Object:   &snapshot.ObjectMeta,
				Owners:   snapshot.ObjectMeta.GetOwnerReferences(),
			}
			diagnostics = append(diagnostics, d)
		}
	}
	return diagnostics, nil
}
