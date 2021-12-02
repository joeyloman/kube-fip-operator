/*
Copyright The Kubernetes Authors.

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
	"context"
	time "time"

	kubefipk8sbinbashorgv1 "github.com/joeyloman/kube-fip-operator/pkg/apis/kubefip.k8s.binbash.org/v1"
	versioned "github.com/joeyloman/kube-fip-operator/pkg/generated/clientset/versioned"
	internalinterfaces "github.com/joeyloman/kube-fip-operator/pkg/generated/informers/externalversions/internalinterfaces"
	v1 "github.com/joeyloman/kube-fip-operator/pkg/generated/listers/kubefip.k8s.binbash.org/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// FloatingIPRangeInformer provides access to a shared informer and lister for
// FloatingIPRanges.
type FloatingIPRangeInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.FloatingIPRangeLister
}

type floatingIPRangeInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewFloatingIPRangeInformer constructs a new informer for FloatingIPRange type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFloatingIPRangeInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredFloatingIPRangeInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredFloatingIPRangeInformer constructs a new informer for FloatingIPRange type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredFloatingIPRangeInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.KubefipV1().FloatingIPRanges().List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.KubefipV1().FloatingIPRanges().Watch(context.TODO(), options)
			},
		},
		&kubefipk8sbinbashorgv1.FloatingIPRange{},
		resyncPeriod,
		indexers,
	)
}

func (f *floatingIPRangeInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredFloatingIPRangeInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *floatingIPRangeInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&kubefipk8sbinbashorgv1.FloatingIPRange{}, f.defaultInformer)
}

func (f *floatingIPRangeInformer) Lister() v1.FloatingIPRangeLister {
	return v1.NewFloatingIPRangeLister(f.Informer().GetIndexer())
}
