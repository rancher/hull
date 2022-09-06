package resource

import (
	"github.com/aiyengar2/hull/pkg/test"
	corev1 "k8s.io/api/core/v1"
)

func init() {
	corev1.AddToScheme(test.Scheme)
}

type Storage struct {
	PersistentVolumes
	PersistentVolumeClaims
}

type PersistentVolumes []*corev1.PersistentVolume
type PersistentVolumeClaims []*corev1.PersistentVolumeClaim
