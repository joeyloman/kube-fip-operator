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
	context "context"
	time "time"

	apiskubefipk8sbinbashorgv1 "github.com/joeyloman/kube-fip-operator/pkg/apis/kubefip.k8s.binbash.org/v1"
	versioned "github.com/joeyloman/kube-fip-operator/pkg/generated/clientset/versioned"
	internalinterfaces "github.com/joeyloman/kube-fip-operator/pkg/generated/informers/externalversions/internalinterfaces"
	kubefipk8sbinbashorgv1 "github.com/joeyloman/kube-fip-operator/pkg/generated/listers/kubefip.k8s.binbash.org/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// FloatingIPInformer provides access to a shared informer and lister for
// FloatingIPs.
type FloatingIPInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() kubefipk8sbinbashorgv1.FloatingIPLister
}

type floatingIPInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewFloatingIPInformer constructs a new informer for FloatingIP type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFloatingIPInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredFloatingIPInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredFloatingIPInformer constructs a new informer for FloatingIP type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredFloatingIPInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.KubefipV1().FloatingIPs(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.KubefipV1().FloatingIPs(namespace).Watch(context.TODO(), options)
			},
		},
		&apiskubefipk8sbinbashorgv1.FloatingIP{},
		resyncPeriod,
		indexers,
	)
}

func (f *floatingIPInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredFloatingIPInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *floatingIPInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&apiskubefipk8sbinbashorgv1.FloatingIP{}, f.defaultInformer)
}

func (f *floatingIPInformer) Lister() kubefipk8sbinbashorgv1.FloatingIPLister {
	return kubefipk8sbinbashorgv1.NewFloatingIPLister(f.Informer().GetIndexer())
}
