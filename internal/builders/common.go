package builders

import (
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sTypes "k8s.io/apimachinery/pkg/types"
)

// OwnerReference creates builder for k8s.io/apimachinery/pkg/types/OwnerReference type
func OwnerReference(name, apiVersion, kind string, uid k8sTypes.UID) *ownerReference {
	return &ownerReference{
		val: &k8sMeta.OwnerReference{
			Name:       name,
			APIVersion: apiVersion,
			Kind:       kind,
			UID:        uid,
		},
	}
}

type ownerReference struct {
	val *k8sMeta.OwnerReference
}

func (b *ownerReference) Controller(ctrl bool) *ownerReference {
	b.val.Controller = &ctrl
	return b
}

func (b *ownerReference) Get() *k8sMeta.OwnerReference {
	return b.val
}
