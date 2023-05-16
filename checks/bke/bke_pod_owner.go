
/*
Copyright 2021 DigitalOcean
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
	corev1 "k8s.io/api/core/v1"
	st "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	BizflyCloudCSIDriver        = "volume.csi.bizflycloud.vn"
	BizflyCloudBlockStorageName = "bizflycloud-s3"
)

func init() {
	checks.Register(&bkePodOwner{})
}

type bkePodOwner struct{}

// Name returns a unique name for this check.
func (p *bkePodOwner) Name() string {
	return "bke-pod-owner"
}

// Groups returns a list of group names this check should be part of.
func (p *bkePodOwner) Groups() []string {
	return []string{"bke"}
}

// Description returns a detailed human-readable description of what this check
// does.
func (p *bkePodOwner) Description() string {
	return "Checks if pods referencing bke volumes are owned by a stateful set."
}

// Run runs this check on a set of Kubernetes objects. It can return warnings
// (low-priority problems) and errors (high-priority problems) as well as an
// error value indicating that the check failed to run.
func (p *bkePodOwner) Run(objects *kube.Objects) ([]checks.Diagnostic, error) {
	var diagnostics []checks.Diagnostic
	var bkePods []corev1.Pod
	for _, pod := range objects.Pods.Items {
		pod := pod
		for _, volume := range pod.Spec.Volumes {
			if isBKEVolume(volume, pod.Namespace, objects) {
				bkePods = append(bkePods, pod)
			}
		}
	}
	for _, pod := range bkePods {
		pod := pod
		if pod.OwnerReferences != nil && ownedByStatefulSet(pod.OwnerReferences) {
			continue
		}
		d := checks.Diagnostic{
			Severity: checks.Warning,
			Message:  "Pod referencing BKE volumes must be owned by StatefulSet",
			Kind:     checks.Pod,
			Object:   &pod.ObjectMeta,
			Owners:   pod.ObjectMeta.GetOwnerReferences(),
		}
		diagnostics = append(diagnostics, d)
	}

	return diagnostics, nil
}

func isBKEVolume(volume corev1.Volume, namespace string, objects *kube.Objects) bool {
	if volume.PersistentVolumeClaim != nil {
		claim := getPVC(objects.PersistentVolumeClaims, volume.PersistentVolumeClaim.ClaimName, namespace)
		if claim == nil {
			return false
		}

		scn := getStorageClassName(claim)
		if scn == nil && isBKECSI(objects.DefaultStorageClass.Provisioner) {
			return true
		}
		sc := getStorageClass(objects.StorageClasses, scn)
		if sc != nil && isBKECSI(sc.Provisioner) {
			return true
		}
	}

	if volume.CSI != nil {
		if isBKECSI(volume.CSI.Driver) {
			return true
		}
	}
	return false
}

func getStorageClassName(claim *corev1.PersistentVolumeClaim) *string{
	if claim.Spec.StorageClassName != nil {
		return claim.Spec.StorageClassName
	}
	if scn, found := claim.Annotations["volume.beta.kubernetes.io/storage-class"]; found {
		return &scn
	}
	return nil
}

func isBKECSI(referrer string) bool {
	return referrer == BizflyCloudCSIDriver
}

func getStorageClass(classes *st.StorageClassList, name *string) *st.StorageClass {
	for _, c := range classes.Items {
		if c.Name == *name {
			return &c
		}
	}
	return nil
}

func getPVC(claims *corev1.PersistentVolumeClaimList, name string, namespace string) *corev1.PersistentVolumeClaim {
	for _, c := range claims.Items {
		if c.Name == name && c.Namespace == namespace {
			return &c
		}
	}
	return nil
}

func ownedByStatefulSet(references []metav1.OwnerReference) bool {
	for _, ref := range references {
		if ref.Kind == "StatefulSet" {
			return true
		}
	}
	return false
}
