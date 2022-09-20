package resource

import (
	"github.com/aiyengar2/hull/pkg/checker"
	corev1 "k8s.io/api/core/v1"
)

func init() {
	if err := corev1.AddToScheme(checker.Scheme); err != nil {
		panic(err)
	}
}

type PersistentVolumes []*corev1.PersistentVolume
type PersistentVolumeClaims []*corev1.PersistentVolumeClaim
