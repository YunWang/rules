/*
Copyright 2018 The Kubernetes Authors.

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

package v1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
	v1 "k8s.io/rules/pkg/apis/crd/v1"
	scheme "k8s.io/rules/pkg/client/clientset/versioned/scheme"
)

// RulesGetter has a method to return a RuleInterface.
// A group's client should implement this interface.
type RulesGetter interface {
	Rules(namespace string) RuleInterface
}

// RuleInterface has methods to work with Rule resources.
type RuleInterface interface {
	Create(*v1.Rule) (*v1.Rule, error)
	Update(*v1.Rule) (*v1.Rule, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.Rule, error)
	List(opts meta_v1.ListOptions) (*v1.RuleList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Rule, err error)
	RuleExpansion
}

// rules implements RuleInterface
type rules struct {
	client rest.Interface
	ns     string
}

// newRules returns a Rules
func newRules(c *CrdV1Client, namespace string) *rules {
	return &rules{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the rule, and returns the corresponding rule object, and an error if there is any.
func (c *rules) Get(name string, options meta_v1.GetOptions) (result *v1.Rule, err error) {
	result = &v1.Rule{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("rules").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Rules that match those selectors.
func (c *rules) List(opts meta_v1.ListOptions) (result *v1.RuleList, err error) {
	result = &v1.RuleList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("rules").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested rules.
func (c *rules) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("rules").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a rule and creates it.  Returns the server's representation of the rule, and an error, if there is any.
func (c *rules) Create(rule *v1.Rule) (result *v1.Rule, err error) {
	result = &v1.Rule{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("rules").
		Body(rule).
		Do().
		Into(result)
	return
}

// Update takes the representation of a rule and updates it. Returns the server's representation of the rule, and an error, if there is any.
func (c *rules) Update(rule *v1.Rule) (result *v1.Rule, err error) {
	result = &v1.Rule{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("rules").
		Name(rule.Name).
		Body(rule).
		Do().
		Into(result)
	return
}

// Delete takes name of the rule and deletes it. Returns an error if one occurs.
func (c *rules) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("rules").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *rules) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("rules").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched rule.
func (c *rules) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Rule, err error) {
	result = &v1.Rule{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("rules").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
