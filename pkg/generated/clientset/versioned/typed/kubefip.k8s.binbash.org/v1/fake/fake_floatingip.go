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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	context "context"
	json "encoding/json"
	fmt "fmt"

	v1 "github.com/joeyloman/kube-fip-operator/pkg/apis/kubefip.k8s.binbash.org/v1"
	kubefipk8sbinbashorgv1 "github.com/joeyloman/kube-fip-operator/pkg/generated/applyconfiguration/kubefip.k8s.binbash.org/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeFloatingIPs implements FloatingIPInterface
type FakeFloatingIPs struct {
	Fake *FakeKubefipV1
	ns   string
}

var floatingipsResource = v1.SchemeGroupVersion.WithResource("floatingips")

var floatingipsKind = v1.SchemeGroupVersion.WithKind("FloatingIP")

// Get takes name of the floatingIP, and returns the corresponding floatingIP object, and an error if there is any.
func (c *FakeFloatingIPs) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.FloatingIP, err error) {
	emptyResult := &v1.FloatingIP{}
	obj, err := c.Fake.
		Invokes(testing.NewGetActionWithOptions(floatingipsResource, c.ns, name, options), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.FloatingIP), err
}

// List takes label and field selectors, and returns the list of FloatingIPs that match those selectors.
func (c *FakeFloatingIPs) List(ctx context.Context, opts metav1.ListOptions) (result *v1.FloatingIPList, err error) {
	emptyResult := &v1.FloatingIPList{}
	obj, err := c.Fake.
		Invokes(testing.NewListActionWithOptions(floatingipsResource, floatingipsKind, c.ns, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.FloatingIPList{ListMeta: obj.(*v1.FloatingIPList).ListMeta}
	for _, item := range obj.(*v1.FloatingIPList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested floatingIPs.
func (c *FakeFloatingIPs) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchActionWithOptions(floatingipsResource, c.ns, opts))

}

// Create takes the representation of a floatingIP and creates it.  Returns the server's representation of the floatingIP, and an error, if there is any.
func (c *FakeFloatingIPs) Create(ctx context.Context, floatingIP *v1.FloatingIP, opts metav1.CreateOptions) (result *v1.FloatingIP, err error) {
	emptyResult := &v1.FloatingIP{}
	obj, err := c.Fake.
		Invokes(testing.NewCreateActionWithOptions(floatingipsResource, c.ns, floatingIP, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.FloatingIP), err
}

// Update takes the representation of a floatingIP and updates it. Returns the server's representation of the floatingIP, and an error, if there is any.
func (c *FakeFloatingIPs) Update(ctx context.Context, floatingIP *v1.FloatingIP, opts metav1.UpdateOptions) (result *v1.FloatingIP, err error) {
	emptyResult := &v1.FloatingIP{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateActionWithOptions(floatingipsResource, c.ns, floatingIP, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.FloatingIP), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeFloatingIPs) UpdateStatus(ctx context.Context, floatingIP *v1.FloatingIP, opts metav1.UpdateOptions) (result *v1.FloatingIP, err error) {
	emptyResult := &v1.FloatingIP{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceActionWithOptions(floatingipsResource, "status", c.ns, floatingIP, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.FloatingIP), err
}

// Delete takes name of the floatingIP and deletes it. Returns an error if one occurs.
func (c *FakeFloatingIPs) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(floatingipsResource, c.ns, name, opts), &v1.FloatingIP{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeFloatingIPs) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	action := testing.NewDeleteCollectionActionWithOptions(floatingipsResource, c.ns, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1.FloatingIPList{})
	return err
}

// Patch applies the patch and returns the patched floatingIP.
func (c *FakeFloatingIPs) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.FloatingIP, err error) {
	emptyResult := &v1.FloatingIP{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithOptions(floatingipsResource, c.ns, name, pt, data, opts, subresources...), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.FloatingIP), err
}

// Apply takes the given apply declarative configuration, applies it and returns the applied floatingIP.
func (c *FakeFloatingIPs) Apply(ctx context.Context, floatingIP *kubefipk8sbinbashorgv1.FloatingIPApplyConfiguration, opts metav1.ApplyOptions) (result *v1.FloatingIP, err error) {
	if floatingIP == nil {
		return nil, fmt.Errorf("floatingIP provided to Apply must not be nil")
	}
	data, err := json.Marshal(floatingIP)
	if err != nil {
		return nil, err
	}
	name := floatingIP.Name
	if name == nil {
		return nil, fmt.Errorf("floatingIP.Name must be provided to Apply")
	}
	emptyResult := &v1.FloatingIP{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithOptions(floatingipsResource, c.ns, *name, types.ApplyPatchType, data, opts.ToPatchOptions()), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.FloatingIP), err
}

// ApplyStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
func (c *FakeFloatingIPs) ApplyStatus(ctx context.Context, floatingIP *kubefipk8sbinbashorgv1.FloatingIPApplyConfiguration, opts metav1.ApplyOptions) (result *v1.FloatingIP, err error) {
	if floatingIP == nil {
		return nil, fmt.Errorf("floatingIP provided to Apply must not be nil")
	}
	data, err := json.Marshal(floatingIP)
	if err != nil {
		return nil, err
	}
	name := floatingIP.Name
	if name == nil {
		return nil, fmt.Errorf("floatingIP.Name must be provided to Apply")
	}
	emptyResult := &v1.FloatingIP{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithOptions(floatingipsResource, c.ns, *name, types.ApplyPatchType, data, opts.ToPatchOptions(), "status"), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.FloatingIP), err
}
