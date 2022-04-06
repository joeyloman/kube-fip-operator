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

// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/joeyloman/kube-fip-operator/pkg/apis/kubefip.k8s.binbash.org/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// FloatingIPRangeLister helps list FloatingIPRanges.
// All objects returned here must be treated as read-only.
type FloatingIPRangeLister interface {
	// List lists all FloatingIPRanges in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.FloatingIPRange, err error)
	// Get retrieves the FloatingIPRange from the index for a given name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1.FloatingIPRange, error)
	FloatingIPRangeListerExpansion
}

// floatingIPRangeLister implements the FloatingIPRangeLister interface.
type floatingIPRangeLister struct {
	indexer cache.Indexer
}

// NewFloatingIPRangeLister returns a new FloatingIPRangeLister.
func NewFloatingIPRangeLister(indexer cache.Indexer) FloatingIPRangeLister {
	return &floatingIPRangeLister{indexer: indexer}
}

// List lists all FloatingIPRanges in the indexer.
func (s *floatingIPRangeLister) List(selector labels.Selector) (ret []*v1.FloatingIPRange, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.FloatingIPRange))
	})
	return ret, err
}

// Get retrieves the FloatingIPRange from the index for a given name.
func (s *floatingIPRangeLister) Get(name string) (*v1.FloatingIPRange, error) {
	obj, exists, err := s.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("floatingiprange"), name)
	}
	return obj.(*v1.FloatingIPRange), nil
}