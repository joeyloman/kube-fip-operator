/*
Copyright 2024 The Kubernetes Authors.

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

// Code generated by informer-gen. DO NOT EDIT.

package v1

import (
	internalinterfaces "github.com/joeyloman/kube-fip-operator/pkg/generated/informers/externalversions/internalinterfaces"
)

// Interface provides access to all the informers in this group version.
type Interface interface {
	// FloatingIPs returns a FloatingIPInformer.
	FloatingIPs() FloatingIPInformer
	// FloatingIPRanges returns a FloatingIPRangeInformer.
	FloatingIPRanges() FloatingIPRangeInformer
}

type version struct {
	factory          internalinterfaces.SharedInformerFactory
	namespace        string
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
	return &version{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
}

// FloatingIPs returns a FloatingIPInformer.
func (v *version) FloatingIPs() FloatingIPInformer {
	return &floatingIPInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// FloatingIPRanges returns a FloatingIPRangeInformer.
func (v *version) FloatingIPRanges() FloatingIPRangeInformer {
	return &floatingIPRangeInformer{factory: v.factory, tweakListOptions: v.tweakListOptions}
}
